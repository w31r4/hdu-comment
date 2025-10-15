package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/services"
)

// ReviewAdminHandler contains endpoints reserved for administrators.
type ReviewAdminHandler struct {
	reviews *services.ReviewService
}

// NewReviewAdminHandler constructs a new handler.
func NewReviewAdminHandler(reviews *services.ReviewService) *ReviewAdminHandler {
	return &ReviewAdminHandler{reviews: reviews}
}

// Pending lists reviews awaiting approval.
func (h *ReviewAdminHandler) Pending(c *gin.Context) {
	filters := services.ListFilters{
		Page:     mustAtoi(c.DefaultQuery("page", "1")),
		PageSize: mustAtoi(c.DefaultQuery("page_size", "10")),
		Query:    strings.TrimSpace(c.Query("query")),
		SortBy:   c.DefaultQuery("sort", "created_at"),
		SortDir:  c.DefaultQuery("order", "desc"),
	}

	result, err := h.reviews.ListPending(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Approve marks a review as approved.
func (h *ReviewAdminHandler) Approve(c *gin.Context) {
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

	if err := h.reviews.Approve(review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, review)
}

// Reject marks a review as rejected.
func (h *ReviewAdminHandler) Reject(c *gin.Context) {
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

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.reviews.Reject(review, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, review)
}

// Delete removes a review permanently.
func (h *ReviewAdminHandler) Delete(c *gin.Context) {
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

	if err := h.reviews.DeleteReview(c.Request.Context(), review); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func mustAtoi(val string) int {
	n, _ := strconv.Atoi(val)
	return n
}
