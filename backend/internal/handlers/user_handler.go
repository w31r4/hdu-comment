package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/common/problem"
	"github.com/hdu-dp/backend/internal/repository"
	"github.com/hdu-dp/backend/internal/services"
)

// UserHandler exposes user profile endpoints.
type UserHandler struct {
	users   *repository.UserRepository
	reviews *services.ReviewService
}

// NewUserHandler constructs a UserHandler.
func NewUserHandler(users *repository.UserRepository, reviews *services.ReviewService) *UserHandler {
	return &UserHandler{users: users, reviews: reviews}
}

// @Summary      获取当前用户信息
// @Description  获取当前已认证用户的详细信息。
// @Tags         用户
// @Produce      json
// @Success      200 {object} object{id=integer,email=string,display_name=string,role=string,created_at=string} "用户信息"
// @Failure      401 {object} problem.Details "未认证"
// @Failure      404 {object} problem.Details "用户不存在"
// @Security     ApiKeyAuth
// @Router       /users/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	idVal, exists := c.Get("user_id")
	if !exists {
		problem.Unauthorized("missing user").Send(c)
		return
	}

	userID, ok := idVal.(uuid.UUID)
	if !ok {
		problem.Unauthorized("invalid user id").Send(c)
		return
	}

	user, err := h.users.FindByID(userID)
	if err != nil {
		problem.NotFound("user not found").Send(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"role":         user.Role,
		"created_at":   user.CreatedAt,
	})
}

// @Summary      我的点评列表
// @Description  获取当前认证用户提交的所有点评列表，支持分页、搜索和排序。
// @Tags         用户
// @Produce      json
// @Param        page      query int    false "页码" default(1)
// @Param        limit     query int    false "每页数量" default(10)
// @Param        sort      query string false "排序字段 (e.g., -created_at, rating)" default(-created_at)
// @Success      200 {object} services.ReviewListResult
// @Failure      500 {object} problem.Details "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /users/me/reviews [get]
func (h *UserHandler) MyReviews(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	filters := services.ParseListFilters(c)
	result, err := h.reviews.ListByAuthor(userID, filters)
	if err != nil {
		problem.InternalServerError(err.Error()).Send(c)
		return
	}
	c.JSON(http.StatusOK, result)
}
