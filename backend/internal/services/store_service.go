package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
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

// CreateStoreInput bundles parameters for a new store.
type CreateStoreInput struct {
	Name        string
	Address     string
	Phone       string
	Category    string
	Description string
}

// StoreListResult wraps store list responses with pagination info.
type StoreListResult struct {
	Data       []models.Store `json:"data"`
	Pagination Pagination     `json:"pagination"`
}

// Create creates a new store in pending state.
func (s *StoreService) Create(createdBy uuid.UUID, input CreateStoreInput) (*models.Store, error) {
	name := strings.TrimSpace(input.Name)
	address := strings.TrimSpace(input.Address)
	phone := strings.TrimSpace(input.Phone)
	category := strings.TrimSpace(input.Category)
	description := strings.TrimSpace(input.Description)

	if name == "" || address == "" {
		return nil, errors.New("store name and address are required")
	}

	// Check if store already exists
	existing, err := s.stores.FindByNameAndAddress(name, address)
	if err == nil && existing != nil {
		return nil, errors.New("store already exists")
	}

	store := &models.Store{
		ID:          uuid.New(),
		Name:        name,
		Address:     address,
		Phone:       phone,
		Category:    category,
		Description: description,
		Status:      models.StoreStatusPending,
		CreatedBy:   createdBy,
	}

	if err := s.stores.Create(store); err != nil {
		return nil, err
	}
	return store, nil
}

// ListApproved returns approved stores with pagination.
func (s *StoreService) ListApproved(page, pageSize int, query string) (StoreListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	stores, total, err := s.stores.SearchStores(query, []models.StoreStatus{models.StoreStatusApproved}, pageSize, offset)
	if err != nil {
		return StoreListResult{}, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return StoreListResult{
		Data: stores,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
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
		Data: stores,
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
	if store.Status != models.StoreStatusPending {
		return fmt.Errorf("store already processed")
	}
	store.Status = models.StoreStatusApproved
	store.RejectionReason = ""
	return s.stores.Update(store)
}

// Reject marks a store as rejected with reason.
func (s *StoreService) Reject(store *models.Store, reason string) error {
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

// CreateStoreWithReview creates a new store and review together.
func (s *StoreService) CreateStoreWithReview(ctx context.Context, userID uuid.UUID, storeInput CreateStoreInput, reviewTitle, reviewContent string, rating float32) (*models.Store, *models.Review, error) {
	// Create store
	store, err := s.Create(userID, storeInput)
	if err != nil {
		return nil, nil, err
	}

	// Create review for the store
	review := &models.Review{
		ID:       uuid.New(),
		StoreID:  store.ID,
		AuthorID: userID,
		Title:    reviewTitle,
		Content:  reviewContent,
		Rating:   rating,
		Status:   models.ReviewStatusPending,
	}

	if err := s.db.Create(review).Error; err != nil {
		return nil, nil, err
	}

	return store, review, nil
}

// FindByUserAndStore finds a review by user and store (for one-user-one-store constraint).
func (s *StoreService) FindByUserAndStore(userID, storeID uuid.UUID) (*models.Review, error) {
	var review models.Review
	err := s.db.Where("author_id = ? AND store_id = ?", userID, storeID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// DeleteStore removes a store by ID.
func (s *StoreService) DeleteStore(id uuid.UUID) error {
	return s.stores.Delete(id)
}
