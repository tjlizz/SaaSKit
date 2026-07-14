package app

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Server) listPaymentConfigs(c *gin.Context) {
	var items []PaymentConfig
	s.DB.Find(&items)
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{"provider": item.Provider, "enabled": item.Enabled, "sandbox": item.Sandbox, "configured": item.Ciphertext != "", "updated_at": item.UpdatedAt})
	}
	ok(c, result)
}
func (s *Server) savePaymentConfig(c *gin.Context) {
	provider := c.Param("provider")
	if provider != "alipay" && provider != "wechat" {
		fail(c, 400, "unsupported provider")
		return
	}
	var input struct {
		Enabled bool              `json:"enabled"`
		Sandbox bool              `json:"sandbox"`
		Config  map[string]string `json:"config"`
	}
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "invalid request body")
		return
	}
	for key, value := range input.Config {
		input.Config[key] = strings.TrimSpace(value)
	}
	if input.Enabled {
		required := []string{"app_id", "private_key"}
		if provider == "alipay" {
			required = append(required, "public_key")
		} else {
			required = append(required, "mch_id", "serial_no", "api_v3_key", "platform_public_key")
		}
		for _, key := range required {
			if input.Config[key] == "" {
				fail(c, 400, key+" is required when enabled")
				return
			}
		}
	}
	input.Config["sandbox"] = "false"
	if input.Sandbox {
		input.Config["sandbox"] = "true"
	}
	ciphertext, err := encryptConfig(s.Config.PaymentKey, input.Config)
	if err != nil {
		fail(c, 500, "could not encrypt payment config")
		return
	}
	item := PaymentConfig{ID: uuid.NewString(), Provider: provider, Enabled: input.Enabled, Sandbox: input.Sandbox, Ciphertext: ciphertext}
	err = s.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "provider"}}, DoUpdates: clause.AssignmentColumns([]string{"enabled", "sandbox", "ciphertext", "updated_at"})}).Create(&item).Error
	if err != nil {
		fail(c, 500, "could not save payment config")
		return
	}
	ok(c, gin.H{"provider": provider, "enabled": input.Enabled, "sandbox": input.Sandbox, "configured": true})
}
func (s *Server) provider(name string) (PaymentProvider, error) {
	if s.Config.PaymentMock {
		return mockProvider{name: name}, nil
	}
	var item PaymentConfig
	if err := s.DB.Where("provider = ? AND enabled = ?", name, true).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(name + " payment is not configured")
		}
		return nil, err
	}
	config, err := decryptConfig(s.Config.PaymentKey, item.Ciphertext)
	if err != nil {
		return nil, errors.New("payment config cannot be decrypted")
	}
	switch name {
	case "alipay":
		return alipayProvider{Config: config}, nil
	case "wechat":
		return wechatProvider{Config: config}, nil
	default:
		return nil, errors.New("unsupported provider")
	}
}
