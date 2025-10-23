package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReviewImage stores uploaded image metadata for a review.
type ReviewImage struct {
	ID         uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	ReviewID   uuid.UUID      `gorm:"type:char(36);index;not null" json:"review_id"`
	StorageKey string         `gorm:"size:255;not null" json:"storage_key"`
	URL        string         `gorm:"size:512;not null" json:"url"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate assigns UUIDs automatically.
func (ri *ReviewImage) BeforeCreate(tx *gorm.DB) error {
	if ri.ID == uuid.Nil {
		ri.ID = uuid.New()
	}
	return nil
}
