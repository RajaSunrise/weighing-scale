package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client based on environment variables.
// REDIS_HOST: defaults to "localhost"
// REDIS_PORT: defaults to "6379"
// REDIS_PASSWORD: defaults to "" (empty)
// REDIS_DB: defaults to 0
func NewRedisClient() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	addr := fmt.Sprintf("%s:%s", host, port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	// Ping to verify connection with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Printf("WARNING: Failed to connect to Redis at %s: %v. Caching will be disabled/broken.", addr, err)
		// We return the client anyway; callers should handle errors or we might want to return nil
		// But usually we want the app to start even if cache is down, depending on strictness.
		// For this plan, we return it.
	} else {
		log.Printf("Connected to Redis at %s", addr)
	}

	return rdb
}
