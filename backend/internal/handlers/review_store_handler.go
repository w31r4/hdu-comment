package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/services"
)

// ReviewStoreHandler manages review operations related to stores.
type ReviewStoreHandler struct {
	stores  *services.StoreService
	reviews *services.ReviewService
}

// NewReviewStoreHandler constructs a ReviewStoreHandler.
func NewReviewStoreHandler(stores *services.StoreService, reviews *services.ReviewService) *ReviewStoreHandler {
	return &ReviewStoreHandler{stores: stores, reviews: reviews}
}

// SubmitReviewInput bundles parameters for submitting a review to a store.
type SubmitReviewInput struct {
	StoreID string  `json:"store_id" binding:"required"`
	Title   string  `json:"title" binding:"required"`
	Content string  `json:"content" binding:"required"`
	Rating  float32 `json:"rating" binding:"required,min=0,max=5"`
}

// @Summary      提交店铺评价
// @Description  用户对指定店铺提交评价，检查是否已有评价。
// @Tags         评价 - 店铺
// @Accept       json
// @Produce      json
// @Param        body body SubmitReviewInput true "评价信息"
// @Success      201 {object} models.Review "创建成功"
// @Failure      400 {object} object{error=string} "请求参数错误"
// @Failure      409 {object} object{error=string} "已存在评价"
// @Security     ApiKeyAuth
// @Router       /reviews/store [post]
func (h *ReviewStoreHandler) SubmitReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var input SubmitReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	storeID, err := uuid.Parse(input.StoreID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	// 检查用户是否已有该店铺的评价
	existing, err := h.stores.FindByUserAndStore(userID, storeID)
	if err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "您已经对该店铺有过评价，请更新现有评价"})
		return
	}

	// 检查店铺是否存在且已审核通过
	store, err := h.stores.Get(storeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "店铺不存在"})
		return
	}
	if store.Status != models.StoreStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该店铺尚未通过审核"})
		return
	}

	// 创建评价输入
	reviewInput := services.CreateReviewInput{
		Title:       input.Title,
		Description: input.Content,
		Rating:      input.Rating,
	}

	review, err := h.reviews.Submit(userID, storeID, reviewInput)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// UpdateReviewInput bundles parameters for updating a review.
type UpdateReviewInput struct {
	Title   string  `json:"title" binding:"required"`
	Content string  `json:"content" binding:"required"`
	Rating  float32 `json:"rating" binding:"required,min=0,max=5"`
}

// @Summary      更新店铺评价
// @Description  用户更新自己对指定店铺的现有评价。
// @Tags         评价 - 店铺
// @Accept       json
// @Produce      json
// @Param        id   path string true "评价 ID"
// @Param        body body UpdateReviewInput true "更新内容"
// @Success      200 {object} models.Review "更新后的评价"
// @Failure      400 {object} object{error=string} "请求参数错误"
// @Failure      403 {object} object{error=string} "无权操作"
// @Failure      404 {object} object{error=string} "评价不存在"
// @Security     ApiKeyAuth
// @Router       /reviews/store/{id} [put]
func (h *ReviewStoreHandler) UpdateReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	var input UpdateReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// 获取现有评价
	review, err := h.reviews.Get(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	// 验证所有权
	if err := services.ValidateOwnership(review, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not owner"})
		return
	}

	// 更新评价内容
	review.Title = input.Title
	review.Content = input.Content
	review.Rating = input.Rating
	// 更新后需要重新审核
	review.Status = models.ReviewStatusPending
	review.RejectionReason = ""

	if err := h.reviews.Update(review); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, review)
}

// @Summary      获取店铺的所有评价
// @Description  获取指定店铺的所有已审核通过的评价，支持分页。
// @Tags         评价 - 店铺
// @Produce      json
// @Param        id        path string true "店铺 ID"
// @Param        page      query int    false "页码" default(1)
// @Param        page_size query int    false "每页数量" default(10)
// @Success      200 {object} services.ReviewListResult
// @Failure      400 {object} object{error=string} "无效的店铺 ID"
// @Failure      404 {object} object{error=string} "店铺不存在"
// @Router       /stores/{id}/reviews [get]
func (h *ReviewStoreHandler) GetStoreReviews(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	// 检查店铺是否存在且已审核通过
	store, err := h.stores.Get(storeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}
	if store.Status != models.StoreStatusApproved {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 获取该店铺的所有已审核评价
	filters := services.ListFilters{
		Page:     page,
		PageSize: pageSize,
		SortBy:   "created_at",
		SortDir:  "desc",
		StoreID:  &storeID,
	}

	result, err := h.reviews.ListPublic(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
