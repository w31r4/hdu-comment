package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
)

// StoreResponse is the DTO for a single store response.
type StoreResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	Phone         string    `json:"phone"`
	Category      string    `json:"category"`
	Description   string    `json:"description"`
	AverageRating float32   `json:"average_rating"`
	TotalReviews  int       `json:"total_reviews"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateStoreRequest is the DTO for creating a new store.
type CreateStoreRequest struct {
	Name        string `json:"name" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Phone       string `json:"phone"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// ToStoreResponse converts a Store model to a StoreResponse DTO.
func ToStoreResponse(store *models.Store) StoreResponse {
	return StoreResponse{
		ID:            store.ID,
		Name:          store.Name,
		Address:       store.Address,
		Phone:         store.Phone,
		Category:      store.Category,
		Description:   store.Description,
		AverageRating: store.AverageRating,
		TotalReviews:  store.TotalReviews,
		CreatedAt:     store.CreatedAt,
	}
}

// ToStoreListResponse converts a slice of Store models to a slice of StoreResponse DTOs.
func ToStoreListResponse(stores []models.Store) []StoreResponse {
	res := make([]StoreResponse, len(stores))
	for i, s := range stores {
		res[i] = ToStoreResponse(&s)
	}
	return res
}
