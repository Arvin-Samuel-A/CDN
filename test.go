// package redisclient

// import (
// 	"context"
// 	"fmt"
// 	"log"

// 	"github.com/redis/go-redis/v9"
// )

// // Global Redis client
// var (
// 	ctx    = context.Background()
// 	client *redis.Client
// )

// // Initialize Redis connection
// func InitRedis() {
// 	client = redis.NewClient(&redis.Options{
// 		Addr:     "localhost:6379",
// 		Password: "", // No password by default
// 		DB:       0,  // Default DB
// 	})

// 	// Test connection
// 	if _, err := client.Ping(ctx).Result(); err != nil {
// 		log.Fatalf("Failed to connect to Redis: %v", err)
// 	}

// 	fmt.Println("Connected to Redis")
// }

// // Set a key-value pair in Redis
// func Set(key, value string) error {
// 	return client.Set(ctx, key, value, 0).Err()
// }

// // Get a value from Redis
// func Get(key string) (string, error) {
// 	return client.Get(ctx, key).Result()
// }


// package main

// import (
// 	"fmt"
// 	"log"

// 	"./redisclient" // Import the redis package
// )

// func main() {
// 	redisclient.InitRedis() // Initialize Redis connection

// 	// Store a value
// 	err := redisclient.Set("key", "Hello from Redis!")
// 	if err != nil {
// 		log.Fatalf("Could not set value: %v", err)
// 	}

// 	// Retrieve the value
// 	val, err := redisclient.Get("key")
// 	if err != nil {
// 		log.Fatalf("Could not get value: %v", err)
// 	}

// 	fmt.Println("Value from Redis:", val)
// }
