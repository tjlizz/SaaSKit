package app

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saaskit-community/saaskit/internal/users"
)

const applicationIDContextKey = "app_id"

func applicationID(c *gin.Context) string {
	value, _ := c.Get(applicationIDContextKey)
	id, _ := value.(string)
	return id
}

func (s *Server) applicationContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader("X-App-ID"))
		var item Application
		if id == "" || s.DB.Where("id = ? AND status = ?", id, "active").First(&item).Error != nil {
			fail(c, 400, "a valid active application must be selected with X-App-ID")
			return
		}
		c.Set(applicationIDContextKey, item.ID)
		c.Next()
	}
}

func (s *Server) publicApplicationContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader("X-App-Key"))
		var item Application
		if key == "" || s.DB.Where("app_key = ? AND status = ?", key, "active").First(&item).Error != nil {
			fail(c, 400, "a valid active application must be specified with X-App-Key")
			return
		}
		c.Set(applicationIDContextKey, item.ID)
		c.Next()
	}
}

func (s *Server) listApplications(c *gin.Context) {
	var items []Application
	s.DB.Order("created_at asc").Find(&items)
	ok(c, items)
}

func bindApplication(c *gin.Context, item *Application) bool {
	var input struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Status      *string `json:"status"`
	}
	if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Name) == "" {
		fail(c, 400, "name is required")
		return false
	}
	item.Name = strings.TrimSpace(input.Name)
	item.Description = strings.TrimSpace(input.Description)
	if input.Status != nil {
		if *input.Status != "active" && *input.Status != "disabled" {
			fail(c, 400, "status must be active or disabled")
			return false
		}
		item.Status = *input.Status
	}
	return true
}

func (s *Server) createApplication(c *gin.Context) {
	item := Application{ID: uuid.NewString(), AppKey: "app_" + randomSecret(16), Status: "active"}
	if !bindApplication(c, &item) {
		return
	}
	if err := s.DB.Create(&item).Error; err != nil {
		fail(c, 500, "could not create application")
		return
	}
	created(c, item)
}

func (s *Server) updateApplication(c *gin.Context) {
	var item Application
	if s.DB.First(&item, "id = ?", c.Param("id")).Error != nil {
		fail(c, 404, "application not found")
		return
	}
	if !bindApplication(c, &item) {
		return
	}
	if err := s.DB.Save(&item).Error; err != nil {
		fail(c, 500, "could not update application")
		return
	}
	ok(c, item)
}

func (s *Server) deleteApplication(c *gin.Context) {
	id := c.Param("id")
	var count int64
	var userCount int64
	s.DB.Unscoped().Model(&users.User{}).Where("app_id = ?", id).Count(&userCount)
	count += userCount
	models := []any{&APIClient{}, &Plan{}, &Order{}, &Subscription{}}
	for _, model := range models {
		var current int64
		s.DB.Model(model).Where("app_id = ?", id).Count(&current)
		count += current
	}
	if count > 0 {
		fail(c, 409, "application with business data cannot be deleted; disable it instead")
		return
	}
	result := s.DB.Delete(&Application{}, "id = ?", id)
	if result.RowsAffected == 0 {
		fail(c, 404, "application not found")
		return
	}
	ok(c, nil)
}
