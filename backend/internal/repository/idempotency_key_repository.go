package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/gorm"
)

type IdempotencyKeyRepository struct {
	db *gorm.DB
}

func NewIdempotencyKeyRepository(db *gorm.DB) *IdempotencyKeyRepository {
	return &IdempotencyKeyRepository{db: db}
}

func (r *IdempotencyKeyRepository) FindByKey(key string) (*models.IdempotencyKey, error) {
	var record models.IdempotencyKey
	err := r.db.Where("key = ? AND expires_at > ?", key, time.Now()).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *IdempotencyKeyRepository) Create(record *models.IdempotencyKey) error {
	return r.db.Create(record).Error
}

func (r *IdempotencyKeyRepository) Update(record *models.IdempotencyKey) error {
	return r.db.Save(record).Error
}

func (r *IdempotencyKeyRepository) FindByUserIDAndHash(userID uuid.UUID, hash string) (*models.IdempotencyKey, error) {
	var record models.IdempotencyKey
	err := r.db.Where("user_id = ? AND request_hash = ? AND expires_at > ?", userID, hash, time.Now()).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}
