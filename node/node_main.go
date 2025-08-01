package main

import (
	"fmt"
	"net/http"
	"strings"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"context"
)

const (
	updateChannel = "update-channel"
	deleteChannel = "delete-channel"
)

func response(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/")
	if key == "" {
		http.Error(w, "Missing the Key", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGet(w, key)
	case http.MethodPost:
		handlePost(w, r, key)
	case http.MethodDelete:
		handleDelete(w, key)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGet(w http.ResponseWriter, key string)  {
	result, err := NodeGet(key);
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := RequestAndResponse{Value: result}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	fmt.Printf("GET - %d\n", http.StatusOK)
}

func handlePost(w http.ResponseWriter, r *http.Request, key string) {
    var data RequestAndResponse
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    if err := NodeSet(key, data.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
    w.WriteHeader(http.StatusOK)
	fmt.Printf("POST - %d\n", http.StatusOK)
}

func handleDelete(w http.ResponseWriter, key string){
	if err := NodeDelete(key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Printf("DELETE - %d\n", http.StatusOK)
}

func KeyExists(key string) (bool, error) {
    count, err := client.Exists(ctx, key).Result()
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

func main() {
	InitRedis()
	InitCentralRedis()

	pubsub, err := Subscribe(updateChannel, deleteChannel)
	if err != nil {
		log.Fatal("Subsription to Pub/Sub Channel Failed")
		return
	}

	ch := pubsub.Channel()
	go func() {
		for msg := range ch {
			switch msg.Channel {
			case updateChannel:
				fmt.Println("Received a Update request from Pub/Sub Channel")
				res := strings.Split(msg.Payload, ";")
				key := strings.TrimPrefix(res[0], "key:")
				value := strings.TrimPrefix(res[1], "value:")
				if exists, _ := KeyExists(key); exists {
					if err := Set(key, value); err != nil {
						log.Print(err)
					}
				}

			case deleteChannel:
				fmt.Println("Received a Delete request from Pub/Sub Channel")
				key := strings.TrimPrefix(msg.Payload, "key:")
				if exists, _ := KeyExists(key); exists {
					if err := Delete(key); err != nil {
						log.Print(err)
					}
				}
			}
		}
	}()

	port := os.Getenv("PORT")
	http.HandleFunc("/", response)
	srv := &http.Server{Addr: ":" + port}

	go func() {
		fmt.Println("Listening on port " + port )
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %s\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	fmt.Println("\nShutting down node...")

	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Node shutdown failed: %s\n", err)
	}

	fmt.Println("Node exited properly.")
}