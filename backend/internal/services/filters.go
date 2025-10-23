package services

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListFilters holds common query parameters for paginated lists.
type ListFilters struct {
	Page     int
	PageSize int
	Query    string
	SortBy   string
	SortDir  string
	Status   string
	Category string
	UserID   string
	StoreID  string
}

// ParseListFilters extracts list-related query parameters from the Gin context.
func ParseListFilters(c *gin.Context) ListFilters {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "10")) // Use limit for page size
	if pageSize <= 0 {
		pageSize = 10
	}

	// Handle sorting (e.g., "-created_at" -> sort by "created_at" descending)
	sort := c.DefaultQuery("sort", "-created_at")
	sortBy := "created_at"
	sortDir := "desc"
	if strings.HasPrefix(sort, "-") {
		sortBy = strings.TrimPrefix(sort, "-")
		sortDir = "desc"
	} else {
		sortBy = sort
		sortDir = "asc"
	}

	return ListFilters{
		Page:     page,
		PageSize: pageSize,
		Query:    strings.TrimSpace(c.Query("q")), // Use "q" for query
		SortBy:   sortBy,
		SortDir:  sortDir,
		Status:   c.Query("status"),
		Category: c.Query("category"),
		UserID:   c.Query("userId"),
		StoreID:  c.Query("storeId"),
	}
}
