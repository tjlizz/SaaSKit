package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saaskit-community/saaskit/internal/users"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type createOrderInput struct {
	PlanCode  string `json:"plan_code"`
	UserID    string `json:"user_id"`
	Provider  string `json:"provider"`
	Channel   string `json:"channel"`
	ReturnURL string `json:"return_url"`
	Quantity  int    `json:"quantity"`
	RequestID string `json:"request_id"`
}

func (s *Server) createOrder(c *gin.Context) {
	client, _ := c.Get("api_client")
	apiClient, _ := client.(APIClient)
	appID := apiClient.AppID
	var input createOrderInput
	if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.PlanCode) == "" || strings.TrimSpace(input.UserID) == "" {
		fail(c, 400, "plan_code and user_id are required")
		return
	}
	var user users.User
	if s.DB.Where("id = ? AND app_id = ? AND status = ?", input.UserID, appID, "active").First(&user).Error != nil {
		fail(c, 404, "active user not found")
		return
	}
	if input.Quantity <= 0 {
		input.Quantity = 1
	}
	if input.Provider == "" {
		input.Provider = "alipay"
	}
	var plan Plan
	if s.DB.Where("app_id = ? AND plan_code = ? AND enabled = ?", appID, input.PlanCode, true).First(&plan).Error != nil {
		fail(c, 404, "enabled plan not found")
		return
	}
	if plan.BillingCycle == "free" {
		input.Provider = "free"
	}
	if input.Provider != "free" && input.Provider != "alipay" && input.Provider != "wechat" {
		fail(c, 400, "provider must be alipay or wechat")
		return
	}
	orderNo := "SK" + time.Now().UTC().Format("20060102150405") + strings.ToUpper(randomSecret(5))
	if strings.TrimSpace(input.RequestID) != "" {
		var existing Order
		if s.DB.Where("app_id = ? AND order_no = ?", appID, input.RequestID).First(&existing).Error == nil {
			if existing.UserID != input.UserID {
				fail(c, 409, "request_id belongs to another user")
				return
			}
			ok(c, gin.H{"order": existing})
			return
		}
		orderNo = input.RequestID
	}
	item := Order{ID: uuid.NewString(), AppID: appID, OrderNo: orderNo, UserID: user.ID, PlanID: plan.ID, Provider: input.Provider, Status: "pending", AmountCents: plan.PriceCents * int64(input.Quantity), Currency: plan.Currency, Quantity: input.Quantity}
	if err := s.DB.Create(&item).Error; err != nil {
		fail(c, 409, "could not create order; request_id may already exist")
		return
	}
	if input.Provider == "free" {
		if err := s.markPaid(item.OrderNo, "free_"+item.OrderNo); err != nil {
			fail(c, 500, "could not activate free subscription")
			return
		}
		s.DB.First(&item, "id = ?", item.ID)
		created(c, gin.H{"order": item, "payment": PaymentResult{Provider: "free", Type: "none"}})
		return
	}
	provider, err := s.provider(input.Provider)
	if err != nil {
		fail(c, 422, err.Error())
		return
	}
	payment, err := provider.CreatePayment(c, PaymentRequest{OrderNo: item.OrderNo, Subject: plan.Name, Channel: input.Channel, AmountCents: item.AmountCents, NotifyURL: s.Config.PublicURL + "/api/payments/" + input.Provider + "/notify", ReturnURL: input.ReturnURL})
	if err != nil {
		fail(c, 502, "payment provider error: "+err.Error())
		return
	}
	created(c, gin.H{"order": item, "payment": payment})
}

func (s *Server) queryOrder(c *gin.Context) {
	client, _ := c.Get("api_client")
	apiClient, _ := client.(APIClient)
	var item Order
	if s.DB.Where("app_id = ? AND order_no = ?", apiClient.AppID, c.Param("orderNo")).First(&item).Error != nil {
		fail(c, 404, "order not found")
		return
	}
	ok(c, item)
}
func (s *Server) queryProviderOrder(c *gin.Context) {
	client, _ := c.Get("api_client")
	apiClient, _ := client.(APIClient)
	var item Order
	if s.DB.Where("app_id = ? AND order_no = ?", apiClient.AppID, c.Param("orderNo")).First(&item).Error != nil {
		fail(c, 404, "order not found")
		return
	}
	if item.Provider == "free" {
		ok(c, gin.H{"order_no": item.OrderNo, "local_status": item.Status, "provider_status": "FREE"})
		return
	}
	provider, err := s.provider(item.Provider)
	if err != nil {
		fail(c, 422, err.Error())
		return
	}
	status, err := provider.QueryPayment(c, item.OrderNo)
	if err != nil {
		fail(c, 502, "payment provider error: "+err.Error())
		return
	}
	ok(c, gin.H{"order_no": item.OrderNo, "local_status": item.Status, "provider_status": status})
}
func (s *Server) checkSubscription(c *gin.Context) {
	client, _ := c.Get("api_client")
	apiClient, _ := client.(APIClient)
	userID := strings.TrimSpace(c.Query("user_id"))
	if userID == "" {
		fail(c, 400, "user_id is required")
		return
	}
	var item Subscription
	err := s.DB.Preload("Plan").Where("app_id = ? AND user_id = ?", apiClient.AppID, userID).First(&item).Error
	if err != nil && isNotFound(err) {
		ok(c, gin.H{"valid": false, "user_id": userID, "plan": nil, "expires_at": nil})
		return
	}
	if err != nil {
		fail(c, 500, "could not query subscription")
		return
	}
	valid := item.SubscriptionStatus == "active" && item.CurrentPeriodEnd.After(time.Now())
	ok(c, gin.H{"valid": valid, "user_id": userID, "plan": item.Plan, "expires_at": item.CurrentPeriodEnd, "subscription": item})
}

