package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/dto"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
	"gorm.io/gorm"
)

// StoreService contains business logic around store workflows.
type StoreService struct {
	stores  *repository.StoreRepository
	reviews *repository.ReviewRepository
	db      *gorm.DB
}

// NewStoreService constructs a store service instance.
func NewStoreService(stores *repository.StoreRepository, reviews *repository.ReviewRepository, db *gorm.DB) *StoreService {
	return &StoreService{stores: stores, reviews: reviews, db: db}
}

// StoreListResult wraps store list responses with pagination info.
type StoreListResult struct {
	Data       []dto.StoreResponse `json:"data"`
	Pagination Pagination          `json:"pagination"`
}

// CreateStore creates a new store. Admins can create approved stores directly.
func (s *StoreService) CreateStore(ctx context.Context, createdBy uuid.UUID, isAdmin bool, req dto.CreateStoreRequest) (*models.Store, error) {
	name := strings.TrimSpace(req.Name)
	address := strings.TrimSpace(req.Address)
	category := strings.TrimSpace(req.Category)

	if name == "" || address == "" || category == "" {
		return nil, errors.New("store name, address, and category are required")
	}

	// Check if store already exists
	existing, err := s.stores.FindByNameAndAddress(name, address)
	if err == nil && existing != nil {
		return nil, errors.New("store already exists")
	}

	status := models.StoreStatusPending
	if isAdmin {
		status = models.StoreStatusApproved
	}

	store := &models.Store{
		Name:        name,
		Address:     address,
		Phone:       strings.TrimSpace(req.Phone),
		Category:    category,
		Description: strings.TrimSpace(req.Description),
		Status:      status,
		CreatedBy:   createdBy,
	}

	if err := s.stores.Create(store); err != nil {
		return nil, err
	}
	return store, nil
}

// ListStores returns stores based on filters.
func (s *StoreService) ListStores(filters ListFilters) (StoreListResult, error) {
	offset := (filters.Page - 1) * filters.PageSize

	// Convert status string to model type
	var statuses []models.StoreStatus
	if filters.Status != "" {
		statuses = []models.StoreStatus{models.StoreStatus(filters.Status)}
	}

	stores, total, err := s.stores.SearchStores(repository.StoreSearchFilters{
		Query:    filters.Query,
		Statuses: statuses,
		Category: filters.Category,
		SortBy:   filters.SortBy,
		SortDir:  filters.SortDir,
		Limit:    filters.PageSize,
		Offset:   offset,
	})
	if err != nil {
		return StoreListResult{}, err
	}

	totalPages := int((total + int64(filters.PageSize) - 1) / int64(filters.PageSize))

	return StoreListResult{
		Data: dto.ToStoreListResponse(stores),
		Pagination: Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// ListPending returns pending stores for admin review.
func (s *StoreService) ListPending(page, pageSize int) (StoreListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	stores, total, err := s.stores.ListByStatus([]models.StoreStatus{models.StoreStatusPending}, pageSize, offset)
	if err != nil {
		return StoreListResult{}, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return StoreListResult{
		Data: dto.ToStoreListResponse(stores),
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// Get returns a store by ID.
func (s *StoreService) Get(id uuid.UUID) (*models.Store, error) {
	return s.stores.FindByID(id)
}

// Approve marks a store as approved.
func (s *StoreService) Approve(store *models.Store) error {
	if store.Status == models.StoreStatusApproved {
		return nil // Already approved, idempotent success
	}
	if store.Status != models.StoreStatusPending {
		return fmt.Errorf("store already processed")
	}
	store.Status = models.StoreStatusApproved
	store.RejectionReason = ""
	return s.stores.Update(store)
}

// Reject marks a store as rejected with reason.
func (s *StoreService) Reject(store *models.Store, reason string) error {
	if store.Status == models.StoreStatusRejected {
		return nil // Already rejected, idempotent success
	}
	if store.Status != models.StoreStatusPending {
		return fmt.Errorf("store already processed")
	}
	store.Status = models.StoreStatusRejected
	store.RejectionReason = strings.TrimSpace(reason)
	return s.stores.Update(store)
}

// UpdateStoreRating recalculates the average rating for a store.
func (s *StoreService) UpdateStoreRating(ctx context.Context, storeID uuid.UUID) error {
	// Get all approved reviews for this store
	var result struct {
		AverageRating float32
		TotalReviews  int64
	}

	err := s.db.Model(&models.Review{}).
		Where("store_id = ? AND status = ?", storeID, models.ReviewStatusApproved).
		Select("AVG(rating) as average_rating, COUNT(*) as total_reviews").
		Scan(&result).Error

	if err != nil {
		return err
	}

	return s.stores.UpdateAverageRating(storeID, result.AverageRating, int(result.TotalReviews))
}

// DeleteStore removes a store by ID.
func (s *StoreService) DeleteStore(id uuid.UUID) error {
	return s.stores.Delete(id)
}
