package app

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type PaymentRequest struct {
	OrderNo, Subject, Channel, NotifyURL, ReturnURL string
	AmountCents                                     int64
}
type PaymentResult struct {
	Provider   string `json:"provider"`
	Type       string `json:"type"`
	PaymentURL string `json:"payment_url,omitempty"`
	QRCode     string `json:"qr_code,omitempty"`
}
type NotifyResult struct {
	OrderNo, TradeNo string
	Paid             bool
}
type PaymentProvider interface {
	CreatePayment(context.Context, PaymentRequest) (PaymentResult, error)
	QueryPayment(context.Context, string) (string, error)
	VerifyNotify(*http.Request, []byte) (NotifyResult, error)
}

type mockProvider struct{ name string }

func (p mockProvider) CreatePayment(_ context.Context, r PaymentRequest) (PaymentResult, error) {
	u := r.NotifyURL + "?mock=1&out_trade_no=" + url.QueryEscape(r.OrderNo)
	return PaymentResult{Provider: p.name, Type: "mock", PaymentURL: u, QRCode: u}, nil
}
func (p mockProvider) QueryPayment(context.Context, string) (string, error) { return "pending", nil }
func (p mockProvider) VerifyNotify(r *http.Request, body []byte) (NotifyResult, error) {
	order := r.URL.Query().Get("out_trade_no")
	if order == "" {
		var v map[string]any
		_ = json.Unmarshal(body, &v)
		order, _ = v["out_trade_no"].(string)
	}
	if order == "" {
		return NotifyResult{}, errors.New("missing out_trade_no")
	}
	return NotifyResult{OrderNo: order, TradeNo: "mock_" + order, Paid: true}, nil
}

type alipayProvider struct{ Config map[string]string }

