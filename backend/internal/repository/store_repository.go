package repository

import (
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/gorm"
)

// StoreRepository manages persistence for store entities.
type StoreRepository struct {
	db *gorm.DB
}

// NewStoreRepository constructs a store repository.
func NewStoreRepository(db *gorm.DB) *StoreRepository {
	return &StoreRepository{db: db}
}

// Create inserts a new store entry.
func (r *StoreRepository) Create(store *models.Store) error {
	return r.db.Create(store).Error
}

// FindByID returns a store by UUID.
func (r *StoreRepository) FindByID(id uuid.UUID) (*models.Store, error) {
	var store models.Store
	if err := r.db.First(&store, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &store, nil
}

// FindByNameAndAddress returns a store by name and address.
func (r *StoreRepository) FindByNameAndAddress(name, address string) (*models.Store, error) {
	var store models.Store
	if err := r.db.Where("name = ? AND address = ?", name, address).First(&store).Error; err != nil {
		return nil, err
	}
	return &store, nil
}

// ListByStatus returns stores filtered by status with pagination.
func (r *StoreRepository) ListByStatus(statuses []models.StoreStatus, limit, offset int) ([]models.Store, int64, error) {
	var stores []models.Store
	var total int64

	base := r.db.Model(&models.Store{})

	if len(statuses) > 0 {
		base = base.Where("status IN ?", statuses)
	}

	// Count total
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := base.Limit(limit).Offset(offset).Order("created_at DESC")
	if err := query.Find(&stores).Error; err != nil {
		return nil, 0, err
	}

	return stores, total, nil
}

// SearchStores searches stores by name or address.
func (r *StoreRepository) SearchStores(query string, statuses []models.StoreStatus, limit, offset int) ([]models.Store, int64, error) {
	var stores []models.Store
	var total int64

	base := r.db.Model(&models.Store{})

	if len(statuses) > 0 {
		base = base.Where("status IN ?", statuses)
	}

	if query != "" {
		likeQuery := "%" + query + "%"
		base = base.Where("name LIKE ? OR address LIKE ?", likeQuery, likeQuery)
	}

	// Count total
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	queryDB := base.Limit(limit).Offset(offset).Order("created_at DESC")
	if err := queryDB.Find(&stores).Error; err != nil {
		return nil, 0, err
	}

	return stores, total, nil
}

// Update persists changes to a store.
func (r *StoreRepository) Update(store *models.Store) error {
	return r.db.Save(store).Error
}

// Delete removes a store by ID.
func (r *StoreRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Store{}, "id = ?", id).Error
}

// UpdateAverageRating updates the average rating for a store.
func (r *StoreRepository) UpdateAverageRating(storeID uuid.UUID, averageRating float32, totalReviews int) error {
	return r.db.Model(&models.Store{}).
		Where("id = ?", storeID).
		Updates(map[string]interface{}{
			"average_rating": averageRating,
			"total_reviews":  totalReviews,
		}).Error
}
