package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/common/problem"
)

// ErrorHandler creates a middleware to handle errors and format them as RFC 7807 Problem Details.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			// This is a simple example. You might want to inspect the error type
			// and return different status codes or details.
			problem.FromError(500, err.Err).Send(c)
		}
	}
}
