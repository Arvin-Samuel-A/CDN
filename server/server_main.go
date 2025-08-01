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

// Pull load CDN
// GET (For fetching), POST (For adding new content), DELETE
// sudo service redis-server start

type RequestAndResponse struct {
	Value string `json:"value"`
}

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
	result, err := Get(key);
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

    if err := Set(key, data.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}

    w.WriteHeader(http.StatusOK)
	Publish(updateChannel, fmt.Sprintf("key:%s;value:%s", key, data.Value))
	fmt.Printf("POST - %d\n", http.StatusOK)
}

func handleDelete(w http.ResponseWriter, key string){
	if err := Delete(key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	Publish(deleteChannel, fmt.Sprintf("key:%s", key))
	fmt.Printf("DELETE - %d\n", http.StatusOK)
}

func main() {
	InitRedis()
	InitCentralRedis()

	port := os.Getenv("PORT")
	http.HandleFunc("/", response)
	srv := &http.Server{Addr: ":" + port}

	go func() {
		fmt.Println("Listening on port " + port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %s\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	fmt.Println("\nShutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %s\n", err)
	}

	fmt.Println("Server exited properly.")
}
