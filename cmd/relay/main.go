package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"chat-server/internal/config"
	"chat-server/internal/database"
	"chat-server/internal/handlers"
	"chat-server/internal/middleware"
)

func main() {
	godotenv.Load()

	cfg := config.Load()
	cfg.Port = os.Getenv("RELAY_PORT")
	if cfg.Port == "" {
		cfg.Port = "8003"
	}

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	r := gin.Default()
	relayHandler := handlers.NewRelayHandler()

	r.GET("/health", relayHandler.HealthCheck)

	ws := r.Group("/ws")
	ws.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		ws.GET("/connect", relayHandler.HandleWebSocket)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Message relay service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}