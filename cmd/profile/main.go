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
	cfg.Port = os.Getenv("PROFILE_PORT")
	if cfg.Port == "" {
		cfg.Port = "8002"
	}

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	r := gin.Default()
	profileHandler := handlers.NewProfileHandler()
	channelHandler := handlers.NewChannelHandler()

	// Public endpoints (no auth required)
	r.GET("/profiles/search", profileHandler.SearchProfiles)
	r.GET("/profiles/phone/:phone", profileHandler.SearchByPhone)
	r.GET("/channels/search", channelHandler.SearchChannels)

	// Protected endpoints
	authorized := r.Group("/profiles")
	authorized.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		authorized.GET("/:userId", profileHandler.GetProfile)
		authorized.PUT("/:userId", profileHandler.UpdateProfile)
	}

	channels := r.Group("/channels")
	channels.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		channels.POST("", channelHandler.CreateChannel)
		channels.GET("", channelHandler.GetChannels)
		channels.GET("/:channelId", channelHandler.GetChannel)
		channels.POST("/:channelId/join", channelHandler.JoinChannel)
		channels.POST("/:channelId/leave", channelHandler.LeaveChannel)
		channels.GET("/:channelId/members", channelHandler.GetChannelMembers)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Profile service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}