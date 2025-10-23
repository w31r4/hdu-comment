package middleware

import (
	"bytes"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/common/problem"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/services"
	"gorm.io/gorm"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Idempotency(service *services.IdempotencyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			problem.Unauthorized("user not authenticated for idempotent request").Send(c)
			c.Abort()
			return
		}

		record, err := service.GetKey(key)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			problem.InternalServerError("failed to check idempotency key").Send(c)
			c.Abort()
			return
		}

		if record != nil {
			if record.Status == models.IdempotencyKeyStatusInProgress {
				problem.TooManyRequests("request with this key is already in progress").Send(c)
				c.Abort()
				return
			}
			if record.Status == models.IdempotencyKeyStatusCompleted {
				c.Data(record.ResponseCode, "application/json", record.ResponseBody)
				c.Abort()
				return
			}
		}

		// No record found, create one
		record, err = service.CreateKey(key, userID.(uuid.UUID), c)
		if err != nil {
			problem.InternalServerError("failed to create idempotency key").Send(c)
			c.Abort()
			return
		}

		// Replace the response writer to capture the response
		rbw := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = rbw

		c.Next()

		// After handler execution, update the key with the response
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			err := service.UpdateKey(record, c.Writer.Status(), rbw.body.Bytes())
			if err != nil {
				// Log the error, but don't fail the request
				// The client will just have to retry
				c.Error(err)
			}
		}
	}
}
