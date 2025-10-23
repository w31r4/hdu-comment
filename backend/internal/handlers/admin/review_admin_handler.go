package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/common/problem"
	"github.com/hdu-dp/backend/internal/services"
)

// ReviewAdminHandler contains endpoints reserved for administrators.
type ReviewAdminHandler struct {
	reviews *services.ReviewService
	stores  *services.StoreService
}

// NewReviewAdminHandler constructs a new handler.
func NewReviewAdminHandler(reviews *services.ReviewService, stores *services.StoreService) *ReviewAdminHandler {
	return &ReviewAdminHandler{reviews: reviews, stores: stores}
}

// @Summary      待审核点评列表
// @Description  获取等待管理员审核的点评列表，支持分页、搜索和排序。
// @Tags         管理
// @Produce      json
// @Param        page      query int    false "页码" default(1)
// @Param        page_size query int    false "每页数量" default(10)
// @Param        query     query string false "搜索关键词"
// @Param        sort      query string false "排序字段 (e.g., -created_at, rating)" default(-created_at)
// @Success      200 {object} services.ReviewListResult
// @Failure      500 {object} problem.Details "服务器内部错误"
// @Security     BearerAuth
// @Router       /admin/reviews/pending [get]
func (h *ReviewAdminHandler) Pending(c *gin.Context) {
	filters := services.ParseListFilters(c)
	result, err := h.reviews.ListPending(filters)
	if err != nil {
		problem.InternalServerError(err.Error()).Send(c)
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary      更新点评状态
// @Description  批准或拒绝一个点评。
// @Tags         管理
// @Accept       json
// @Produce      json
// @Param        id   path      string true "点评 ID"
// @Param        body body      object{status=string,reason=string} true "状态更新请求 (status: 'approved' 或 'rejected')"
// @Success      200  {object}  models.Review "更新成功"
// @Failure      400  {object}  problem.Details "无效的请求"
// @Failure      404  {object}  problem.Details "点评不存在"
// @Failure      409  {object}  problem.Details "点评状态已被处理，无法再次修改"
// @Security     BearerAuth
// @Router       /admin/reviews/{id}/status [put]
func (h *ReviewAdminHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problem.BadRequest("invalid review id").Send(c)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=approved rejected"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.BadRequest("invalid status or payload: " + err.Error()).Send(c)
		return
	}

	review, err := h.reviews.Get(id)
	if err != nil {
		problem.NotFound("review not found").Send(c)
		return
	}

	switch req.Status {
	case "approved":
		if err := h.reviews.Approve(review); err != nil {
			problem.Conflict(err.Error()).Send(c)
			return
		}
		// 触发店铺平均分更新
		if err := h.stores.UpdateStoreRating(c.Request.Context(), review.StoreID); err != nil {
			// 即使更新失败，也只记录日志，不阻塞主流程
			// log.Printf("failed to update store rating for store %s: %v", review.StoreID, err)
		}
	case "rejected":
		if err := h.reviews.Reject(review, req.Reason); err != nil {
			problem.Conflict(err.Error()).Send(c)
			return
		}
	}

	c.JSON(http.StatusOK, review)
}

// @Summary      删除点评
// @Description  永久删除指定 ID 的点评及其关联的图片。
// @Tags         管理
// @Produce      json
// @Param        id path string true "点评 ID"
// @Success      204 "删除成功"
// @Failure      400 {object} problem.Details "无效的点评 ID"
// @Failure      404 {object} problem.Details "点评不存在"
// @Failure      500 {object} problem.Details "服务器内部错误"
// @Security     BearerAuth
// @Router       /admin/reviews/{id} [delete]
func (h *ReviewAdminHandler) Delete(c *gin.Context) {
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

	// The Delete service now handles ownership check
	if err := h.reviews.Delete(c.Request.Context(), review.AuthorID, review.ID); err != nil {
		problem.InternalServerError(err.Error()).Send(c)
		return
	}

	c.Status(http.StatusNoContent)
}
