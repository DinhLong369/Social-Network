package app

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func NewRedisClient(addr, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0, // Upstash thường chỉ dùng DB 0

		// BẮT BUỘC: Upstash yêu cầu TLS/SSL để kết nối từ bên ngoài
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},

		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
	})
}

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	pass := os.Getenv("REDIS_PASSWORD")

	if addr == "" || pass == "" {
		log.Fatalf("REDIS_ADDR hoặc REDIS_PASSWORD chưa được cấu hình")
	}

	RedisClient = NewRedisClient(addr, pass)
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Không thể kết nối Redis Upstash: %v", err)
	}

	log.Println("✅ Kết nối Redis Upstash thành công!")
}
