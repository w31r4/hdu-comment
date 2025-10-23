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
	"github.com/hdu-dp/backend/internal/storage"
	"gorm.io/gorm"
)

// StoreService contains business logic around store workflows.
type StoreService struct {
	stores  *repository.StoreRepository
	reviews *repository.ReviewRepository
	storage storage.FileStorage
	db      *gorm.DB
}

// NewStoreService constructs a store service instance.
func NewStoreService(stores *repository.StoreRepository, reviews *repository.ReviewRepository, storage storage.FileStorage, db *gorm.DB) *StoreService {
	return &StoreService{stores: stores, reviews: reviews, storage: storage, db: db}
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

	if err := s.stores.Create(nil, store); err != nil {
		return nil, err
	}
	return store, nil
}

// ListStores returns stores based on filters.
func (s *StoreService) ListStores(filters ListFilters) (StoreListResult, error) {
	offset := (filters.Page - 1) * filters.Limit

	// Convert status string to model type
	var statuses []models.StoreStatus
	if filters.Status != "" {
		statuses = []models.StoreStatus{models.StoreStatus(filters.Status)}
	}

	stores, total, err := s.stores.SearchStores(repository.StoreSearchFilters{
		Query:    filters.Query,
		Statuses: statuses,
		Category: filters.Category,
		Sort:     filters.Sort,
		Limit:    filters.Limit,
		Offset:   offset,
	})
	if err != nil {
		return StoreListResult{}, err
	}

	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))

	return StoreListResult{
		Data: dto.ToStoreListResponse(stores),
		Pagination: Pagination{
			Page:       filters.Page,
			PageSize:   filters.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// ListPending returns pending stores for admin review, using the standard list filters.
func (s *StoreService) ListPending(filters ListFilters) (StoreListResult, error) {
	offset := (filters.Page - 1) * filters.Limit

	// Hardcode the status to pending for this specific list function
	statuses := []models.StoreStatus{models.StoreStatusPending}

	stores, total, err := s.stores.SearchStores(repository.StoreSearchFilters{
		// Query, Category, etc., are not used here but we could support them in the future.
		Statuses: statuses,
		Sort:     filters.Sort,
		Limit:    filters.Limit,
		Offset:   offset,
	})
	if err != nil {
		return StoreListResult{}, err
	}

	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))

	return StoreListResult{
		Data: dto.ToStoreListResponse(stores),
		Pagination: Pagination{
			Page:       filters.Page,
			PageSize:   filters.Limit,
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

// DeleteStore soft-deletes a store and all its associated reviews and images.
// It also hard-deletes the image files from the underlying storage.
func (s *StoreService) DeleteStore(ctx context.Context, id uuid.UUID) error {
	// 1. Find all review images that need to be deleted from storage.
	var imagesToDelete []models.ReviewImage
	err := s.db.Model(&models.ReviewImage{}).
		Joins("JOIN reviews ON reviews.id = review_images.review_id").
		Where("reviews.store_id = ?", id).
		Find(&imagesToDelete).Error
	if err != nil {
		return fmt.Errorf("failed to find images for deletion: %w", err)
	}

	// 2. Perform all database soft deletes in a single transaction.
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Get review IDs for the store
		var reviewIDs []uuid.UUID
		if err := tx.Model(&models.Review{}).Where("store_id = ?", id).Pluck("id", &reviewIDs).Error; err != nil {
			return err
		}

		if len(reviewIDs) > 0 {
			// Soft delete review images for those reviews
			if err := tx.Where("review_id IN ?", reviewIDs).Delete(&models.ReviewImage{}).Error; err != nil {
				return err
			}

			// Soft delete reviews
			if err := tx.Where("id IN ?", reviewIDs).Delete(&models.Review{}).Error; err != nil {
				return err
			}
		}

		// Soft delete the store
		if err := tx.Delete(&models.Store{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to soft-delete store and associated data: %w", err)
	}

	// 3. After the transaction is successful, delete the files from storage.
	for _, image := range imagesToDelete {
		if image.StorageKey != "" {
			if err := s.storage.Delete(ctx, image.StorageKey); err != nil {
				// Log the error but don't fail the whole operation.
				// In a real app, you'd use a structured logger.
				fmt.Printf("warning: failed to delete image %s from storage: %v\n", image.StorageKey, err)
			}
		}
	}

	return nil
}
