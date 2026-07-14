package app

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type adminCredentials struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

const superAdminRole = "super_admin"

var errInstanceInitialized = errors.New("instance has already been initialized")

func (s *Server) initializationStatus(c *gin.Context) {
	var count int64
	if err := s.DB.Model(&Admin{}).Count(&count).Error; err != nil {
		fail(c, 500, "failed to read instance initialization status")
		return
	}
	ok(c, gin.H{"initialized": count > 0, "registration_required": count == 0})
}

func (s *Server) bootstrapAdmin(c *gin.Context) {
	var count int64
	s.DB.Model(&Admin{}).Count(&count)
	if count > 0 {
		fail(c, 409, "instance has already been initialized")
		return
	}
	var input adminCredentials
	if c.ShouldBindJSON(&input) != nil || !strings.Contains(input.Email, "@") || len(input.Password) < 8 {
		fail(c, 400, "valid email and password of at least 8 characters are required")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	item := Admin{ID: uuid.NewString(), Email: strings.ToLower(strings.TrimSpace(input.Email)), Name: strings.TrimSpace(input.Name), Role: superAdminRole, PasswordHash: string(hash), Status: "active"}
	if item.Name == "" {
		item.Name = item.Email
	}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// A transaction-level advisory lock also protects deployments running
		// multiple API replicas against two simultaneous first registrations.
		if tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(7_246_454_831)).Error; err != nil {
				return err
			}
		}
		var current int64
		if err := tx.Model(&Admin{}).Count(&current).Error; err != nil {
			return err
		}
		if current > 0 {
			return errInstanceInitialized
		}
		return tx.Create(&item).Error
	})
	if errors.Is(err, errInstanceInitialized) {
		fail(c, 409, "instance has already been initialized")
		return
	}
	if err != nil {
		fail(c, 409, "administrator initialization conflict")
		return
	}
	s.issueTokens(c, item.ID)
	accessToken := s.accessToken(item.ID)
	created(c, gin.H{"accessToken": accessToken, "access_token": accessToken, "admin": item})
}

func (s *Server) login(c *gin.Context) {
	var input adminCredentials
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "invalid request body")
		return
	}
	identity := input.Email
	if strings.TrimSpace(identity) == "" {
		identity = input.Username
	}
	identity = strings.ToLower(strings.TrimSpace(identity))
	var adminCount int64
	s.DB.Model(&Admin{}).Count(&adminCount)
	if adminCount == 0 {
		fail(c, 428, "instance is not initialized; register the first administrator")
		return
	}
	var item Admin
	if err := s.DB.Where("LOWER(email) = ? OR LOWER(name) = ?", identity, identity).First(&item).Error; err != nil || bcrypt.CompareHashAndPassword([]byte(item.PasswordHash), []byte(input.Password)) != nil {
		fail(c, 401, "invalid username or password")
		return
	}
	if item.Status != "active" {
		fail(c, 403, "admin is disabled")
		return
	}
	now := time.Now()
	s.DB.Model(&item).Update("last_login_at", &now)
	item.LastLoginAt = &now
	s.issueTokens(c, item.ID)
	accessToken := s.accessToken(item.ID)
	ok(c, gin.H{"accessToken": accessToken, "access_token": accessToken, "admin": item})
}

func (s *Server) refresh(c *gin.Context) {
	raw, err := c.Cookie("saaskit_admin_refresh")
	if err != nil {
		fail(c, 401, "missing refresh token")
		return
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return []byte(s.Config.JWTSecret), nil }, jwt.WithValidMethods([]string{"HS256"}))
	id, _ := claims["sub"].(string)
	kind, _ := claims["kind"].(string)
	if err != nil || !token.Valid || id == "" || kind != "admin_refresh" {
		fail(c, 401, "invalid refresh token")
		return
	}
	ok(c, s.accessToken(id))
}
func (s *Server) logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("saaskit_admin_refresh", "", -1, "/api", "", strings.HasPrefix(s.Config.PublicURL, "https://"), true)
	ok(c, nil)
}
func (s *Server) accessToken(id string) string {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": id, "kind": "admin", "exp": time.Now().Add(2 * time.Hour).Unix(), "iat": time.Now().Unix()}).SignedString([]byte(s.Config.JWTSecret))
	return token
}
func (s *Server) issueTokens(c *gin.Context, id string) {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": id, "kind": "admin_refresh", "exp": time.Now().Add(30 * 24 * time.Hour).Unix(), "iat": time.Now().Unix()}).SignedString([]byte(s.Config.JWTSecret))
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("saaskit_admin_refresh", token, 30*24*3600, "/api", "", strings.HasPrefix(s.Config.PublicURL, "https://"), true)
}
func (s *Server) adminInfo(c *gin.Context) {
	var item Admin
	if s.DB.First(&item, "id = ?", adminID(c)).Error != nil {
		fail(c, 404, "admin not found")
		return
	}
	ok(c, item)
}

func (s *Server) vbenUserInfo(c *gin.Context) {
	var item Admin
	if s.DB.First(&item, "id = ?", adminID(c)).Error != nil {
		fail(c, 404, "admin not found")
		return
	}
	role := item.Role
	if role == "" {
		role = superAdminRole
	}
	ok(c, gin.H{"userId": item.ID, "username": item.Email, "realName": item.Name, "avatar": "", "roles": []string{role}, "desc": "SaaSKit administrator", "homePath": "/dashboard/analytics", "token": ""})
}

func (s *Server) vbenAccessCodes(c *gin.Context) {
	var item Admin
	if s.DB.Select("role").First(&item, "id = ?", adminID(c)).Error != nil {
		fail(c, 404, "admin not found")
		return
	}
	if item.Role == superAdminRole || item.Role == "" {
		ok(c, []string{"SAASKIT_SUPER_ADMIN"})
		return
	}
	ok(c, []string{"SAASKIT_ADMIN"})
}
func (s *Server) vbenMenus(c *gin.Context) { ok(c, []any{}) }
