package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
)

type IdempotencyService struct {
	repo *repository.IdempotencyKeyRepository
	ttl  time.Duration
}

func NewIdempotencyService(repo *repository.IdempotencyKeyRepository, ttl time.Duration) *IdempotencyService {
	return &IdempotencyService{repo: repo, ttl: ttl}
}

func (s *IdempotencyService) GetKey(key string) (*models.IdempotencyKey, error) {
	return s.repo.FindByKey(key)
}

func (s *IdempotencyService) CreateKey(key string, userID uuid.UUID, c *gin.Context) (*models.IdempotencyKey, error) {
	hash, err := s.hashRequestBody(c)
	if err != nil {
		return nil, err
	}

	record := &models.IdempotencyKey{
		Key:         key,
		UserID:      userID,
		RequestHash: hash,
		Status:      models.IdempotencyKeyStatusInProgress,
		ExpiresAt:   time.Now().Add(s.ttl),
	}

	err = s.repo.Create(record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s *IdempotencyService) UpdateKey(record *models.IdempotencyKey, responseCode int, responseBody []byte) error {
	record.Status = models.IdempotencyKeyStatusCompleted
	record.ResponseCode = responseCode
	record.ResponseBody = responseBody
	return s.repo.Update(record)
}

func (s *IdempotencyService) hashRequestBody(c *gin.Context) (string, error) {
	if c.Request.Body == nil {
		return "", nil
	}
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}
	// Restore the body so it can be read again by the handler
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	hasher := sha256.New()
	hasher.Write(bodyBytes)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
