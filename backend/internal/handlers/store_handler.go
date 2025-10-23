package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/common/problem"
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
// @Description  根据多种条件搜索已审核通过的店铺，支持分页、排序和过滤。
// @Tags         店铺
// @Produce      json
// @Param        q        query string false "搜索关键词 (名称或地址)"
// @Param        page     query int    false "页码" default(1)
// @Param        limit    query int    false "每页数量" default(10)
// @Param        sort     query string false "排序字段 (e.g., -created_at, average_rating)"
// @Param        category query string false "店铺分类"
// @Param        status   query string false "店铺状态 (仅管理员可用)"
// @Success      200 {object} services.StoreListResult
// @Failure      500 {object} problem.Details "服务器内部错误"
// @Router       /stores [get]
func (h *StoreHandler) SearchStores(c *gin.Context) {
	// 统一从 ListFilters 解析
	filters := services.ParseListFilters(c)

	// 允许管理员按状态查询
	role, _ := c.Get("role")
	if role != "admin" {
		filters.Status = string(models.StoreStatusApproved)
	}

	result, err := h.stores.ListStores(filters)
	if err != nil {
		problem.InternalServerError(err.Error()).Send(c)
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
// @Failure      400 {object} problem.Details "无效的店铺 ID"
// @Failure      404 {object} problem.Details "店铺不存在"
// @Router       /stores/{id} [get]
func (h *StoreHandler) GetStore(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problem.BadRequest("invalid store id").Send(c)
		return
	}

	store, err := h.stores.Get(id)
	if err != nil {
		problem.NotFound("store not found").Send(c)
		return
	}

	// 只返回已审核的店铺
	if store.Status != models.StoreStatusApproved {
		// 如果是管理员，可以查看
		role, _ := c.Get("role")
		if role != "admin" {
			problem.NotFound("store not found").Send(c)
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
// @Failure      400 {object} problem.Details "请求参数错误"
// @Failure      409 {object} problem.Details "店铺已存在"
// @Security     ApiKeyAuth
// @Router       /stores [post]
func (h *StoreHandler) CreateStore(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req dto.CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.BadRequest("invalid payload").Send(c)
		return
	}

	store, err := h.stores.CreateStore(c.Request.Context(), userID, false, req)
	if err != nil {
		if err.Error() == "store already exists" {
			problem.Conflict("该店铺已存在，请直接评价").Send(c)
			return
		}
		problem.BadRequest(err.Error()).Send(c)
		return
	}

	c.Header("Location", "/api/v1/stores/"+store.ID.String())
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
// @Failure      400 {object} problem.Details "无效的店铺 ID"
// @Failure      500 {object} problem.Details "服务器内部错误"
// @Router       /stores/{id}/reviews [get]
func (h *StoreHandler) GetStoreReviews(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problem.BadRequest("invalid store id").Send(c)
		return
	}

	filters := services.ParseListFilters(c)
	result, err := h.reviews.ListByStore(storeID, filters)
	if err != nil {
		problem.InternalServerError(err.Error()).Send(c)
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary      提交店铺评价
// @Description  为指定店铺提交一条新评价。支持通过 `?autoCreate=true` 在评价时自动创建不存在的店铺。该接口支持通过 `Idempotency-Key` 请求头实现幂等性。
// @Tags         店铺
// @Accept       json
// @Produce      json
// @Param        id              path      string                  true  "店铺 ID (当 autoCreate=false 时)"
// @Param        autoCreate      query     bool                    false "是否在店铺不存在时自动创建"
// @Param        Idempotency-Key header    string                  false "幂等键 (UUID)，用于防止重复提交"
// @Param        body            body      dto.CreateReviewRequest true  "评价内容 (当 autoCreate=false)"
// @Param        body            body      dto.CreateReviewForNewStoreRequest true "评价和新店铺信息 (当 autoCreate=true)"
// @Success      201             {object}  dto.ReviewResponse "创建成功"
// @Failure      400             {object}  problem.Details "请求参数错误"
// @Failure      401             {object}  problem.Details "未认证"
// @Failure      409             {object}  problem.Details "用户已评价过该店铺"
// @Failure      429             {object}  problem.Details "请求正在处理中"
// @Security     ApiKeyAuth
// @Router       /stores/{id}/reviews [post]
func (h *StoreHandler) CreateReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	autoCreate := c.Query("autoCreate") == "true"

	if autoCreate {
		var req dto.CreateReviewForNewStoreRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			problem.BadRequest("invalid payload for auto-create").Send(c)
			return
		}
		store, review, err := h.reviews.CreateReviewForNewStore(c.Request.Context(), userID, req)
		if err != nil {
			problem.BadRequest(err.Error()).Send(c)
			return
		}
		c.Header("Location", "/api/v1/reviews/"+review.ID.String())
		c.JSON(http.StatusCreated, gin.H{
			"store":        dto.ToStoreResponse(store),
			"review":       dto.ToReviewResponse(review),
			"is_new_store": true,
		})
		return
	}

	storeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problem.BadRequest("invalid store id").Send(c)
		return
	}

	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.BadRequest("invalid payload").Send(c)
		return
	}

	review, err := h.reviews.Submit(userID, storeID, req)
	if err != nil {
		if err.Error() == "user has already reviewed this store" {
			problem.Conflict("you have already reviewed this store").Send(c)
			return
		}
		problem.Conflict(err.Error()).Send(c)
		return
	}

	c.Header("Location", "/api/v1/reviews/"+review.ID.String())
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
// @Failure      400        {object}  problem.Details "请求参数错误"
// @Failure      401        {object}  problem.Details "未认证"
// @Failure      403        {object}  problem.Details "无权操作"
// @Failure      404        {object}  problem.Details "评价不存在"
// @Security     ApiKeyAuth
// @Router       /stores/{id}/reviews/{reviewId} [patch]
func (h *StoreHandler) UpdateReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	reviewID, err := uuid.Parse(c.Param("reviewId"))
	if err != nil {
		problem.BadRequest("invalid review id").Send(c)
		return
	}

	var req dto.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.BadRequest("invalid payload").Send(c)
		return
	}

	review, err := h.reviews.Update(c.Request.Context(), userID, reviewID, req)
	if err != nil {
		// TODO: More specific error handling
		problem.BadRequest(err.Error()).Send(c)
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
// @Failure      401      {object}  problem.Details "未认证"
// @Failure      403      {object}  problem.Details "无权操作"
// @Failure      404      {object}  problem.Details "评价不存在"
// @Security     ApiKeyAuth
// @Router       /stores/{id}/reviews/{reviewId} [delete]
func (h *StoreHandler) DeleteReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	reviewID, err := uuid.Parse(c.Param("reviewId"))
	if err != nil {
		problem.BadRequest("invalid review id").Send(c)
		return
	}

	err = h.reviews.Delete(c.Request.Context(), userID, reviewID)
	if err != nil {
		// TODO: More specific error handling
		problem.BadRequest(err.Error()).Send(c)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetSummary is a placeholder for store summary analytics.
func (h *StoreHandler) GetSummary(c *gin.Context) {
	problem.NotImplemented("not implemented").Send(c)
}

// GetTrend is a placeholder for store review trend analytics.
func (h *StoreHandler) GetTrend(c *gin.Context) {
	problem.NotImplemented("not implemented").Send(c)
}

// UploadImage is a placeholder for uploading store images.
func (h *StoreHandler) UploadImage(c *gin.Context) {
	problem.NotImplemented("not implemented").Send(c)
}
