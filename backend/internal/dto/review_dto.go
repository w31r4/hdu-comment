package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
)

// AuthorResponse is a simplified DTO for the author of a review.
type AuthorResponse struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
}

// ReviewResponse is the DTO for a single review response.
type ReviewResponse struct {
	ID        uuid.UUID       `json:"id"`
	Author    AuthorResponse  `json:"author"`
	Title     string          `json:"title"`
	Content   string          `json:"content"`
	Rating    float32         `json:"rating"`
	Images    []ImageResponse `json:"images"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ImageResponse is the DTO for a review image.
type ImageResponse struct {
	ID  uuid.UUID `json:"id"`
	URL string    `json:"url"`
}

// CreateReviewRequest is the DTO for creating a new review.
type CreateReviewRequest struct {
	Title   string  `json:"title" binding:"required"`
	Content string  `json:"content" binding:"required"`
	Rating  float32 `json:"rating" binding:"required,min=0,max=5"`
}

// CreateReviewForNewStoreRequest is used when submitting a review for a store that may not exist yet.
type CreateReviewForNewStoreRequest struct {
	CreateReviewRequest
	StoreName    string `json:"store_name" binding:"required"`
	StoreAddress string `json:"store_address" binding:"required"`
}

// UpdateReviewRequest is the DTO for updating an existing review.
type UpdateReviewRequest struct {
	Title   *string  `json:"title"`
	Content *string  `json:"content"`
	Rating  *float32 `json:"rating" binding:"omitempty,min=0,max=5"`
}

// ToReviewResponse converts a Review model to a ReviewResponse DTO.
func ToReviewResponse(review *models.Review) ReviewResponse {
	images := make([]ImageResponse, len(review.Images))
	for i, img := range review.Images {
		images[i] = ImageResponse{ID: img.ID, URL: img.URL}
	}

	return ReviewResponse{
		ID: review.ID,
		Author: AuthorResponse{
			ID:          review.Author.ID,
			DisplayName: review.Author.DisplayName,
		},
		Title:     review.Title,
		Content:   review.Content,
		Rating:    review.Rating,
		Images:    images,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	}
}

// ToReviewListResponse converts a slice of Review models to a slice of ReviewResponse DTOs.
func ToReviewListResponse(reviews []models.Review) []ReviewResponse {
	res := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		res[i] = ToReviewResponse(&r)
	}
	return res
}

// AutoCreateReviewResponse is the DTO for the response of auto-creating a store with a review.
type AutoCreateReviewResponse struct {
	Store      StoreResponse  `json:"store"`
	Review     ReviewResponse `json:"review"`
	IsNewStore bool           `json:"is_new_store"`
}