func (s *Server) listAccountOrders(c *gin.Context) {
	var items []Order
	s.DB.Where("app_id = ? AND user_id = ?", users.ApplicationID(c), users.UserID(c)).Order("created_at desc").Find(&items)
	ok(c, items)
}

func (s *Server) accountSubscription(c *gin.Context) {
	var item Subscription
	err := s.DB.Preload("Plan").Where("app_id = ? AND user_id = ?", users.ApplicationID(c), users.UserID(c)).First(&item).Error
	if err != nil && isNotFound(err) {
		ok(c, gin.H{"valid": false, "plan": nil, "expires_at": nil})
		return
	}
	if err != nil {
		fail(c, 500, "could not query subscription")
		return
	}
	ok(c, gin.H{"valid": item.SubscriptionStatus == "active" && item.CurrentPeriodEnd.After(time.Now()), "plan": item.Plan, "expires_at": item.CurrentPeriodEnd, "subscription": item})
}

func (s *Server) paymentNotify(c *gin.Context) {
	providerName := c.Param("provider")
	if providerName != "alipay" && providerName != "wechat" {
		c.String(404, "fail")
		return
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 2<<20))
	if err != nil {
		c.String(400, "fail")
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	orderNo := c.Query("out_trade_no")
	if orderNo == "" && providerName == "alipay" {
		_ = c.Request.ParseForm()
		orderNo = c.Request.PostForm.Get("out_trade_no")
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
	}
	if orderNo == "" {
		var raw map[string]any
		_ = json.Unmarshal(body, &raw)
		orderNo, _ = raw["out_trade_no"].(string)
	}
	provider, err := s.provider(providerName)
	if err != nil {
		c.String(400, "fail")
		return
	}
	result, err := provider.VerifyNotify(c.Request, body)
	if err != nil || !result.Paid {
		c.String(400, "fail")
		return
	}
	if orderNo != "" && result.OrderNo != orderNo {
		c.String(400, "fail")
		return
	}
	orderNo = result.OrderNo
	var order Order
	if s.DB.Where("order_no = ? AND provider = ?", orderNo, providerName).First(&order).Error != nil {
		c.String(404, "fail")
		return
	}
	if err = s.markPaid(order.OrderNo, result.TradeNo); err != nil {
		c.String(500, "fail")
		return
	}
	if providerName == "wechat" {
		c.JSON(200, gin.H{"code": "SUCCESS", "message": "成功"})
	} else {
		c.String(200, "success")
	}
}

func (s *Server) markPaid(orderNo, tradeNo string) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var order Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		if order.Status == "paid" {
			return nil
		}
		if order.Status != "pending" {
			return fmt.Errorf("order status is %s", order.Status)
		}
		now := time.Now()
		if err := tx.Model(&order).Updates(map[string]any{"status": "paid", "provider_trade_no": tradeNo, "paid_at": &now}).Error; err != nil {
			return err
		}
		var plan Plan
		if err := tx.First(&plan, "id = ?", order.PlanID).Error; err != nil {
			return err
		}
		var subscription Subscription
		err := tx.Where("app_id = ? AND user_id = ?", order.AppID, order.UserID).First(&subscription).Error
		start := now
		if err == nil && subscription.SubscriptionStatus == "active" && subscription.CurrentPeriodEnd.After(now) {
			start = subscription.CurrentPeriodEnd
		} else if err != nil && !isNotFound(err) {
			return err
		}
		end := periodEnd(start, plan.BillingCycle)
		if isNotFound(err) {
			subscription = Subscription{ID: uuid.NewString(), AppID: order.AppID, UserID: order.UserID}
		}
		subscription.PlanID = plan.ID
		subscription.OrderID = order.ID
		subscription.SubscriptionStatus = "active"
		subscription.CurrentPeriodStart = now
		subscription.CurrentPeriodEnd = end
		subscription.AutoRenew = false
		subscription.RemainingCredits = plan.CreditQuota
		return tx.Save(&subscription).Error
	})
}
func periodEnd(start time.Time, cycle string) time.Time {
	switch cycle {
	case "monthly":
		return start.AddDate(0, 1, 0)
	case "yearly":
		return start.AddDate(1, 0, 0)
	default:
		return start.AddDate(100, 0, 0)
	}
}
func isNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }
func (s *Server) updateOrderStatus(c *gin.Context) {
	var input struct {
		Status string `json:"status"`
	}
	if c.ShouldBindJSON(&input) != nil || (input.Status != "cancelled" && input.Status != "refunded") {
		fail(c, 400, "status must be cancelled or refunded")
		return
	}
	var item Order
	if s.DB.Where("id = ? AND app_id = ?", c.Param("orderId"), applicationID(c)).First(&item).Error != nil {
		fail(c, 404, "order not found")
		return
	}
	if input.Status == "cancelled" && item.Status != "pending" {
		fail(c, 409, "only pending orders can be cancelled")
		return
	}
	if input.Status == "refunded" && item.Status != "paid" {
		fail(c, 409, "only paid orders can be marked refunded")
		return
	}
	s.DB.Model(&item).Update("status", input.Status)
	if input.Status == "refunded" {
		s.DB.Model(&Subscription{}).Where("app_id = ? AND order_id = ?", applicationID(c), item.ID).Updates(map[string]any{"subscription_status": "cancelled", "current_period_end": time.Now()})
	}
	s.DB.First(&item, "id = ?", item.ID)
	ok(c, item)
}
