package main

import(
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"crypto/sha512"
	"os"
)

var (
	ctx = context.Background()
	client *redis.Client
	centralClient *redis.Client
	redisAddr = os.Getenv("REDIS_ADDR")
	centralRedisAddr = os.Getenv("CENTRAL_REDIS_ADDR")
)

func GetHash(base64 string) string {
	hasher := sha512.New()
	hasher.Write([]byte(base64))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func InitRedis() {
	client = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		Password: "",
		DB: 0,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		fmt.Println("Failed to connect to Redis: " + err.Error())
		return
	}

	fmt.Println("Connected to Redis")
}

func Get(key string) (string, error) {
	hash, err := client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	} 

	return client.Get(ctx, hash).Result()
}

func Set(key, value string) error {
	hash := GetHash(value)

	if _, err := client.Get(ctx, hash).Result(); err != nil {
		client.Set(ctx, hash, value, 0).Err()
	}
	
	return client.Set(ctx, key, hash, 0).Err()
}

func Delete(key string) error {
	hash, err := client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	if _, err := client.Del(ctx, key).Result(); err != nil {
		return err
	}

	count, err := client.Keys(ctx, "*").Result()
	if err != nil {
		return err
	}

	otherReferences := false
	for _, k := range count {
		h, _ := client.Get(ctx, k).Result()
		if h == hash {
			otherReferences = true
			break
		}
	}

	if !otherReferences {
		client.Del(ctx, hash)
	}

	return nil
}

func InitCentralRedis() {
	centralClient = redis.NewClient(&redis.Options{
		Addr: centralRedisAddr,
		Password: "",
		DB: 0,
	})

	if _, err := centralClient.Ping(ctx).Result(); err != nil {
		fmt.Println("Failed to connect to Redis: " + err.Error())
		return
	}

	fmt.Println("Connected to Central Redis")
}

func Publish(channel, message string) error {
	return centralClient.Publish(ctx, channel, message).Err()
}

func Subscribe(channel1, channel2 string) (*redis.PubSub, error) {
	pubsub := centralClient.Subscribe(ctx, channel1, channel2)
	if _, err := pubsub.Receive(ctx); err != nil {
		return nil, err
	}
	return pubsub, nil
}