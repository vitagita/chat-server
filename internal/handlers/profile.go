package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"chat-server/internal/database"
)

type ProfileHandler struct{}

func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	ctx := context.Background()
	var userIDOut uuid.UUID
	var displayName, avatarURL, status, phone *string
	var username string
	var updatedAt interface{}

	err = database.Pool.QueryRow(ctx, `
		SELECT p.user_id, p.display_name, COALESCE(p.avatar_url, ''), COALESCE(p.status, ''), COALESCE(p.phone, ''), p.updated_at, u.username
		FROM profiles p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = $1
	`, userID).Scan(&userIDOut, &displayName, &avatarURL, &status, &phone, &updatedAt, &username)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":      userIDOut,
		"username":    username,
		"display_name": displayName,
		"avatar_url":  avatarURL,
		"status":     status,
		"phone":     phone,
	})
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if userID != targetUserID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot update other user's profile"})
		return
	}

	var req models.ProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	_, err = database.Pool.Exec(ctx, `
		UPDATE profiles 
		SET display_name = COALESCE(NULLIF($1, ''), display_name),
		    avatar_url = COALESCE(NULLIF($2, ''), avatar_url),
		    status = COALESCE(NULLIF($3, ''), status),
		    phone = COALESCE(NULLIF($4, ''), phone),
		    updated_at = NOW()
		WHERE user_id = $5
	`, req.DisplayName, req.AvatarURL, req.Status, req.Phone, targetUserID)

	_, err = database.Pool.Exec(ctx, `
		UPDATE users 
		SET phone = COALESCE(NULLIF($1, ''), phone),
		    updated_at = NOW()
		WHERE id = $2
	`, req.Phone, targetUserID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func (h *ProfileHandler) SearchProfiles(c *gin.Context) {
	query := c.Query("q")
	query = strings.TrimSpace(query)
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query required"})
		return
	}

	ctx := context.Background()
	
	searchPattern := "%" + query + "%"
	fmt.Printf("DEBUG: SearchProfiles query=%s pattern=%s\n", query, searchPattern)
	
	rows, err := database.Pool.Query(ctx, `
		SELECT u.id, u.username, p.display_name, p.avatar_url, p.status, COALESCE(p.phone, ''), COALESCE(u.phone, '')
		FROM users u
		LEFT JOIN profiles p ON u.id = p.user_id
		WHERE u.username ILIKE $1 OR u.phone ILIKE $1 OR p.phone ILIKE $1
		LIMIT 20
	`, searchPattern)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}
	defer rows.Close()

	var results []gin.H
	for rows.Next() {
		var userID uuid.UUID
		var username, displayName, avatarURL, status, phone, userPhone *string

		if err := rows.Scan(&userID, &username, &displayName, &avatarURL, &status, &phone, &userPhone); err != nil {
			continue
		}

		phoneVal := *phone
		if phoneVal == "" && userPhone != nil {
			phoneVal = *userPhone
		}

		results = append(results, gin.H{
			"user_id":      userID,
			"username":    username,
			"display_name": displayName,
			"avatar_url":  avatarURL,
			"status":      status,
			"phone":      phoneVal,
		})
	}

	fmt.Printf("DEBUG: SearchProfiles found %d results\n", len(results))
	c.JSON(http.StatusOK, results)
}

func (h *ProfileHandler) SearchByPhone(c *gin.Context) {
	phone := c.Param("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone required"})
		return
	}

	ctx := context.Background()
	
	rows, err := database.Pool.Query(ctx, `
		SELECT u.id, u.username, p.display_name, p.avatar_url, p.status, COALESCE(p.phone, ''), COALESCE(u.phone, '')
		FROM users u
		LEFT JOIN profiles p ON u.id = p.user_id
		WHERE p.phone = $1 OR u.phone = $1
	`, phone)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}
	defer rows.Close()

	var results []gin.H
	for rows.Next() {
		var userID uuid.UUID
		var username, displayName, avatarURL, status, phone, userPhone *string

		if err := rows.Scan(&userID, &username, &displayName, &avatarURL, &status, &phone, &userPhone); err != nil {
			continue
		}

		phoneVal := *phone
		if phoneVal == "" && userPhone != nil {
			phoneVal = *userPhone
		}

		results = append(results, gin.H{
			"user_id":      userID,
			"username":    username,
			"display_name": displayName,
			"avatar_url":  avatarURL,
			"status":      status,
			"phone":      phoneVal,
		})
	}

	c.JSON(http.StatusOK, results)
}