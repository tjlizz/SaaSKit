//go:build integration

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/saaskit-community/saaskit/internal/users"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type apiEnvelope struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

func testServer(t *testing.T) *Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{JWTSecret: "test-secret", PublicURL: "http://saaskit.test", PaymentMock: true, AutoMigrate: true, AllowUserRegistration: true}
	cfg.PaymentKey = [32]byte{1, 2, 3}
	server, err := NewServer(db, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return server
}
func call(t *testing.T, s *Server, method, path string, body any, headers map[string]string) (int, apiEnvelope) {
	t.Helper()
	var raw []byte
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	s.Router.ServeHTTP(rec, req)
	var env apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil && rec.Code/100 == 2 {
		t.Fatalf("invalid JSON response %s: %s", path, rec.Body.String())
	}
	return rec.Code, env
}
func decode[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	return value
}
func bootstrap(t *testing.T, s *Server) (string, APIClient, string, users.User) {
	status, env := call(t, s, "POST", "/api/admin-auth/bootstrap", map[string]any{"email": "admin@example.com", "password": "password123", "name": "Admin"}, nil)
	if status != 201 {
		t.Fatalf("bootstrap: %d %s", status, env.Message)
	}
	auth := decode[map[string]any](t, env.Data)
	token := auth["access_token"].(string)
	application := auth["application"].(map[string]any)
	admin := map[string]string{"Authorization": "Bearer " + token, "X-App-ID": application["id"].(string)}
	status, env = call(t, s, "POST", "/api/admin/api-clients", map[string]any{"name": "Backend"}, admin)
	if status != 201 {
		t.Fatalf("client: %d %s", status, env.Message)
	}
	clientResult := decode[struct {
		Client APIClient `json:"client"`
		Secret string    `json:"client_secret"`
	}](t, env.Data)
	status, env = call(t, s, "POST", "/api/admin/users", map[string]any{"email": "user@example.com", "password": "password123", "name": "User"}, admin)
	if status != 201 {
		t.Fatalf("user: %d %s", status, env.Message)
	}
	user := decode[users.User](t, env.Data)
	return token, clientResult.Client, clientResult.Secret, user
}

func TestApplicationsIsolateUsersAndPlanCodes(t *testing.T) {
	s := testServer(t)
	token, _, _, firstUser := bootstrap(t, s)
	baseHeaders := map[string]string{"Authorization": "Bearer " + token}
	status, env := call(t, s, "POST", "/api/admin/applications", map[string]any{"name": "Second App"}, baseHeaders)
	if status != http.StatusCreated {
		t.Fatalf("application: %d %s", status, env.Message)
	}
	second := decode[Application](t, env.Data)
	secondHeaders := map[string]string{"Authorization": "Bearer " + token, "X-App-ID": second.ID}
	status, env = call(t, s, "POST", "/api/admin/users", map[string]any{"email": firstUser.Email, "password": "password123", "name": "Second User"}, secondHeaders)
	if status != http.StatusCreated {
		t.Fatalf("same email in second application: %d %s", status, env.Message)
	}
	secondUser := decode[users.User](t, env.Data)
	if secondUser.AppID != second.ID || secondUser.ID == firstUser.ID {
		t.Fatalf("unexpected second user: %+v", secondUser)
	}
	for _, headers := range []map[string]string{{"Authorization": "Bearer " + token, "X-App-ID": firstUser.AppID}, secondHeaders} {
		status, env = call(t, s, "POST", "/api/admin/plans", map[string]any{"plan_code": "pro", "name": "Pro", "billing_cycle": "monthly", "price_cents": 9900}, headers)
		if status != http.StatusCreated {
			t.Fatalf("same plan code across applications: %d %s", status, env.Message)
		}
	}
}

func TestFreePlanCreatesSubscriptionForOwnedUser(t *testing.T) {
	s := testServer(t)
	token, client, secret, user := bootstrap(t, s)
	admin := map[string]string{"Authorization": "Bearer " + token}
	status, env := call(t, s, "POST", "/api/admin/plans", map[string]any{"plan_code": "free", "name": "Free", "billing_cycle": "free", "price_cents": 0, "enabled": true}, admin)
	if status != 201 {
		t.Fatalf("plan: %d %s", status, env.Message)
	}
	apiHeaders := map[string]string{"X-API-Key": client.ClientKey, "X-API-Secret": secret}
	status, env = call(t, s, "POST", "/api/client/orders", map[string]any{"plan_code": "free", "user_id": user.ID}, apiHeaders)
	if status != 201 {
		t.Fatalf("order: %d %s", status, env.Message)
	}
	order := decode[struct {
		Order Order `json:"order"`
	}](t, env.Data).Order
	if order.Status != "paid" || order.UserID != user.ID {
		t.Fatalf("unexpected order: %+v", order)
	}
	status, env = call(t, s, "GET", "/api/client/subscription/check?user_id="+user.ID, nil, apiHeaders)
	check := decode[struct {
		Valid bool `json:"valid"`
	}](t, env.Data)
	if status != 200 || !check.Valid {
		t.Fatalf("subscription check failed: %d %s", status, env.Message)
	}
}

