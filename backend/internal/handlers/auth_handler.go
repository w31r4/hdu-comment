package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/common"
	"github.com/hdu-dp/backend/internal/services"
)

// AuthHandler exposes HTTP endpoints for authentication flows.
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registrations.
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result, err := h.authService.Register(req.Email, req.Password, req.DisplayName)
	if err != nil {
		switch err {
		case common.ErrEmailAlreadyUsed:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	respondAuthSuccess(c, http.StatusCreated, result)
}

// Login handles user login and token issuance.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		switch err {
		case common.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	respondAuthSuccess(c, http.StatusOK, result)
}

// Refresh rotates refresh tokens and returns new access/refresh pair.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		switch err {
		case common.ErrInvalidRefreshToken:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	respondAuthSuccess(c, http.StatusOK, result)
}

// Logout revokes the given refresh token.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		switch err {
		case common.ErrInvalidRefreshToken:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func respondAuthSuccess(c *gin.Context, status int, result *services.AuthResult) {
	c.JSON(status, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user": gin.H{
			"id":           result.User.ID,
			"email":        result.User.Email,
			"display_name": result.User.DisplayName,
			"role":         result.User.Role,
			"created_at":   result.User.CreatedAt,
		},
	})
}
