package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAdminMiddlewareRejectsRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{Config: Config{JWTSecret: "test-secret"}}
	router := gin.New()
	router.GET("/protected", server.adminAuth(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "admin-1", "kind": "admin_refresh", "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("test-secret"))
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}
