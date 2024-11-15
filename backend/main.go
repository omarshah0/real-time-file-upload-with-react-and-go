package main

import (
	"log"
	"net/http"

	"example.com/handlers"

	"github.com/rs/cors"
)

func main() {
	mux := http.NewServeMux()

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST"},
	})

	// WebSocket endpoint
	mux.HandleFunc("/ws", handlers.HandleWebSocket)

	handler := c.Handler(mux)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
