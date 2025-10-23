package problem

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Details represents a standard RFC 7807 problem details response.
type Details struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// New creates a new problem details response.
func New(status int, title string) *Details {
	return &Details{
		Type:   "about:blank",
		Title:  title,
		Status: status,
	}
}

// WithType sets the type URI for the problem.
func (p *Details) WithType(problemType string) *Details {
	p.Type = problemType
	return p
}

// WithDetail adds a human-readable explanation.
func (p *Details) WithDetail(detail string) *Details {
	p.Detail = detail
	return p
}

// WithInstance sets the specific occurrence of the problem.
func (p *Details) WithInstance(instance string) *Details {
	p.Instance = instance
	return p
}

// Send writes the problem details to the Gin context.
func (p *Details) Send(c *gin.Context) {
	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(p.Status, p)
}

// FromError creates a problem details response from a standard error.
func FromError(status int, err error) *Details {
	return New(status, http.StatusText(status)).WithDetail(err.Error())
}

// BadRequest creates a 400 Bad Request problem.
func BadRequest(detail string) *Details {
	return New(http.StatusBadRequest, "Bad Request").WithDetail(detail)
}

// NotFound creates a 404 Not Found problem.
func NotFound(detail string) *Details {
	return New(http.StatusNotFound, "Not Found").WithDetail(detail)
}

// InternalServerError creates a 500 Internal Server Error problem.
func InternalServerError(detail string) *Details {
	return New(http.StatusInternalServerError, "Internal Server Error").WithDetail(detail)
}

// Conflict creates a 409 Conflict problem.
func Conflict(detail string) *Details {
	return New(http.StatusConflict, "Conflict").WithDetail(detail)
}

// Unauthorized creates a 401 Unauthorized problem.
func Unauthorized(detail string) *Details {
	return New(http.StatusUnauthorized, "Unauthorized").WithDetail(detail)
}

// Forbidden creates a 403 Forbidden problem.
func Forbidden(detail string) *Details {
	return New(http.StatusForbidden, "Forbidden").WithDetail(detail)
}

// NotImplemented creates a 501 Not Implemented problem.
func NotImplemented(detail string) *Details {
	return New(http.StatusNotImplemented, "Not Implemented").WithDetail(detail)
}
