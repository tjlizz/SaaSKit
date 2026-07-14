package app

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPaymentConfigEncryptionRoundTrip(t *testing.T) {
	key := [32]byte{9, 8, 7}
	input := map[string]string{"app_id": "123", "private_key": "secret"}
	encrypted, err := encryptConfig(key, input)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains([]byte(encrypted), []byte("secret")) {
		t.Fatal("ciphertext leaks plaintext")
	}
	output, err := decryptConfig(key, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if output["private_key"] != "secret" {
		t.Fatalf("round trip failed: %#v", output)
	}
}

func TestPeriodEnd(t *testing.T) {
	start := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	if got := periodEnd(start, "monthly"); !got.Equal(time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("monthly: %s", got)
	}
	if got := periodEnd(start, "yearly"); !got.Equal(time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("yearly: %s", got)
	}
}

func TestWechatNotificationVerificationAndDecryption(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privatePEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: mustPKCS8(t, privateKey)}))
	publicDER, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	publicPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}))
	apiKey := "12345678901234567890123456789012"
	nonce := "123456789012"
	aad := "transaction"
	block, _ := aes.NewCipher([]byte(apiKey))
	gcm, _ := cipher.NewGCM(block)
	transaction := []byte(`{"out_trade_no":"order-1","transaction_id":"wx-1","trade_state":"SUCCESS"}`)
	ciphertext := gcm.Seal(nil, []byte(nonce), transaction, []byte(aad))
	body, _ := json.Marshal(map[string]any{"resource": map[string]string{"ciphertext": base64.StdEncoding.EncodeToString(ciphertext), "nonce": nonce, "associated_data": aad}})
	timestamp := "1770000000"
	requestNonce := "notify-nonce"
	signature, err := rsaSign(timestamp+"\n"+requestNonce+"\n"+string(body)+"\n", privatePEM)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Wechatpay-Timestamp", timestamp)
	req.Header.Set("Wechatpay-Nonce", requestNonce)
	req.Header.Set("Wechatpay-Signature", signature)
	result, err := (wechatProvider{Config: map[string]string{"platform_public_key": publicPEM, "api_v3_key": apiKey}}).VerifyNotify(req, body)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Paid || result.OrderNo != "order-1" || result.TradeNo != "wx-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func mustPKCS8(t *testing.T, key *rsa.PrivateKey) []byte {
	t.Helper()
	value, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
