package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/dto"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/services"
)

// StoreHandler manages store related HTTP endpoints for users.
type StoreHandler struct {
	stores  *services.StoreService
	reviews *services.ReviewService
}

// NewStoreHandler constructs a StoreHandler.
func NewStoreHandler(stores *services.StoreService, reviews *services.ReviewService) *StoreHandler {
	return &StoreHandler{stores: stores, reviews: reviews}
}

// @Summary      搜索店铺
// @Description  根据名称或地址搜索已审核通过的店铺，支持分页。
// @Tags         店铺
// @Produce      json
// @Param        query     query string false "搜索关键词"
// @Param        page      query int    false "页码" default(1)
// @Param        page_size query int    false "每页数量" default(10)
// @Success      200 {object} services.StoreListResult
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Router       /stores [get]
func (h *StoreHandler) SearchStores(c *gin.Context) {
	query := c.Query("query")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.stores.ListApproved(page, pageSize, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary      获取店铺详情
// @Description  根据 ID 获取单个店铺的详细信息。
// @Tags         店铺
// @Produce      json
// @Param        id path string true "店铺 ID"
// @Success      200 {object} dto.StoreResponse "店铺详情"
// @Failure      400 {object} object{error=string} "无效的店铺 ID"
// @Failure      404 {object} object{error=string} "店铺不存在"
// @Router       /stores/{id} [get]
func (h *StoreHandler) GetStore(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	store, err := h.stores.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}

	// 只返回已审核的店铺
	if store.Status != models.StoreStatusApproved {
		// 如果是管理员，可以查看
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
			return
		}
	}

	c.JSON(http.StatusOK, dto.ToStoreResponse(store))
}

// @Summary      创建新店铺
// @Description  用户创建一个新的店铺，需要等待管理员审核。
// @Tags         店铺
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateStoreRequest true "店铺信息"
// @Success      201 {object} dto.StoreResponse "创建成功"
// @Failure      400 {object} object{error=string} "请求参数错误"
// @Failure      409 {object} object{error=string} "店铺已存在"
// @Security     ApiKeyAuth
// @Router       /stores [post]
func (h *StoreHandler) CreateStore(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req dto.CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	store, err := h.stores.CreateStore(c.Request.Context(), userID, false, req)
	if err != nil {
		if err.Error() == "store already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "该店铺已存在，请直接评价"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToStoreResponse(store))
}

// @Summary      获取店铺的评价列表
// @Description  获取指定店铺下所有已审核通过的评价。
// @Tags         店铺
// @Produce      json
// @Param        id        path      string true "店铺 ID"
// @Param        page      query     int    false "页码" default(1)
// @Param        page_size query     int    false "每页数量" default(10)
// @Success      200 {object} services.ReviewListResult
// @Failure      400 {object} object{error=string} "无效的店铺 ID"
// @Failure      500 {object} object{error=string} "服务器内部错误"
// @Router       /stores/{id}/reviews [get]
func (h *StoreHandler) GetStoreReviews(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	filters := parseListFilters(c)
	result, err := h.reviews.ListByStore(storeID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary      提交店铺评价
// @Description  为指定店铺提交一条新评价。
// @Tags         店铺
// @Accept       json
// @Produce      json
// @Param        id   path      string               true "店铺 ID"
// @Param        body body      dto.CreateReviewRequest true "评价内容"
// @Success      201  {object}  dto.ReviewResponse "创建成功"
// @Failure      400  {object}  object{error=string} "请求参数错误"
// @Failure      401  {object}  object{error=string} "未认证"
// @Failure      409  {object}  object{error=string} "用户已评价过该店铺"
// @Security     ApiKeyAuth
// @Router       /stores/{id}/reviews [post]
func (h *StoreHandler) CreateReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	autoCreate := c.Query("autoCreate") == "true"

	if autoCreate {
		var req dto.CreateReviewForNewStoreRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload for auto-create"})
			return
		}
		store, review, err := h.reviews.CreateReviewForNewStore(c.Request.Context(), userID, req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"store":        dto.ToStoreResponse(store),
			"review":       dto.ToReviewResponse(review),
			"is_new_store": true,
		})
		return
	}

	storeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	review, err := h.reviews.Submit(userID, storeID, req)
	if err != nil {
		if err.Error() == "user has already reviewed this store" {
			c.JSON(http.StatusConflict, gin.H{"error": "you have already reviewed this store"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"review":       dto.ToReviewResponse(review),
		"is_new_store": false,
	})
}

// @Summary      更新店铺评价
// @Description  更新用户对指定店铺的评价。
// @Tags         店铺
// @Accept       json
// @Produce      json
// @Param        id         path      string                  true "店铺 ID"
// @Param        reviewId   path      string                  true "评价 ID"
// @Param        body       body      dto.UpdateReviewRequest true "要更新的评价内容"
// @Success      200        {object}  dto.ReviewResponse "更新成功"
// @Failure      400        {object}  object{error=string} "请求参数错误"
// @Failure      401        {object}  object{error=string} "未认证"
// @Failure      403        {object}  object{error=string} "无权操作"
// @Failure      404        {object}  object{error=string} "评价不存在"
// @Security     ApiKeyAuth
// @Router       /stores/{id}/reviews/{reviewId} [put]
func (h *StoreHandler) UpdateReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	reviewID, err := uuid.Parse(c.Param("reviewId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	var req dto.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	review, err := h.reviews.Update(c.Request.Context(), userID, reviewID, req)
	if err != nil {
		// TODO: More specific error handling
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToReviewResponse(review))
}

// @Summary      删除店铺评价
// @Description  删除用户对指定店铺的评价。
// @Tags         店铺
// @Produce      json
// @Param        id       path      string true "店铺 ID"
// @Param        reviewId path      string true "评价 ID"
// @Success      204      "删除成功"
// @Failure      401      {object}  object{error=string} "未认证"
// @Failure      403      {object}  object{error=string} "无权操作"
// @Failure      404      {object}  object{error=string} "评价不存在"
// @Security     ApiKeyAuth
// @Router       /stores/{id}/reviews/{reviewId} [delete]
func (h *StoreHandler) DeleteReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	reviewID, err := uuid.Parse(c.Param("reviewId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	err = h.reviews.Delete(c.Request.Context(), userID, reviewID)
	if err != nil {
		// TODO: More specific error handling
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
