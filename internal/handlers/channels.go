package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"chat-server/internal/database"
)

type ChannelHandler struct{}

func NewChannelHandler() *ChannelHandler {
	return &ChannelHandler{}
}

func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required,min=1,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	channelID := uuid.New()

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create channel"})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO channels (id, name, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, channelID, req.Name, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create channel"})
		return
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO channel_members (channel_id, user_id, role, joined_at)
		VALUES ($1, $2, 'admin', NOW())
	`, channelID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member"})
		return
	}

	tx.Commit(ctx)

	c.JSON(http.StatusCreated, gin.H{
		"id":         channelID,
		"name":      req.Name,
		"created_by": userID,
	})
}

func (h *ChannelHandler) GetChannels(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx := context.Background()
	rows, err := database.Pool.Query(ctx, `
		SELECT c.id, c.name, c.created_by, c.created_at, c.updated_at
		FROM channels c
		JOIN channel_members cm ON c.id = cm.channel_id
		WHERE cm.user_id = $1
		ORDER BY c.updated_at DESC
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get channels"})
		return
	}
	defer rows.Close()

	var channels []gin.H
	for rows.Next() {
		var id, createdBy uuid.UUID
		var name string
		var createdAt, updatedAt interface{}

		if err := rows.Scan(&id, &name, &createdBy, &createdAt, &updatedAt); err != nil {
			continue
		}

		channels = append(channels, gin.H{
			"id":          id,
			"name":       name,
			"created_by": createdBy,
		})
	}

	c.JSON(http.StatusOK, channels)
}

func (h *ChannelHandler) JoinChannel(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	channelIDStr := c.Param("channelId")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	ctx := context.Background()

	var channelName string
	err = database.Pool.QueryRow(ctx, `SELECT name FROM channels WHERE id = $1`, channelID).Scan(&channelName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	_, err = database.Pool.Exec(ctx, `
		INSERT INTO channel_members (channel_id, user_id, role, joined_at)
		VALUES ($1, $2, 'member', NOW())
		ON CONFLICT (channel_id, user_id) DO NOTHING
	`, channelID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "joined channel", "channel_id": channelID})
}

func (h *ChannelHandler) LeaveChannel(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	channelIDStr := c.Param("channelId")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	ctx := context.Background()

	_, err = database.Pool.Exec(ctx, `
		DELETE FROM channel_members WHERE channel_id = $1 AND user_id = $2
	`, channelID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to leave channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "left channel"})
}

func (h *ChannelHandler) GetChannelMembers(c *gin.Context) {
	channelIDStr := c.Param("channelId")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	ctx := context.Background()
	rows, err := database.Pool.Query(ctx, `
		SELECT cm.user_id, u.username, cm.role, cm.joined_at
		FROM channel_members cm
		JOIN users u ON cm.user_id = u.id
		WHERE cm.channel_id = $1
	`, channelID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get members"})
		return
	}
	defer rows.Close()

	var members []gin.H
	for rows.Next() {
		var userID uuid.UUID
		var username, role string
		var joinedAt interface{}

		if err := rows.Scan(&userID, &username, &role, &joinedAt); err != nil {
			continue
		}

		members = append(members, gin.H{
			"user_id":  userID,
			"username": username,
			"role":    role,
		})
	}

	c.JSON(http.StatusOK, members)
}

func (h *ChannelHandler) GetChannel(c *gin.Context) {
	channelIDStr := c.Param("channelId")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	ctx := context.Background()
	var id, createdBy uuid.UUID
	var name string
	var createdAt, updatedAt interface{}

	err = database.Pool.QueryRow(ctx, `
		SELECT id, name, created_by, created_at, updated_at FROM channels WHERE id = $1
	`, channelID).Scan(&id, &name, &createdBy, &createdAt, &updatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          id,
		"name":       name,
		"created_by": createdBy,
	})
}

func (h *ChannelHandler) SearchChannels(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query required"})
		return
	}

	ctx := context.Background()
	rows, err := database.Pool.Query(ctx, `
		SELECT c.id, c.name, c.created_by, c.created_at
		FROM channels c
		WHERE c.name ILIKE $1
		LIMIT 20
	`, "%"+query+"%")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}
	defer rows.Close()

	var channels []gin.H
	for rows.Next() {
		var id, createdBy uuid.UUID
		var name string
		var createdAt interface{}

		if err := rows.Scan(&id, &name, &createdBy, &createdAt); err != nil {
			continue
		}

		channels = append(channels, gin.H{
			"id":          id,
			"name":       name,
			"created_by": createdBy,
		})
	}

	c.JSON(http.StatusOK, channels)
}