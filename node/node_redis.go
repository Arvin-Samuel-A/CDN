package main

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	client *redis.Client
	centralClient *redis.Client
	serverURL = os.Getenv("SERVER_URL")
	redisAddr = os.Getenv("REDIS_ADDR")
	centralRedisAddr = os.Getenv("CENTRAL_REDIS_ADDR")
)

type RequestAndResponse struct {
	Value string `json:"value"`
}


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
	
	ttl := 24 * time.Hour
	return client.Set(ctx, key, hash, ttl).Err()
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


func NodeGet(key string) (string, error) {
	base64, err := Get(key)
	if err != nil {
		resp, err := http.Get(serverURL + key)
		if err != nil {
			return "", err
		} 
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		var result RequestAndResponse
		if err = json.Unmarshal(body, &result); err != nil {
			return "", err
		}

		if err = Set(key, result.Value); err != nil {
			return "", err
		}

		fmt.Println("Value fetched from Main Server")
		return result.Value, nil
	} else {
		return base64, nil
	}
}

func NodeSet(key, value string) error {
	if err := Set(key, value); err != nil {
		return err
	}

	body := RequestAndResponse{Value: value}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	_, err = http.Post(serverURL + key, "application/json", bytes.NewReader(payload))
	return err
}

func NodeDelete(key string) error {
	if err := Delete(key); err != nil {
		return err
	}

	body := RequestAndResponse{Value: key}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodDelete, serverURL + key, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete, status code: %d", resp.StatusCode)
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