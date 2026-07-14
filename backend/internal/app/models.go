package app

import (
	"time"

	"github.com/saaskit-community/saaskit/internal/users"
)

// Admin is an operator of this self-hosted SaaS instance.
type Admin struct {
	ID           string     `gorm:"primaryKey;size:36" json:"id"`
	Email        string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Name         string     `gorm:"size:120;not null" json:"name"`
	Role         string     `gorm:"index;size:32;not null;default:super_admin" json:"role"`
	PasswordHash string     `json:"-"`
	Status       string     `gorm:"index;size:20;not null;default:active" json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Application is an independently operated SaaS product in this deployment.
type Application struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	AppKey      string    `gorm:"uniqueIndex;size:64;not null" json:"app_key"`
	Description string    `gorm:"size:1000" json:"description"`
	Status      string    `gorm:"index;size:20;not null;default:active" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// APIClient is a server-to-server credential bound to one application.
type APIClient struct {
	ID              string    `gorm:"primaryKey;size:36" json:"id"`
	AppID           string    `gorm:"index;size:36;not null" json:"app_id"`
	Name            string    `gorm:"size:120;not null" json:"name"`
	ClientKey       string    `gorm:"uniqueIndex;size:64;not null" json:"client_key"`
	SecretHash      string    `json:"-"`
	Enabled         bool      `gorm:"not null" json:"enabled"`
	RateLimitPerMin int       `gorm:"not null;default:300" json:"rate_limit_per_min"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Plan struct {
	ID                 string    `gorm:"primaryKey;size:36" json:"id"`
	AppID              string    `gorm:"uniqueIndex:ux_product_plans_app_code;size:36;not null" json:"app_id"`
	PlanCode           string    `gorm:"uniqueIndex:ux_product_plans_app_code;size:80;not null" json:"plan_code"`
	Name               string    `gorm:"size:120;not null" json:"name"`
	Description        string    `gorm:"size:1000" json:"description"`
	BillingCycle       string    `gorm:"size:32;not null" json:"billing_cycle"`
	PriceCents         int64     `gorm:"not null" json:"price_cents"`
	Currency           string    `gorm:"size:8;default:CNY" json:"currency"`
	AutoRenewSupported bool      `json:"auto_renew_supported"`
	DeviceLimit        int       `gorm:"default:1" json:"device_limit"`
	CreditQuota        int64     `json:"credit_quota"`
	SeatLimit          int       `json:"seat_limit"`
	Recommended        bool      `json:"recommended"`
	Enabled            bool      `gorm:"not null" json:"enabled"`
	SortOrder          int       `json:"sort_order"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (Plan) TableName() string { return "product_plans" }

type Order struct {
	ID              string     `gorm:"primaryKey;size:36" json:"id"`
	AppID           string     `gorm:"index;size:36;not null" json:"app_id"`
	OrderNo         string     `gorm:"uniqueIndex;size:64;not null" json:"order_no"`
	UserID          string     `gorm:"index;size:36;not null" json:"user_id"`
	User            users.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PlanID          string     `gorm:"index;size:36;not null" json:"plan_id"`
	Provider        string     `gorm:"size:20;not null" json:"provider"`
	ProviderTradeNo string     `gorm:"size:128" json:"provider_trade_no"`
	Status          string     `gorm:"index;size:20;not null" json:"status"`
	AmountCents     int64      `json:"amount_cents"`
	Currency        string     `gorm:"size:8" json:"currency"`
	Quantity        int        `gorm:"default:1" json:"quantity"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (Order) TableName() string { return "billing_orders" }

type Subscription struct {
	ID                 string     `gorm:"primaryKey;size:36" json:"id"`
	AppID              string     `gorm:"uniqueIndex:ux_subscriptions_app_user;size:36;not null" json:"app_id"`
	UserID             string     `gorm:"uniqueIndex:ux_subscriptions_app_user;size:36;not null" json:"user_id"`
	User               users.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PlanID             string     `gorm:"index;size:36;not null" json:"plan_id"`
	OrderID            string     `gorm:"index;size:36;not null" json:"order_id"`
	SubscriptionStatus string     `gorm:"index;size:20;not null" json:"subscription_status"`
	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end"`
	AutoRenew          bool       `json:"auto_renew"`
	RemainingCredits   int64      `json:"remaining_credits"`
	Plan               Plan       `gorm:"foreignKey:PlanID" json:"plan"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (Subscription) TableName() string { return "user_subscriptions" }

type PaymentConfig struct {
	ID         string    `gorm:"primaryKey;size:36" json:"id"`
	Provider   string    `gorm:"uniqueIndex;size:20;not null" json:"provider"`
	Enabled    bool      `json:"enabled"`
	Sandbox    bool      `json:"sandbox"`
	Ciphertext string    `gorm:"type:text" json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (PaymentConfig) TableName() string { return "merchant_payment_configs" }

func Models() []any {
	return []any{&Admin{}, &Application{}, &APIClient{}, &users.User{}, &Plan{}, &Order{}, &Subscription{}, &PaymentConfig{}}
}
