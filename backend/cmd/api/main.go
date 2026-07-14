package main

import (
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/saaskit-community/saaskit/internal/app"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg := app.LoadConfig()
	var dialector gorm.Dialector
	if strings.HasPrefix(cfg.DatabaseURL, "sqlite:") {
		dialector = sqlite.Open(strings.TrimPrefix(cfg.DatabaseURL, "sqlite:"))
	} else {
		dialector = postgres.Open(cfg.DatabaseURL)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		options, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			log.Fatal(err)
		}
		redisClient = redis.NewClient(options)
	}
	server, err := app.NewServer(db, redisClient, cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SaaSKit single-product API listening on :%s", cfg.Port)
	log.Fatal(server.Router.Run(":" + cfg.Port))
}
