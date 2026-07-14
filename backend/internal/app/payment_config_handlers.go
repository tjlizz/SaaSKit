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

func mergePaymentConfig(stored, submitted map[string]string) map[string]string {
	merged := make(map[string]string, len(stored)+len(submitted))
	for key, value := range stored {
		merged[key] = value
	}
	for key, value := range submitted {
		if value = strings.TrimSpace(value); value != "" {
			merged[key] = value
		}
	}
	return merged
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
	var existing PaymentConfig
	findErr := s.DB.Where("provider = ?", provider).First(&existing).Error
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		fail(c, 500, "could not read payment config")
		return
	}
	stored := map[string]string{}
	if findErr == nil && existing.Ciphertext != "" {
		decrypted, err := decryptConfig(s.Config.PaymentKey, existing.Ciphertext)
		if err != nil {
			fail(c, 500, "payment config cannot be decrypted")
			return
		}
		stored = decrypted
	}
	// Secret fields are never returned to the browser. Empty submitted values
	// therefore mean "keep the stored value" rather than "erase the secret".
	merged := mergePaymentConfig(stored, input.Config)
	if input.Enabled {
		required := []string{"app_id", "private_key"}
		if provider == "alipay" {
			required = append(required, "public_key")
		} else {
			required = append(required, "mch_id", "serial_no", "api_v3_key", "platform_public_key")
		}
		for _, key := range required {
			if merged[key] == "" {
				fail(c, 400, key+" is required when enabled")
				return
			}
		}
	}
	merged["sandbox"] = "false"
	if input.Sandbox {
		merged["sandbox"] = "true"
	}
	ciphertext, err := encryptConfig(s.Config.PaymentKey, merged)
	if err != nil {
		fail(c, 500, "could not encrypt payment config")
		return
	}
	id := existing.ID
	if id == "" {
		id = uuid.NewString()
	}
	item := PaymentConfig{ID: id, Provider: provider, Enabled: input.Enabled, Sandbox: input.Sandbox, Ciphertext: ciphertext}
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
