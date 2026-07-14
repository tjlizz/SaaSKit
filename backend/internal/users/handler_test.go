package users

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestUserMiddlewareRejectsAdminToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &Handler{JWTSecret: "test-secret"}
	router := gin.New()
	router.GET("/account", handler.UserAuth(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "admin-1", "kind": "admin", "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("test-secret"))
	req := httptest.NewRequest(http.MethodGet, "/account", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestUserValidationHelpers(t *testing.T) {
	if !validStatus("active") || !validStatus("disabled") || validStatus("pending") {
		t.Fatal("status validation failed")
	}
	if phoneValue(" ") != nil {
		t.Fatal("blank phone should be nil")
	}
	if got := normalizedEmail(" User@Example.COM "); got != "user@example.com" {
		t.Fatalf("unexpected email %s", got)
	}
}
