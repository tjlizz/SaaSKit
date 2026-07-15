package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/saaskit-community/saaskit/internal/users"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Server struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Config Config
	Router *gin.Engine
}

func NewServer(db *gorm.DB, redisClient *redis.Client, cfg Config) (*Server, error) {
	if cfg.AutoMigrate {
		if err := db.AutoMigrate(Models()...); err != nil {
			return nil, err
		}
		// Existing installations predate explicit roles. Their current
		// administrators own the instance and therefore become super admins.
		if err := db.Model(&Admin{}).Where("role = ? OR role IS NULL", "").Update("role", "super_admin").Error; err != nil {
			return nil, err
		}
		// Existing installations predate multi-application support. Backfill
		// app_id on all records that were created before the column existed.
		var defaultApp Application
		if err := db.Where("status = ?", "active").Order("created_at ASC").First(&defaultApp).Error; err == nil {
			for _, table := range []string{"product_plans", "billing_orders", "user_subscriptions"} {
				db.Table(table).Where("app_id = ? OR app_id = ''", "").Update("app_id", defaultApp.ID)
			}
		}
	}
	s := &Server{DB: db, Redis: redisClient, Config: cfg}
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(cors.New(cors.Config{AllowOrigins: cfg.FrontendOrigins, AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Authorization", "Content-Type", "X-API-Key", "X-API-Secret", "X-App-Key"}, AllowCredentials: true, MaxAge: 12 * time.Hour}))
	r.GET("/health", s.health)
	api := r.Group("/api")
	api.GET("/auth/initialization", s.initializationStatus)
	api.POST("/auth/register", s.bootstrapAdmin)
	api.POST("/admin-auth/bootstrap", s.bootstrapAdmin)
	api.POST("/admin-auth/login", s.login)
	api.POST("/admin-auth/refresh", s.refresh)
	api.POST("/admin-auth/logout", s.logout)
	// Vben Admin compatibility endpoints. Keep the canonical admin-auth routes
	// so existing API clients continue to work.
	api.POST("/auth/login", s.login)
	api.POST("/auth/refresh", s.refresh)
	api.POST("/auth/logout", s.logout)
	api.POST("/payments/:provider/notify", s.paymentNotify)
	client := api.Group("/client")
	client.Use(s.clientAuth())
	client.POST("/orders", s.createOrder)
	client.GET("/orders/:orderNo", s.queryOrder)
	client.GET("/orders/:orderNo/provider-status", s.queryProviderOrder)
	client.GET("/subscription/check", s.checkSubscription)
	admin := api.Group("/admin")
	admin.Use(s.adminAuth())
	admin.GET("/me", s.adminInfo)
	admin.GET("/applications", s.listApplications)
	admin.POST("/applications", s.createApplication)
	admin.PUT("/applications/:appId", s.updateApplication)
	admin.DELETE("/applications/:appId", s.deleteApplication)
	// Payment credentials belong to the deployment and are shared by all applications.
	admin.GET("/payment-configs", s.listPaymentConfigs)
	admin.PUT("/payment-configs/:provider", s.savePaymentConfig)
	appAdmin := admin.Group("/applications/:appId")
	appAdmin.Use(s.applicationContext())
	appAdmin.GET("/api-clients", s.listAPIClients)
	appAdmin.POST("/api-clients", s.createAPIClient)
	appAdmin.PUT("/api-clients/:id", s.updateAPIClient)
	appAdmin.DELETE("/api-clients/:id", s.deleteAPIClient)
	appAdmin.POST("/api-clients/:id/rotate-secret", s.rotateAPIClientSecret)
	appAdmin.GET("/plans", s.listPlans)
	appAdmin.POST("/plans", s.createPlan)
	appAdmin.PUT("/plans/:planId", s.updatePlan)
	appAdmin.DELETE("/plans/:planId", s.deletePlan)
	appAdmin.GET("/orders", s.listOrders)
	appAdmin.PUT("/orders/:orderId/status", s.updateOrderStatus)
	appAdmin.GET("/subscriptions", s.listSubscriptions)
	vben := api.Group("")
	vben.Use(s.adminAuth())
	vben.GET("/auth/codes", s.vbenAccessCodes)
	vben.GET("/user/info", s.vbenUserInfo)
	vben.GET("/menu/all", s.vbenMenus)
	userHandler := &users.Handler{
		DB: db, JWTSecret: cfg.JWTSecret, AllowRegistration: cfg.AllowUserRegistration,
		ApplicationActive: func(appID string) bool {
			var count int64
			db.Model(&Application{}).Where("id = ? AND status = ?", appID, "active").Count(&count)
			return count > 0
		},
	}
	userHandler.Register(api, appAdmin, s.publicApplicationContext())
	// Public endpoints for external website integration (requires X-App-Key header)
	publicAPI := api.Group("/public")
	publicAPI.Use(s.publicApplicationContext())
	publicAPI.GET("/plans", s.listPublicPlans)
	accountBilling := api.Group("/account")
	accountBilling.Use(userHandler.UserAuth())
	accountBilling.GET("/orders", s.listAccountOrders)
	accountBilling.GET("/subscription", s.accountSubscription)
	s.Router = r
	return s, nil
}

func (s *Server) health(c *gin.Context) {
	sqlDB, err := s.DB.DB()
	if err != nil || sqlDB.PingContext(c) != nil {
		fail(c, 503, "database unavailable")
		return
	}
	redisStatus := "disabled"
	if s.Redis != nil {
		redisStatus = "ok"
		if s.Redis.Ping(context.Background()).Err() != nil {
			redisStatus = "unavailable"
		}
	}
	ok(c, gin.H{"status": "ok", "redis": redisStatus, "mode": "multi-application"})
}
func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data, "message": "ok"})
}
func created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": data, "message": "ok"})
}
func fail(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"code": status, "message": message})
}

