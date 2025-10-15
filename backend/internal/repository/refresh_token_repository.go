package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/gorm"
)

// RefreshTokenRepository manages refresh token persistence.
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository constructs repository instance.
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create inserts a refresh token record.
func (r *RefreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

// FindByID retrieves a refresh token by primary key.
func (r *RefreshTokenRepository) FindByID(id uuid.UUID) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.First(&token, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// Save persists modifications to a refresh token.
func (r *RefreshTokenRepository) Save(token *models.RefreshToken) error {
	return r.db.Save(token).Error
}

// DeleteExpired removes expired tokens.
func (r *RefreshTokenRepository) DeleteExpired(now time.Time) error {
	return r.db.Where("expires_at < ?", now).Delete(&models.RefreshToken{}).Error
}

// RevokeAllForUser revokes tokens for a user.
func (r *RefreshTokenRepository) RevokeAllForUser(userID uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).
		Error
}

// DeleteByID removes a token.
func (r *RefreshTokenRepository) DeleteByID(id uuid.UUID) error {
	return r.db.Delete(&models.RefreshToken{}, "id = ?", id).Error
}
