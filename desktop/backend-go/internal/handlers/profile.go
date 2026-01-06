package handlers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// UploadProfilePhoto handles profile photo upload
func (h *Handlers) UploadProfilePhoto(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse multipart form with 5MB max for profile photos
	if err := c.Request.ParseMultipartForm(5 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 5MB)"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: jpg, jpeg, png, gif, webp"})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := filepath.Join("uploads", "profiles", user.ID)
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("avatar_%d%s", time.Now().Unix(), ext)
	filePath := filepath.Join(uploadsDir, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Build the file URL
	fileURL := fmt.Sprintf("/uploads/profiles/%s/%s", user.ID, filename)

	// Update the user's image in Better Auth user table
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = h.pool.Exec(ctx, `UPDATE "user" SET image = $1, "updatedAt" = NOW() WHERE id = $2`, fileURL, user.ID)
	if err != nil {
		// Still return success since file was uploaded, just log the error
		slog.Warn("failed to update user image in database", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": filename,
		"message":  "Profile photo updated",
	})
}

// GetProfilePhoto serves a profile photo file
func (h *Handlers) GetProfilePhoto(c *gin.Context) {
	userID := c.Param("user_id")
	filename := c.Param("filename")

	if userID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID and filename required"})
		return
	}

	// Sanitize to prevent directory traversal
	userID = filepath.Base(userID)
	filename = filepath.Base(filename)
	filePath := filepath.Join("uploads", "profiles", userID, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(filePath)
}

// DeleteProfilePhoto removes the user's profile photo
func (h *Handlers) DeleteProfilePhoto(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Remove profile photo directory
	profileDir := filepath.Join("uploads", "profiles", user.ID)
	if err := os.RemoveAll(profileDir); err != nil && !os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile photo"})
		return
	}

	// Clear the image in Better Auth user table
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err := h.pool.Exec(ctx, `UPDATE "user" SET image = NULL, "updatedAt" = NOW() WHERE id = $1`, user.ID)
	if err != nil {
		slog.Warn("failed to clear user image in database", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile photo deleted"})
}

// UpdateProfile updates the user's profile information
func (h *Handlers) UpdateProfile(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Update user name in Better Auth user table
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err := h.pool.Exec(ctx, `UPDATE "user" SET name = $1, "updatedAt" = NOW() WHERE id = $2`, req.Name, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated",
		"name":    req.Name,
	})
}

// UploadBackground handles background image upload
func (h *Handlers) UploadBackground(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse multipart form with 10MB max
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 10MB)"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: jpg, jpeg, png, gif, webp"})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := filepath.Join("uploads", "backgrounds", user.ID)
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadsDir, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Return the file path/URL
	// In production, this would be a CDN URL or similar
	fileURL := fmt.Sprintf("/uploads/backgrounds/%s/%s", user.ID, filename)

	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": filename,
	})
}

// GetBackground serves a background image file
func (h *Handlers) GetBackground(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename required"})
		return
	}

	// Sanitize filename to prevent directory traversal
	filename = filepath.Base(filename)
	filePath := filepath.Join("uploads", "backgrounds", user.ID, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(filePath)
}

// DeleteBackground deletes the user's background image
func (h *Handlers) DeleteBackground(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get background directory
	backgroundDir := filepath.Join("uploads", "backgrounds", user.ID)

	// Remove entire directory and contents
	if err := os.RemoveAll(backgroundDir); err != nil && !os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete background"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Background deleted"})
}