func (s *Server) adminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			fail(c, 401, "missing bearer token")
			return
		}
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(strings.TrimPrefix(header, "Bearer "), claims, func(*jwt.Token) (any, error) { return []byte(s.Config.JWTSecret), nil }, jwt.WithValidMethods([]string{"HS256"}))
		id, _ := claims["sub"].(string)
		kind, _ := claims["kind"].(string)
		if err != nil || !token.Valid || id == "" || kind != "admin" {
			fail(c, 401, "invalid or expired admin token")
			return
		}
		var count int64
		s.DB.Model(&Admin{}).Where("id = ? AND status = ?", id, "active").Count(&count)
		if count == 0 {
			fail(c, 403, "admin is disabled or missing")
			return
		}
		c.Set("admin_id", id)
		c.Next()
	}
}
func adminID(c *gin.Context) string {
	value, _ := c.Get("admin_id")
	id, _ := value.(string)
	return id
}

func (s *Server) clientAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key, secret := c.GetHeader("X-API-Key"), c.GetHeader("X-API-Secret")
		var item APIClient
		if key == "" || secret == "" || s.DB.Where("client_key = ? AND enabled = ?", key, true).First(&item).Error != nil || bcrypt.CompareHashAndPassword([]byte(item.SecretHash), []byte(secret)) != nil {
			fail(c, 401, "invalid API client credentials")
			return
		}
		var activeApplications int64
		s.DB.Model(&Application{}).Where("id = ? AND status = ?", item.AppID, "active").Count(&activeApplications)
		if activeApplications == 0 {
			fail(c, 403, "application is disabled or missing")
			return
		}
		if s.Redis != nil && item.RateLimitPerMin > 0 {
			rateKey := fmt.Sprintf("saaskit:rate:%s:%d", item.ID, time.Now().Unix()/60)
			count, err := s.Redis.Incr(c, rateKey).Result()
			if err == nil {
				_ = s.Redis.Expire(c, rateKey, 2*time.Minute).Err()
				if count > int64(item.RateLimitPerMin) {
					fail(c, 429, "API client rate limit exceeded")
					return
				}
			}
		}
		c.Set("api_client", item)
		c.Next()
	}
}
