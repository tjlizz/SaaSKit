package app

import (
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMultiApplicationRoutesCanBeRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServer(db, nil, Config{JWTSecret: "test", AutoMigrate: false, FrontendOrigins: []string{"http://localhost"}, AllowUserRegistration: true})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"GET /api/auth/initialization": false, "POST /api/auth/register": false, "POST /api/admin-auth/bootstrap": false, "POST /api/auth/login": false, "GET /api/auth/codes": false, "GET /api/user/info": false, "GET /api/admin/applications": false, "POST /api/admin/applications": false, "POST /api/user-auth/register": false, "GET /api/admin/users": false, "POST /api/client/orders": false, "GET /api/account/subscription": false}
	for _, route := range server.Router.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for route, found := range want {
		if !found {
			t.Errorf("missing route %s", route)
		}
	}
}
