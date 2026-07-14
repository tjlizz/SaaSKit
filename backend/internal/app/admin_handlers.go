package app

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func randomSecret(size int) string {
	value := make([]byte, size)
	_, _ = rand.Read(value)
	return hex.EncodeToString(value)
}

func (s *Server) listAPIClients(c *gin.Context) {
	var items []APIClient
	s.DB.Where("app_id = ?", applicationID(c)).Order("created_at desc").Find(&items)
	ok(c, items)
}
func (s *Server) createAPIClient(c *gin.Context) {
	var input struct {
		Name            string `json:"name"`
		Enabled         *bool  `json:"enabled"`
		RateLimitPerMin int    `json:"rate_limit_per_min"`
	}
	if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Name) == "" {
		fail(c, 400, "name is required")
		return
	}
	secret := randomSecret(32)
	hash, _ := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	if input.RateLimitPerMin <= 0 {
		input.RateLimitPerMin = 300
	}
	item := APIClient{ID: uuid.NewString(), AppID: applicationID(c), Name: strings.TrimSpace(input.Name), ClientKey: "client_" + randomSecret(16), SecretHash: string(hash), Enabled: enabled, RateLimitPerMin: input.RateLimitPerMin}
	if err := s.DB.Create(&item).Error; err != nil {
		fail(c, 500, "could not create API client")
		return
	}
	created(c, gin.H{"client": item, "client_secret": secret})
}
func (s *Server) updateAPIClient(c *gin.Context) {
	var item APIClient
	if s.DB.Where("id = ? AND app_id = ?", c.Param("id"), applicationID(c)).First(&item).Error != nil {
		fail(c, 404, "API client not found")
		return
	}
	var input struct {
		Name            *string `json:"name"`
		Enabled         *bool   `json:"enabled"`
		RateLimitPerMin *int    `json:"rate_limit_per_min"`
	}
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "invalid request body")
		return
	}
	updates := map[string]any{}
	if input.Name != nil && strings.TrimSpace(*input.Name) != "" {
		updates["name"] = strings.TrimSpace(*input.Name)
	}
	if input.Enabled != nil {
		updates["enabled"] = *input.Enabled
	}
	if input.RateLimitPerMin != nil && *input.RateLimitPerMin > 0 {
		updates["rate_limit_per_min"] = *input.RateLimitPerMin
	}
	s.DB.Model(&item).Updates(updates)
	s.DB.Where("id = ? AND app_id = ?", item.ID, applicationID(c)).First(&item)
	ok(c, item)
}
func (s *Server) deleteAPIClient(c *gin.Context) {
	result := s.DB.Where("id = ? AND app_id = ?", c.Param("id"), applicationID(c)).Delete(&APIClient{})
	if result.RowsAffected == 0 {
		fail(c, 404, "API client not found")
		return
	}
	ok(c, nil)
}
func (s *Server) rotateAPIClientSecret(c *gin.Context) {
	var item APIClient
	if s.DB.Where("id = ? AND app_id = ?", c.Param("id"), applicationID(c)).First(&item).Error != nil {
		fail(c, 404, "API client not found")
		return
	}
	secret := randomSecret(32)
	hash, _ := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	s.DB.Model(&item).Update("secret_hash", string(hash))
	ok(c, gin.H{"client_secret": secret})
}

var validCycles = map[string]bool{"free": true, "monthly": true, "yearly": true, "lifetime": true, "one_time": true}

func (s *Server) listPlans(c *gin.Context) {
	var items []Plan
	s.DB.Where("app_id = ?", applicationID(c)).Order("sort_order asc, created_at desc").Find(&items)
	ok(c, items)
}
func bindPlan(c *gin.Context, item *Plan) bool {
	var input struct {
		PlanCode           string `json:"plan_code"`
		Name               string `json:"name"`
		Description        string `json:"description"`
		BillingCycle       string `json:"billing_cycle"`
		PriceCents         int64  `json:"price_cents"`
		Currency           string `json:"currency"`
		AutoRenewSupported bool   `json:"auto_renew_supported"`
		DeviceLimit        int    `json:"device_limit"`
		CreditQuota        int64  `json:"credit_quota"`
		SeatLimit          int    `json:"seat_limit"`
		Recommended        bool   `json:"recommended"`
		Enabled            *bool  `json:"enabled"`
		SortOrder          int    `json:"sort_order"`
	}
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "invalid request body")
		return false
	}
	if strings.TrimSpace(input.PlanCode) == "" || strings.TrimSpace(input.Name) == "" || !validCycles[input.BillingCycle] || input.PriceCents < 0 {
		fail(c, 400, "plan_code, name, valid billing_cycle and non-negative price_cents are required")
		return false
	}
	item.PlanCode = strings.TrimSpace(input.PlanCode)
	item.Name = strings.TrimSpace(input.Name)
	item.Description = input.Description
	item.BillingCycle = input.BillingCycle
	item.PriceCents = input.PriceCents
	item.Currency = input.Currency
	if item.Currency == "" {
		item.Currency = "CNY"
	}
	item.AutoRenewSupported = input.AutoRenewSupported
	item.DeviceLimit = input.DeviceLimit
	item.CreditQuota = input.CreditQuota
	item.SeatLimit = input.SeatLimit
	item.Recommended = input.Recommended
	if input.Enabled != nil {
		item.Enabled = *input.Enabled
	}
	item.SortOrder = input.SortOrder
	return true
}
func (s *Server) createPlan(c *gin.Context) {
	item := Plan{ID: uuid.NewString(), AppID: applicationID(c), Enabled: true, DeviceLimit: 1}
	if !bindPlan(c, &item) {
		return
	}
	if err := s.DB.Create(&item).Error; err != nil {
		fail(c, 409, "plan code already exists")
		return
	}
	created(c, item)
}
func (s *Server) updatePlan(c *gin.Context) {
	var item Plan
	if s.DB.Where("id = ? AND app_id = ?", c.Param("planId"), applicationID(c)).First(&item).Error != nil {
		fail(c, 404, "plan not found")
		return
	}
	if !bindPlan(c, &item) {
		return
	}
	if err := s.DB.Save(&item).Error; err != nil {
		fail(c, 409, "plan code already exists")
		return
	}
	ok(c, item)
}
func (s *Server) deletePlan(c *gin.Context) {
	var count int64
	s.DB.Model(&Order{}).Where("app_id = ? AND plan_id = ?", applicationID(c), c.Param("planId")).Count(&count)
	if count > 0 {
		fail(c, 409, "plan with orders cannot be deleted")
		return
	}
	result := s.DB.Where("id = ? AND app_id = ?", c.Param("planId"), applicationID(c)).Delete(&Plan{})
	if result.RowsAffected == 0 {
		fail(c, 404, "plan not found")
		return
	}
	ok(c, nil)
}
func (s *Server) listOrders(c *gin.Context) {
	query := s.DB.Where("app_id = ?", applicationID(c)).Order("created_at desc")
	if userID := strings.TrimSpace(c.Query("user_id")); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	}
	var items []Order
	query.Preload("User").Find(&items)
	ok(c, items)
}
func (s *Server) listSubscriptions(c *gin.Context) {
	query := s.DB.Where("app_id = ?", applicationID(c)).Preload("Plan").Preload("User").Order("created_at desc")
	if userID := strings.TrimSpace(c.Query("user_id")); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	var items []Subscription
	query.Find(&items)
	ok(c, items)
}
