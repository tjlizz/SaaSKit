package users

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            string         `gorm:"primaryKey;size:36" json:"id"`
	Email         string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Phone         *string        `gorm:"uniqueIndex;size:32" json:"phone,omitempty"`
	Name          string         `gorm:"size:120;not null" json:"name"`
	AvatarURL     string         `gorm:"size:500" json:"avatar_url"`
	PasswordHash  string         `json:"-"`
	Status        string         `gorm:"index;size:20;not null;default:active" json:"status"`
	EmailVerified bool           `json:"email_verified"`
	PhoneVerified bool           `json:"phone_verified"`
	LastLoginAt   *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