func (p alipayProvider) gateway() string {
	if v := p.Config["gateway"]; v != "" {
		return v
	}
	if p.Config["sandbox"] == "true" {
		return "https://openapi-sandbox.dl.alipaydev.com/gateway.do"
	}
	return "https://openapi.alipay.com/gateway.do"
}
func (p alipayProvider) CreatePayment(ctx context.Context, r PaymentRequest) (PaymentResult, error) {
	method := "alipay.trade.page.pay"
	product := "FAST_INSTANT_TRADE_PAY"
	resultType := "redirect"
	if r.Channel == "qr" {
		method = "alipay.trade.precreate"
		product = "FACE_TO_FACE_PAYMENT"
		resultType = "qr"
	}
	biz, _ := json.Marshal(map[string]any{"out_trade_no": r.OrderNo, "total_amount": fmt.Sprintf("%.2f", float64(r.AmountCents)/100), "subject": r.Subject, "product_code": product})
	params := map[string]string{"app_id": p.Config["app_id"], "method": method, "format": "JSON", "charset": "utf-8", "sign_type": "RSA2", "timestamp": time.Now().Format("2006-01-02 15:04:05"), "version": "1.0", "notify_url": r.NotifyURL, "biz_content": string(biz)}
	if r.ReturnURL != "" {
		params["return_url"] = r.ReturnURL
	}
	sig, err := rsaSign(canonical(params), p.Config["private_key"])
	if err != nil {
		return PaymentResult{}, err
	}
	params["sign"] = sig
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	paymentURL := p.gateway() + "?" + values.Encode()
	if r.Channel == "qr" {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.gateway(), strings.NewReader(values.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return PaymentResult{}, err
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		var envelope struct {
			Response struct {
				Code    string `json:"code"`
				Message string `json:"msg"`
				QRCode  string `json:"qr_code"`
			} `json:"alipay_trade_precreate_response"`
		}
		if json.Unmarshal(data, &envelope) != nil || envelope.Response.Code != "10000" || envelope.Response.QRCode == "" {
			return PaymentResult{}, fmt.Errorf("alipay precreate failed: %s", string(data))
		}
		return PaymentResult{Provider: "alipay", Type: "qr", QRCode: envelope.Response.QRCode}, nil
	}
	return PaymentResult{Provider: "alipay", Type: resultType, PaymentURL: paymentURL, QRCode: paymentURL}, nil
}
func (p alipayProvider) QueryPayment(ctx context.Context, orderNo string) (string, error) {
	biz, _ := json.Marshal(map[string]string{"out_trade_no": orderNo})
	params := map[string]string{"app_id": p.Config["app_id"], "method": "alipay.trade.query", "format": "JSON", "charset": "utf-8", "sign_type": "RSA2", "timestamp": time.Now().Format("2006-01-02 15:04:05"), "version": "1.0", "biz_content": string(biz)}
	sig, err := rsaSign(canonical(params), p.Config["private_key"])
	if err != nil {
		return "", err
	}
	params["sign"] = sig
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.gateway(), strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var out struct {
		Response struct {
			Code        string `json:"code"`
			TradeStatus string `json:"trade_status"`
		} `json:"alipay_trade_query_response"`
	}
	if json.Unmarshal(data, &out) != nil || out.Response.Code != "10000" {
		return "", fmt.Errorf("alipay query failed: %s", string(data))
	}
	return out.Response.TradeStatus, nil
}
func (p alipayProvider) VerifyNotify(r *http.Request, body []byte) (NotifyResult, error) {
	if err := r.ParseForm(); err != nil {
		return NotifyResult{}, err
	}
	params := map[string]string{}
	for k := range r.PostForm {
		if k != "sign" && k != "sign_type" {
			params[k] = r.PostForm.Get(k)
		}
	}
	if !rsaVerify(canonical(params), r.PostForm.Get("sign"), p.Config["public_key"]) {
		return NotifyResult{}, errors.New("invalid alipay signature")
	}
	status := r.PostForm.Get("trade_status")
	return NotifyResult{OrderNo: r.PostForm.Get("out_trade_no"), TradeNo: r.PostForm.Get("trade_no"), Paid: status == "TRADE_SUCCESS" || status == "TRADE_FINISHED"}, nil
}

type wechatProvider struct{ Config map[string]string }

func (p wechatProvider) CreatePayment(ctx context.Context, r PaymentRequest) (PaymentResult, error) {
	payload, _ := json.Marshal(map[string]any{"appid": p.Config["app_id"], "mchid": p.Config["mch_id"], "description": r.Subject, "out_trade_no": r.OrderNo, "notify_url": r.NotifyURL, "amount": map[string]any{"total": r.AmountCents, "currency": "CNY"}})
	endpoint := "https://api.mch.weixin.qq.com/v3/pay/transactions/native"
	nonce := randomSecret(16)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := "POST\n/v3/pay/transactions/native\n" + timestamp + "\n" + nonce + "\n" + string(payload) + "\n"
	signature, err := rsaSign(message, p.Config["private_key"])
	if err != nil {
		return PaymentResult{}, err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`, p.Config["mch_id"], nonce, timestamp, p.Config["serial_no"], signature))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return PaymentResult{}, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode/100 != 2 {
		return PaymentResult{}, fmt.Errorf("wechat api %d: %s", resp.StatusCode, string(data))
	}
	var out struct {
		CodeURL string `json:"code_url"`
	}
	if json.Unmarshal(data, &out) != nil || out.CodeURL == "" {
		return PaymentResult{}, errors.New("invalid wechat response")
	}
	return PaymentResult{Provider: "wechat", Type: "qr", QRCode: out.CodeURL}, nil
}
func (p wechatProvider) QueryPayment(ctx context.Context, orderNo string) (string, error) {
	path := "/v3/pay/transactions/out-trade-no/" + url.PathEscape(orderNo) + "?mchid=" + url.QueryEscape(p.Config["mch_id"])
	nonce := randomSecret(16)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := "GET\n" + path + "\n" + timestamp + "\n" + nonce + "\n\n"
	signature, err := rsaSign(message, p.Config["private_key"])
	if err != nil {
		return "", err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.mch.weixin.qq.com"+path, nil)
	req.Header.Set("Authorization", fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`, p.Config["mch_id"], nonce, timestamp, p.Config["serial_no"], signature))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("wechat query failed: %s", string(data))
	}
	var out struct {
		TradeState string `json:"trade_state"`
	}
	if json.Unmarshal(data, &out) != nil {
		return "", errors.New("invalid wechat query response")
	}
	return out.TradeState, nil
}
func (p wechatProvider) VerifyNotify(r *http.Request, body []byte) (NotifyResult, error) {
	timestamp, nonce, signature := r.Header.Get("Wechatpay-Timestamp"), r.Header.Get("Wechatpay-Nonce"), r.Header.Get("Wechatpay-Signature")
	if timestamp == "" || nonce == "" || signature == "" {
		return NotifyResult{}, errors.New("missing wechat signature headers")
	}
	if !rsaVerify(timestamp+"\n"+nonce+"\n"+string(body)+"\n", signature, p.Config["platform_public_key"]) {
		return NotifyResult{}, errors.New("invalid wechat signature")
	}
	var envelope struct {
		Resource struct {
			Ciphertext     string `json:"ciphertext"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return NotifyResult{}, err
	}
	plain, err := decryptWechat(envelope.Resource.Ciphertext, envelope.Resource.Nonce, envelope.Resource.AssociatedData, p.Config["api_v3_key"])
	if err != nil {
		return NotifyResult{}, err
	}
	var tx struct {
		OutTradeNo, TransactionID, TradeState string `json:"-"`
	}
	var raw map[string]any
	if json.Unmarshal(plain, &raw) != nil {
		return NotifyResult{}, errors.New("invalid wechat resource")
	}
	tx.OutTradeNo, _ = raw["out_trade_no"].(string)
	tx.TransactionID, _ = raw["transaction_id"].(string)
	tx.TradeState, _ = raw["trade_state"].(string)
	return NotifyResult{OrderNo: tx.OutTradeNo, TradeNo: tx.TransactionID, Paid: tx.TradeState == "SUCCESS"}, nil
}

func canonical(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for k, v := range values {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+values[k])
	}
	return strings.Join(parts, "&")
}
func parsePrivate(raw string) (*rsa.PrivateKey, error) {
	raw = strings.ReplaceAll(raw, "\\n", "\n")
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, errors.New("invalid RSA private key")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
func parsePublic(raw string) (*rsa.PublicKey, error) {
	raw = strings.ReplaceAll(raw, "\\n", "\n")
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, errors.New("invalid RSA public key")
	}
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err == nil {
		if key, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			return key, nil
		}
	}
	return nil, errors.New("invalid RSA public key")
}
func rsaSign(message, keyPEM string) (string, error) {
	key, err := parsePrivate(keyPEM)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	return base64.StdEncoding.EncodeToString(sig), err
}
func rsaVerify(message, signature, keyPEM string) bool {
	key, err := parsePublic(keyPEM)
	if err != nil {
		return false
	}
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	digest := sha256.Sum256([]byte(message))
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], sig) == nil
}
func decryptWechat(encoded, nonce, aad, key string) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("api_v3_key must be 32 bytes")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, []byte(nonce), ciphertext, []byte(aad))
}

func encryptConfig(key [32]byte, values map[string]string) (string, error) {
	plain, _ := json.Marshal(values)
	block, _ := aes.NewCipher(key[:])
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(append(nonce, gcm.Seal(nil, nonce, plain, nil)...)), nil
}
func decryptConfig(key [32]byte, value string) (map[string]string, error) {
	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	block, _ := aes.NewCipher(key[:])
	gcm, _ := cipher.NewGCM(block)
	if len(raw) < gcm.NonceSize() {
		return nil, errors.New("invalid encrypted config")
	}
	plain, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	return out, json.Unmarshal(plain, &out)
}
