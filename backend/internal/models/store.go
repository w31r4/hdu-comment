package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StoreStatus enumerates store workflow states.
type StoreStatus string

const (
	StoreStatusPending  StoreStatus = "pending"
	StoreStatusApproved StoreStatus = "approved"
	StoreStatusRejected StoreStatus = "rejected"
)

// Store represents a food establishment/shop.
type Store struct {
	ID              uuid.UUID   `gorm:"type:char(36);primaryKey" json:"id"`
	Name            string      `gorm:"size:120;not null;index" json:"name"`
	Address         string      `gorm:"size:255;not null" json:"address"`
	Phone           string      `gorm:"size:20" json:"phone"`
	Category        string      `gorm:"size:50;index" json:"category"`
	Description     string      `gorm:"type:text" json:"description"`
	Status          StoreStatus `gorm:"size:20;default:pending" json:"status"`
	AverageRating   float32     `gorm:"type:decimal(2,1);default:0" json:"average_rating"`
	TotalReviews    int         `gorm:"default:0" json:"total_reviews"`
	RejectionReason string      `gorm:"type:text" json:"rejection_reason"`
	CreatedBy       uuid.UUID   `gorm:"type:char(36);not null" json:"created_by"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// BeforeCreate assigns a UUID if empty.
func (s *Store) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
