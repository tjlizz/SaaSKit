package users

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB                *gorm.DB
	JWTSecret         string
	AllowRegistration bool
	ApplicationActive func(string) bool
}

func (h *Handler) Register(api, admin *gin.RouterGroup, applicationMiddleware gin.HandlerFunc) {
	auth := api.Group("/user-auth")
	auth.Use(applicationMiddleware)
	auth.POST("/register", h.register)
	auth.POST("/login", h.login)
	account := api.Group("/account")
	account.Use(h.UserAuth())
	account.GET("/profile", h.profile)
	account.PUT("/profile", h.updateProfile)
	account.PUT("/password", h.changePassword)
	users := admin.Group("/users")
	users.GET("", h.list)
	users.POST("", h.create)
	users.GET("/:id", h.get)
	users.PUT("/:id", h.update)
	users.DELETE("/:id", h.remove)
	users.PUT("/:id/status", h.setStatus)
	users.PUT("/:id/password", h.resetPassword)
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
func normalizedEmail(value string) string { return strings.ToLower(strings.TrimSpace(value)) }
func passwordHash(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	return string(hash), err
}
func validStatus(value string) bool { return value == "active" || value == "disabled" }
func phoneValue(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func (h *Handler) register(c *gin.Context) {
	if !h.AllowRegistration {
		fail(c, 403, "user registration is disabled")
		return
	}
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Phone    string `json:"phone"`
	}
	if c.ShouldBindJSON(&input) != nil || !strings.Contains(input.Email, "@") || len(input.Password) < 8 {
		fail(c, 400, "valid email and password of at least 8 characters are required")
		return
	}
	hash, _ := passwordHash(input.Password)
	item := User{ID: uuid.NewString(), AppID: ApplicationID(c), Email: normalizedEmail(input.Email), Name: strings.TrimSpace(input.Name), Phone: phoneValue(input.Phone), PasswordHash: hash, Status: "active"}
	if item.Name == "" {
		item.Name = item.Email
	}
	if err := h.DB.Create(&item).Error; err != nil {
		fail(c, 409, "email or phone already exists")
		return
	}
	created(c, gin.H{"access_token": h.token(item.ID, item.AppID), "user": item})
}

func (h *Handler) login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "invalid request body")
		return
	}
	var item User
	if err := h.DB.Where("app_id = ? AND email = ?", ApplicationID(c), normalizedEmail(input.Email)).First(&item).Error; err != nil || bcrypt.CompareHashAndPassword([]byte(item.PasswordHash), []byte(input.Password)) != nil {
		fail(c, 401, "invalid email or password")
		return
	}
	if item.Status != "active" {
		fail(c, 403, "user is disabled")
		return
	}
	now := time.Now()
	h.DB.Model(&item).Update("last_login_at", &now)
	item.LastLoginAt = &now
	ok(c, gin.H{"access_token": h.token(item.ID, item.AppID), "user": item})
}

func (h *Handler) token(id, appID string) string {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": id, "app_id": appID, "kind": "user", "exp": time.Now().Add(12 * time.Hour).Unix(), "iat": time.Now().Unix()}).SignedString([]byte(h.JWTSecret))
	return token
}
func (h *Handler) UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			fail(c, 401, "missing bearer token")
			return
		}
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(strings.TrimPrefix(header, "Bearer "), claims, func(*jwt.Token) (any, error) { return []byte(h.JWTSecret), nil }, jwt.WithValidMethods([]string{"HS256"}))
		id, _ := claims["sub"].(string)
		appID, _ := claims["app_id"].(string)
		kind, _ := claims["kind"].(string)
		if err != nil || !token.Valid || id == "" || appID == "" || kind != "user" {
			fail(c, 401, "invalid or expired user token")
			return
		}
		var count int64
		h.DB.Model(&User{}).Where("id = ? AND app_id = ? AND status = ?", id, appID, "active").Count(&count)
		if count == 0 || (h.ApplicationActive != nil && !h.ApplicationActive(appID)) {
			fail(c, 403, "user is disabled or missing")
			return
		}
		c.Set("user_id", id)
		c.Set("app_id", appID)
		c.Next()
	}
}
func UserID(c *gin.Context) string { value, _ := c.Get("user_id"); id, _ := value.(string); return id }
func ApplicationID(c *gin.Context) string {
	value, _ := c.Get("app_id")
	id, _ := value.(string)
	return id
}

func (h *Handler) profile(c *gin.Context) {
	var item User
	if h.DB.Where("id = ? AND app_id = ?", UserID(c), ApplicationID(c)).First(&item).Error != nil {
		fail(c, 404, "user not found")
		return
	}
	ok(c, item)
}
func (h *Handler) updateProfile(c *gin.Context) {
	var input struct {
		Name      *string `json:"name"`
		Phone     *string `json:"phone"`
		AvatarURL *string `json:"avatar_url"`
	}
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "invalid request body")
		return
	}
	updates := map[string]any{}
	if input.Name != nil && strings.TrimSpace(*input.Name) != "" {
		updates["name"] = strings.TrimSpace(*input.Name)
	}
	if input.Phone != nil {
		updates["phone"] = phoneValue(*input.Phone)
		updates["phone_verified"] = false
	}
	if input.AvatarURL != nil {
		updates["avatar_url"] = strings.TrimSpace(*input.AvatarURL)
	}
	if err := h.DB.Model(&User{}).Where("id = ? AND app_id = ?", UserID(c), ApplicationID(c)).Updates(updates).Error; err != nil {
		fail(c, 409, "phone already exists")
		return
	}
	h.profile(c)
}
func (h *Handler) changePassword(c *gin.Context) {
	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if c.ShouldBindJSON(&input) != nil || len(input.NewPassword) < 8 {
		fail(c, 400, "new_password must be at least 8 characters")
		return
	}
	var item User
	h.DB.Where("id = ? AND app_id = ?", UserID(c), ApplicationID(c)).First(&item)
	if bcrypt.CompareHashAndPassword([]byte(item.PasswordHash), []byte(input.CurrentPassword)) != nil {
		fail(c, 401, "current password is incorrect")
		return
	}
	hash, _ := passwordHash(input.NewPassword)
	h.DB.Model(&item).Update("password_hash", hash)
	ok(c, nil)
}

