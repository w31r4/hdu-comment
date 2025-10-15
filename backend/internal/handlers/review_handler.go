package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/services"
	"github.com/hdu-dp/backend/internal/storage"
)

// ReviewHandler manages review related HTTP endpoints.
type ReviewHandler struct {
	reviews *services.ReviewService
}

// NewReviewHandler constructs a ReviewHandler.
func NewReviewHandler(reviews *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviews: reviews}
}

// ListPublic returns approved reviews for public consumption.
func (h *ReviewHandler) ListPublic(c *gin.Context) {
	filters := parseListFilters(c)
	result, err := h.reviews.ListPublic(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Submit allows authenticated users to submit a new review.
func (h *ReviewHandler) Submit(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	var req struct {
		Title       string  `json:"title"`
		Address     string  `json:"address"`
		Description string  `json:"description"`
		Rating      float32 `json:"rating"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	review, err := h.reviews.Submit(userID, services.CreateReviewInput{
		Title:       req.Title,
		Address:     req.Address,
		Description: req.Description,
		Rating:      req.Rating,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// Detail returns a single review by id.
func (h *ReviewHandler) Detail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	review, err := h.reviews.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	if review.Status != models.ReviewStatusApproved {
		roleVal, ok := c.Get("role")
		role := ""
		if ok {
			role, _ = roleVal.(string)
		}
		if role != "admin" {
			userVal, ok := c.Get("user_id")
			userID, okID := userVal.(uuid.UUID)
			if !ok || !okID || review.AuthorID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "review not accessible"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, review)
}

// MyReviews returns reviews of the authenticated user.
func (h *ReviewHandler) MyReviews(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	filters := parseListFilters(c)
	result, err := h.reviews.ListByAuthor(userID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// UploadImage handles multipart uploads for review images.
func (h *ReviewHandler) UploadImage(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	review, err := h.reviews.Get(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	if err := services.ValidateOwnership(review, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not owner"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	opened, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uploadFile := &storage.UploadFile{
		Reader:      opened,
		Size:        fileHeader.Size,
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
	}

	image, err := h.reviews.StoreImage(c.Request.Context(), reviewID, uploadFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, image)
}

func parseListFilters(c *gin.Context) services.ListFilters {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	query := strings.TrimSpace(c.Query("query"))
	sortBy := c.DefaultQuery("sort", "created_at")
	sortDir := c.DefaultQuery("order", "desc")

	return services.ListFilters{
		Page:     page,
		PageSize: pageSize,
		Query:    query,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}
}
