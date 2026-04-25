package main

import (
	"fmt"
	"log"

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

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	r := gin.Default()

	authHandler := handlers.NewAuthHandler(cfg.JWTSecret)

	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	authorized := r.Group("/auth")
	authorized.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		authorized.GET("/verify", authHandler.Verify)
		authorized.POST("/refresh", authHandler.Refresh)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Auth service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}