package database

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/config"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/utils"
	"gorm.io/gorm"
)

func seedAdmin(db *gorm.DB, cfg *config.Config) error {
	if cfg.Admin.Email == "" || cfg.Admin.Password == "" {
		return nil
	}

	var existing models.User
	err := db.Where("email = ?", cfg.Admin.Email).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("query admin: %w", err)
	}
	if err == nil {
		return nil
	}

	hashed, err := utils.HashPassword(cfg.Admin.Password)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := models.User{
		ID:           uuid.New(),
		Email:        cfg.Admin.Email,
		PasswordHash: hashed,
		DisplayName:  "Administrator",
		Role:         "admin",
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("create admin: %w", err)
	}

	return nil
}
