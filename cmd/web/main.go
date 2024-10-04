package main

import (
	"log"
	"net/http"

	"github.com/konan69/websocket-go/internal/handlers"
)

func main() {
	mux := routes()

	log.Println("Starting server")
	go handlers.ListenToWsChannel()

	log.Println("Listening on port 8080")
	_ = http.ListenAndServe(":8080", mux)
}