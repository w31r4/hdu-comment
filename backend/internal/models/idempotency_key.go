package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IdempotencyKeyStatus string

const (
	IdempotencyKeyStatusInProgress IdempotencyKeyStatus = "in_progress"
	IdempotencyKeyStatusCompleted  IdempotencyKeyStatus = "completed"
)

// IdempotencyKey stores the result of an idempotent request.
type IdempotencyKey struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey"`
	UserID       uuid.UUID `gorm:"type:char(36);not null;index"`
	Key          string    `gorm:"size:50;not null;uniqueIndex"`
	RequestHash  string    `gorm:"size:64;not null"` // SHA-256 hash of the request body
	ResponseCode int
	ResponseBody []byte
	Status       IdempotencyKeyStatus `gorm:"size:20;not null"`
	ExpiresAt    time.Time            `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// BeforeCreate assigns a UUID if empty.
func (i *IdempotencyKey) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
