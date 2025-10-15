package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshToken represents a persistent refresh token for a user session.
type RefreshToken struct {
	ID         uuid.UUID `gorm:"type:char(36);primaryKey"`
	UserID     uuid.UUID `gorm:"type:char(36);index;not null"`
	SecretHash string    `gorm:"size:255;not null"`
	ExpiresAt  time.Time `gorm:"index"`
	Revoked    bool      `gorm:"default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// BeforeCreate assigns UUIDs automatically.
func (t *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
