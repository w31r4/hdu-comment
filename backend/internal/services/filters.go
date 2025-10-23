package services

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListFilters holds common query parameters for paginated lists.
type ListFilters struct {
	Page     int
	Limit    int
	Query    string
	Sort     string
	Status   string
	Category string
	UserID   string
	StoreID  string
}

// ParseListFilters extracts list-related query parameters from the Gin context.
func ParseListFilters(c *gin.Context) ListFilters {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if limit <= 0 {
		limit = 10
	}

	return ListFilters{
		Page:     page,
		Limit:    limit,
		Query:    strings.TrimSpace(c.Query("q")),
		Sort:     c.DefaultQuery("sort", "-created_at"),
		Status:   c.Query("status"),
		Category: c.Query("category"),
		UserID:   c.Query("userId"),
		StoreID:  c.Query("storeId"),
	}
}