func TestMockNotifyIsIdempotent(t *testing.T) {
	s := testServer(t)
	token, client, secret, user := bootstrap(t, s)
	admin := map[string]string{"Authorization": "Bearer " + token}
	call(t, s, "POST", "/api/admin/plans", map[string]any{"plan_code": "pro", "name": "Pro", "billing_cycle": "monthly", "price_cents": 9900, "enabled": true}, admin)
	apiHeaders := map[string]string{"X-API-Key": client.ClientKey, "X-API-Secret": secret}
	status, env := call(t, s, "POST", "/api/client/orders", map[string]any{"plan_code": "pro", "user_id": user.ID, "provider": "alipay"}, apiHeaders)
	if status != 201 {
		t.Fatalf("order: %d %s", status, env.Message)
	}
	result := decode[struct {
		Order   Order         `json:"order"`
		Payment PaymentResult `json:"payment"`
	}](t, env.Data)
	paymentURL, _ := url.Parse(result.Payment.PaymentURL)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, paymentURL.Path+"?"+paymentURL.RawQuery, nil)
		rec := httptest.NewRecorder()
		s.Router.ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Fatalf("notify %d: %d %s", i, rec.Code, rec.Body.String())
		}
	}
	var count int64
	s.DB.Model(&Subscription{}).Where("user_id = ?", user.ID).Count(&count)
	if count != 1 {
		t.Fatalf("want one subscription, got %d", count)
	}
}

func TestRejectsInvalidAPISecret(t *testing.T) {
	s := testServer(t)
	_, client, _, user := bootstrap(t, s)
	status, _ := call(t, s, "GET", "/api/client/subscription/check?user_id="+user.ID, nil, map[string]string{"X-API-Key": client.ClientKey, "X-API-Secret": "wrong"})
	if status != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", status)
	}
}

func TestVbenFirstUserRegistrationInitializesSuperAdmin(t *testing.T) {
	s := testServer(t)
	status, env := call(t, s, "GET", "/api/auth/initialization", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("initialization status: %d %s", status, env.Message)
	}
	initialization := decode[struct {
		Initialized          bool `json:"initialized"`
		RegistrationRequired bool `json:"registration_required"`
	}](t, env.Data)
	if initialization.Initialized || !initialization.RegistrationRequired {
		t.Fatalf("unexpected empty instance status: %+v", initialization)
	}

	status, env = call(t, s, "POST", "/api/auth/login", map[string]any{"username": "owner@example.com", "password": "password123"}, nil)
	if status != http.StatusPreconditionRequired {
		t.Fatalf("login before initialization: want 428, got %d", status)
	}

	status, env = call(t, s, "POST", "/api/auth/register", map[string]any{"email": "owner@example.com", "name": "Owner", "password": "password123"}, nil)
	if status != http.StatusCreated {
		t.Fatalf("first registration: %d %s", status, env.Message)
	}
	result := decode[struct {
		AccessToken string `json:"accessToken"`
	}](t, env.Data)
	if result.AccessToken == "" {
		t.Fatal("missing Vben accessToken")
	}
	status, env = call(t, s, "GET", "/api/user/info", nil, map[string]string{"Authorization": "Bearer " + result.AccessToken})
	if status != http.StatusOK {
		t.Fatalf("user info: %d %s", status, env.Message)
	}
	info := decode[struct {
		Username string   `json:"username"`
		HomePath string   `json:"homePath"`
		Roles    []string `json:"roles"`
	}](t, env.Data)
	if info.Username != "owner@example.com" || info.HomePath != "/dashboard/analytics" || len(info.Roles) != 1 || info.Roles[0] != superAdminRole {
		t.Fatalf("unexpected user info: %+v", info)
	}

	status, env = call(t, s, "GET", "/api/auth/initialization", nil, nil)
	initialization = decode[struct {
		Initialized          bool `json:"initialized"`
		RegistrationRequired bool `json:"registration_required"`
	}](t, env.Data)
	if status != http.StatusOK || !initialization.Initialized || initialization.RegistrationRequired {
		t.Fatalf("unexpected initialized status: %d %+v", status, initialization)
	}

	status, _ = call(t, s, "POST", "/api/auth/register", map[string]any{"email": "second@example.com", "name": "Second", "password": "password123"}, nil)
	if status != http.StatusConflict {
		t.Fatalf("second registration: want 409, got %d", status)
	}
}
