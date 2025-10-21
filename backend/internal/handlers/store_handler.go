package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/services"
)

// StoreHandler manages store related HTTP endpoints for users.
type StoreHandler struct {
	stores *services.StoreService
}

// NewStoreHandler constructs a StoreHandler.
func NewStoreHandler(stores *services.StoreService) *StoreHandler {
	return &StoreHandler{stores: stores}
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
// @Description  根据 ID 获取单个店铺的详细信息，包含综合评分和评价列表。
// @Tags         店铺
// @Produce      json
// @Param        id path string true "店铺 ID"
// @Success      200 {object} models.Store "店铺详情"
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
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}

	c.JSON(http.StatusOK, store)
}

// @Summary      获取用户对店铺的评价
// @Description  获取当前用户对指定店铺的评价（如果有）。
// @Tags         店铺
// @Produce      json
// @Param        id path string true "店铺 ID"
// @Success      200 {object} models.Review "用户的评价"
// @Success      204 "用户没有对该店铺的评价"
// @Failure      400 {object} object{error=string} "无效的店铺 ID"
// @Failure      401 {object} object{error=string} "未认证"
// @Security     ApiKeyAuth
// @Router       /stores/{id}/my-review [get]
func (h *StoreHandler) GetMyReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	storeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	review, err := h.stores.FindByUserAndStore(userID, storeID)
	if err != nil {
		// 如果没有找到评价，返回 204 No Content
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, review)
}

// CreateStoreWithReviewInput bundles parameters for creating a store with review.
type CreateStoreWithReviewInput struct {
	// 店铺信息
	StoreName     string `json:"store_name" binding:"required"`
	StoreAddress  string `json:"store_address" binding:"required"`
	StorePhone    string `json:"store_phone"`
	StoreCategory string `json:"store_category"`
	StoreDesc     string `json:"store_description"`
	// 评价信息
	ReviewTitle   string  `json:"review_title" binding:"required"`
	ReviewContent string  `json:"review_content" binding:"required"`
	Rating        float32 `json:"rating" binding:"required,min=0,max=5"`
}

// @Summary      创建新店铺并提交评价
// @Description  用户创建新店铺并同时提交对该店铺的评价，两者都需要审核。
// @Tags         店铺
// @Accept       json
// @Produce      json
// @Param        body body CreateStoreWithReviewInput true "店铺和评价信息"
// @Success      201 {object} object{store=models.Store,review=models.Review} "创建成功"
// @Failure      400 {object} object{error=string} "请求参数错误"
// @Failure      409 {object} object{error=string} "店铺已存在"
// @Security     ApiKeyAuth
// @Router       /stores/with-review [post]
func (h *StoreHandler) CreateStoreWithReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var input CreateStoreWithReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// 创建店铺输入
	storeInput := services.CreateStoreInput{
		Name:        input.StoreName,
		Address:     input.StoreAddress,
		Phone:       input.StorePhone,
		Category:    input.StoreCategory,
		Description: input.StoreDesc,
	}

	store, review, err := h.stores.CreateStoreWithReview(c.Request.Context(), userID, storeInput, input.ReviewTitle, input.ReviewContent, input.Rating)
	if err != nil {
		if err.Error() == "store already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "该店铺已存在，请直接评价"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"store":  store,
		"review": review,
	})
}