func (h *Handler) list(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	query := h.DB.Model(&User{}).Where("app_id = ?", ApplicationID(c))
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		like := "%" + q + "%"
		query = query.Where("LOWER(email) LIKE LOWER(?) OR LOWER(name) LIKE LOWER(?) OR phone LIKE ?", like, like, like)
	}
	if status := c.Query("status"); validStatus(status) {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)
	var items []User
	query.Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&items)
	ok(c, gin.H{"items": items, "total": total, "page": page, "page_size": size})
}
func (h *Handler) create(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Status   string `json:"status"`
	}
	if c.ShouldBindJSON(&input) != nil || !strings.Contains(input.Email, "@") || len(input.Password) < 8 {
		fail(c, 400, "valid email and password of at least 8 characters are required")
		return
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if !validStatus(input.Status) {
		fail(c, 400, "status must be active or disabled")
		return
	}
	hash, _ := passwordHash(input.Password)
	item := User{ID: uuid.NewString(), AppID: ApplicationID(c), Email: normalizedEmail(input.Email), Name: strings.TrimSpace(input.Name), Phone: phoneValue(input.Phone), PasswordHash: hash, Status: input.Status}
	if item.Name == "" {
		item.Name = item.Email
	}
	if err := h.DB.Create(&item).Error; err != nil {
		fail(c, 409, "email or phone already exists")
		return
	}
	created(c, item)
}
func (h *Handler) get(c *gin.Context) {
	var item User
	if h.DB.Where("id = ? AND app_id = ?", c.Param("id"), ApplicationID(c)).First(&item).Error != nil {
		fail(c, 404, "user not found")
		return
	}
	ok(c, item)
}
func (h *Handler) update(c *gin.Context) {
	var input struct {
		Email         *string `json:"email"`
		Name          *string `json:"name"`
		Phone         *string `json:"phone"`
		AvatarURL     *string `json:"avatar_url"`
		EmailVerified *bool   `json:"email_verified"`
		PhoneVerified *bool   `json:"phone_verified"`
	}
	if c.ShouldBindJSON(&input) != nil {
		fail(c, 400, "invalid request body")
		return
	}
	updates := map[string]any{}
	if input.Email != nil && strings.Contains(*input.Email, "@") {
		updates["email"] = normalizedEmail(*input.Email)
	}
	if input.Name != nil && strings.TrimSpace(*input.Name) != "" {
		updates["name"] = strings.TrimSpace(*input.Name)
	}
	if input.Phone != nil {
		updates["phone"] = phoneValue(*input.Phone)
	}
	if input.AvatarURL != nil {
		updates["avatar_url"] = strings.TrimSpace(*input.AvatarURL)
	}
	if input.EmailVerified != nil {
		updates["email_verified"] = *input.EmailVerified
	}
	if input.PhoneVerified != nil {
		updates["phone_verified"] = *input.PhoneVerified
	}
	result := h.DB.Model(&User{}).Where("id = ? AND app_id = ?", c.Param("id"), ApplicationID(c)).Updates(updates)
	if result.Error != nil {
		fail(c, 409, "email or phone already exists")
		return
	}
	if result.RowsAffected == 0 {
		fail(c, 404, "user not found")
		return
	}
	h.get(c)
}
func (h *Handler) setStatus(c *gin.Context) {
	var input struct {
		Status string `json:"status"`
	}
	if c.ShouldBindJSON(&input) != nil || !validStatus(input.Status) {
		fail(c, 400, "status must be active or disabled")
		return
	}
	result := h.DB.Model(&User{}).Where("id = ? AND app_id = ?", c.Param("id"), ApplicationID(c)).Update("status", input.Status)
	if result.RowsAffected == 0 {
		fail(c, 404, "user not found")
		return
	}
	h.get(c)
}
func (h *Handler) resetPassword(c *gin.Context) {
	var input struct {
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&input) != nil || len(input.Password) < 8 {
		fail(c, 400, "password must be at least 8 characters")
		return
	}
	hash, _ := passwordHash(input.Password)
	result := h.DB.Model(&User{}).Where("id = ? AND app_id = ?", c.Param("id"), ApplicationID(c)).Update("password_hash", hash)
	if result.RowsAffected == 0 {
		fail(c, 404, "user not found")
		return
	}
	ok(c, nil)
}
func (h *Handler) remove(c *gin.Context) {
	var count int64
	h.DB.Table("billing_orders").Where("app_id = ? AND user_id = ?", ApplicationID(c), c.Param("id")).Count(&count)
	if count > 0 {
		fail(c, 409, "user with orders cannot be deleted; disable the user instead")
		return
	}
	result := h.DB.Where("id = ? AND app_id = ?", c.Param("id"), ApplicationID(c)).Delete(&User{})
	if result.RowsAffected == 0 {
		fail(c, 404, "user not found")
		return
	}
	ok(c, nil)
}
