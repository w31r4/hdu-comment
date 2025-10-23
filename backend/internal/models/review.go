package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Review represents a food review submitted by a user for a specific store.
type Review struct {
	ID              uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	StoreID         uuid.UUID      `gorm:"type:char(36);not null;index:idx_user_store,unique" json:"store_id"`
	Store           Store          `gorm:"foreignKey:StoreID" json:"store"`
	AuthorID        uuid.UUID      `gorm:"type:char(36);not null;index:idx_user_store,unique" json:"author_id"`
	Author          User           `gorm:"foreignKey:AuthorID" json:"author"`
	Title           string         `gorm:"size:120;not null" json:"title"`
	Content         string         `gorm:"type:text;not null" json:"content"`
	Rating          float32        `gorm:"type:decimal(2,1);not null" json:"rating"`
	Status          ReviewStatus   `gorm:"size:20;default:pending" json:"status"`
	RejectionReason string         `gorm:"type:text" json:"rejection_reason"`
	Images          []ReviewImage  `gorm:"foreignKey:ReviewID" json:"images"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate assigns a UUID if empty.
func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// ReviewStatus enumerates review workflow states.
type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
)
