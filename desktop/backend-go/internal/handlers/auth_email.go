package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/services"
	"golang.org/x/crypto/bcrypt"
)

// EmailAuthHandler handles email/password authentication
type EmailAuthHandler struct {
	pool                 *pgxpool.Pool
	cfg                  *config.Config
	notificationTriggers *services.NotificationTriggers
}

// SignUpRequest represents the signup request body
type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

// SignInRequest represents the signin request body
type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// NewEmailAuthHandler creates a new Email Auth handler
func NewEmailAuthHandler(pool *pgxpool.Pool, cfg *config.Config, notifTriggers *services.NotificationTriggers) *EmailAuthHandler {
	return &EmailAuthHandler{
		pool:                 pool,
		cfg:                  cfg,
		notificationTriggers: notifTriggers,
	}
}

// SignUp handles user registration with email/password
func (h *EmailAuthHandler) SignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	var existingID string
	err := h.pool.QueryRow(ctx, `SELECT id FROM "user" WHERE email = $1`, req.Email).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create user
	userID := generateUserID()
	_, err = h.pool.Exec(ctx, `
		INSERT INTO "user" (id, name, email, "emailVerified", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, false, NOW(), NOW())
	`, userID, req.Name, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	// Store password in account table (Better Auth compatible)
	accountID := generateUserID()
	_, err = h.pool.Exec(ctx, `
		INSERT INTO account (id, "userId", "accountId", "providerId", password, "createdAt", "updatedAt")
		VALUES ($1, $2, $3, 'credential', $4, NOW(), NOW())
	`, accountID, userID, userID, string(hashedPassword))
	if err != nil {
		// Rollback user creation
		h.pool.Exec(ctx, `DELETE FROM "user" WHERE id = $1`, userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account: " + err.Error()})
		return
	}

	// Create session
	sessionToken, err := h.createSession(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set session cookie
	c.SetCookie("better-auth.session_token", sessionToken, 60*60*24*7, "/", "", false, true)

	// Send welcome notification
	if h.notificationTriggers != nil {
		go h.notificationTriggers.OnWelcome(context.Background(), userID, req.Name)
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":    userID,
			"email": req.Email,
			"name":  req.Name,
		},
		"message": "Account created successfully",
	})
}

// SignIn handles user login with email/password
func (h *EmailAuthHandler) SignIn(c *gin.Context) {
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get user and password
	var userID, userName, storedPassword string
	err := h.pool.QueryRow(ctx, `
		SELECT u.id, u.name, a.password
		FROM "user" u
		JOIN account a ON a."userId" = u.id
		WHERE u.email = $1 AND a."providerId" = 'credential'
	`, req.Email).Scan(&userID, &userName, &storedPassword)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Create session
	sessionToken, err := h.createSession(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set session cookie
	c.SetCookie("better-auth.session_token", sessionToken, 60*60*24*7, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    userID,
			"email": req.Email,
			"name":  userName,
		},
		"message": "Signed in successfully",
	})
}

// createSession creates a new session for the user
func (h *EmailAuthHandler) createSession(ctx context.Context, userID string) (string, error) {
	sessionToken := generateSessionToken()
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

	_, err := h.pool.Exec(ctx, `
		INSERT INTO session (id, "userId", token, "expiresAt", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, sessionID, userID, sessionToken, expiresAt)

	if err != nil {
		return "", err
	}

	return sessionToken, nil
}
