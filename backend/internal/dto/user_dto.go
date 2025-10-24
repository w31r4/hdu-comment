package dto

import (
	"time"

	"github.com/hdu-dp/backend/internal/models"
)

// UserResponse defines the data returned for a user profile.
type UserResponse struct {
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToUserResponse converts a User model to a UserResponse DTO.
func ToUserResponse(user *models.User) UserResponse {
	return UserResponse{
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
	}
}
