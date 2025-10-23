package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/common/problem"
	"github.com/hdu-dp/backend/internal/dto"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/services"
	"github.com/hdu-dp/backend/internal/storage"
)

// ReviewHandler manages review related HTTP endpoints.
type ReviewHandler struct {
	reviews *services.ReviewService
}

// NewReviewHandler constructs a ReviewHandler.
func NewReviewHandler(reviews *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviews: reviews}
}

// @Summary      公开点评列表
// @Description  获取已审核通过的点评列表，支持分页、搜索和排序。
// @Tags         点评
// @Produce      json
// @Param        page      query int    false "页码" default(1)
// @Param        page_size query int    false "每页数量" default(10)
// @Param        query     query string false "搜索关键词"
// @Param        sort      query string false "排序字段 (e.g., -created_at, rating)" default(-created_at)
// @Success      200 {object} services.ReviewListResult
// @Failure      500 {object} problem.Details "服务器内部错误"
// @Router       /reviews [get]
func (h *ReviewHandler) ListPublic(c *gin.Context) {
	filters := services.ParseListFilters(c)
	result, err := h.reviews.ListPublic(filters)
	if err != nil {
		problem.InternalServerError(err.Error()).Send(c)
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary      获取点评详情
// @Description  根据 ID 获取单个点评的详细信息。未审核的点评仅作者和管理员可见。
// @Tags         点评
// @Produce      json
// @Param        id path string true "点评 ID"
// @Success      200 {object} dto.ReviewResponse
// @Failure      400 {object} problem.Details "无效的点评 ID"
// @Failure      403 {object} problem.Details "无权访问"
// @Failure      404 {object} problem.Details "点评不存在"
// @Router       /reviews/{id} [get]
func (h *ReviewHandler) Detail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problem.BadRequest("invalid review id").Send(c)
		return
	}

	review, err := h.reviews.Get(id)
	if err != nil {
		problem.NotFound("review not found").Send(c)
		return
	}

	if review.Status != models.ReviewStatusApproved {
		roleVal, ok := c.Get("role")
		role := ""
		if ok {
			role, _ = roleVal.(string)
		}
		if role != "admin" {
			userVal, ok := c.Get("user_id")
			userID, okID := userVal.(uuid.UUID)
			if !ok || !okID || review.AuthorID != userID {
				problem.Forbidden("review not accessible").Send(c)
				return
			}
		}
	}

	c.JSON(http.StatusOK, dto.ToReviewResponse(review))
}

// @Summary      上传点评图片
// @Description  为指定的点评上传一张图片。用户只能为自己的点评上传。
// @Tags         点评
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path      string true "点评 ID"
// @Param        file formData  file   true "图片文件"
// @Success      201  {object}  models.ReviewImage "上传成功"
// @Failure      400  {object}  problem.Details "请求错误"
// @Failure      403  {object}  problem.Details "无权操作"
// @Failure      404  {object}  problem.Details "点评不存在"
// @Failure      500  {object}  problem.Details "服务器内部错误"
// @Security     BearerAuth
// @Router       /reviews/{id}/images [post]
func (h *ReviewHandler) UploadImage(c *gin.Context) {
	reviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problem.BadRequest("invalid review id").Send(c)
		return
	}

	// Service layer will validate ownership
	userID := c.MustGet("user_id").(uuid.UUID)

	review, err := h.reviews.Get(reviewID)
	if err != nil {
		problem.NotFound("review not found").Send(c)
		return
	}
	if review.AuthorID != userID {
		problem.Forbidden("not owner").Send(c)
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		problem.BadRequest("file is required").Send(c)
		return
	}

	opened, err := fileHeader.Open()
	if err != nil {
		problem.InternalServerError(err.Error()).Send(c)
		return
	}

	uploadFile := &storage.UploadFile{
		Reader:      opened,
		Size:        fileHeader.Size,
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
	}

	image, err := h.reviews.StoreImage(c.Request.Context(), reviewID, uploadFile)
	if err != nil {
		problem.InternalServerError(err.Error()).Send(c)
		return
	}

	c.JSON(http.StatusCreated, image)
}

// @Summary      创建新评价（可自动创建店铺）
// @Description  创建一个新的评价。如果店铺不存在，则根据提供的店铺信息自动创建。此接口为幂等操作。
// @Tags         点评
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key header    string                  false "幂等键 (UUID)，用于防止重复提交"
// @Param        body            body      dto.CreateReviewForNewStoreRequest true "评价和新店铺信息"
// @Success      201             {object}  dto.AutoCreateReviewResponse "创建成功"
// @Failure      400             {object}  problem.Details "请求参数错误"
// @Failure      401             {object}  problem.Details "未认证"
// @Failure      409             {object}  problem.Details "用户已评价过该店铺或请求正在处理中"
// @Failure      429             {object}  problem.Details "请求正在处理中"
// @Security     BearerAuth
// @Router       /reviews [post]
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req dto.CreateReviewForNewStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.BadRequest("invalid payload for auto-create").Send(c)
		return
	}

	store, review, err := h.reviews.CreateReviewForNewStore(c.Request.Context(), userID, req)
	if err != nil {
		if err.Error() == "user has already reviewed this store" {
			problem.Conflict("you have already reviewed this store").Send(c)
			return
		}
		problem.BadRequest(err.Error()).Send(c)
		return
	}

	c.Header("Location", "/api/v1/reviews/"+review.ID.String())
	c.JSON(http.StatusCreated, dto.AutoCreateReviewResponse{
		Store:      dto.ToStoreResponse(store),
		Review:     dto.ToReviewResponse(review),
		IsNewStore: true, // This logic might need refinement if the store could exist
	})
}
