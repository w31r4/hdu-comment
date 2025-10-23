package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/common/problem"
	"github.com/hdu-dp/backend/internal/dto"
	"github.com/hdu-dp/backend/internal/services"
)

// StoreAdminHandler manages store admin related HTTP endpoints.
type StoreAdminHandler struct {
	stores *services.StoreService
}

// NewStoreAdminHandler constructs a StoreAdminHandler.
func NewStoreAdminHandler(stores *services.StoreService) *StoreAdminHandler {
	return &StoreAdminHandler{stores: stores}
}

// @Summary      待审核店铺列表
// @Description  获取待审核的店铺列表，支持分页。
// @Tags         管理员 - 店铺
// @Produce      json
// @Param        page      query int    false "页码" default(1)
// @Param        page_size query int    false "每页数量" default(10)
// @Success      200 {object} services.StoreListResult
// @Failure      500 {object} problem.Details "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /admin/stores/pending [get]
func (h *StoreAdminHandler) Pending(c *gin.Context) {
	filters := services.ParseListFilters(c)
	result, err := h.stores.ListPending(filters.Page, filters.PageSize)
	if err != nil {
		problem.InternalServerError(err.Error()).Send(c)
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary      审核通过店铺
// @Description  管理员审核通过指定的店铺。
// @Tags         管理员 - 店铺
// @Accept       json
// @Produce      json
// @Param        id path string true "店铺 ID"
// @Success      200 {object} models.Store "更新后的店铺"
// @Failure      400 {object} problem.Details "店铺已被处理"
// @Failure      404 {object} problem.Details "店铺不存在"
// @Security     ApiKeyAuth
// @Router       /admin/stores/{id}/approve [put]
func (h *StoreAdminHandler) Approve(c *gin.Context) {
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

	if err := h.stores.Approve(store); err != nil {
		problem.BadRequest(err.Error()).Send(c)
		return
	}

	c.JSON(http.StatusOK, store)
}

// @Summary      驳回店铺
// @Description  管理员驳回指定的店铺并填写原因。
// @Tags         管理员 - 店铺
// @Accept       json
// @Produce      json
// @Param        id   path string true "店铺 ID"
// @Param        body body object{reason=string} true "驳回原因"
// @Success      200 {object} models.Store "更新后的店铺"
// @Failure      400 {object} problem.Details "店铺已被处理"
// @Failure      404 {object} problem.Details "店铺不存在"
// @Security     ApiKeyAuth
// @Router       /admin/stores/{id}/reject [put]
func (h *StoreAdminHandler) Reject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problem.BadRequest("invalid store id").Send(c)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.BadRequest("reason is required").Send(c)
		return
	}

	store, err := h.stores.Get(id)
	if err != nil {
		problem.NotFound("store not found").Send(c)
		return
	}

	if err := h.stores.Reject(store, req.Reason); err != nil {
		problem.BadRequest(err.Error()).Send(c)
		return
	}

	c.JSON(http.StatusOK, store)
}

// @Summary      删除店铺
// @Description  管理员删除指定的店铺（包含关联的评价和图片记录）。
// @Tags         管理员 - 店铺
// @Produce      json
// @Param        id path string true "店铺 ID"
// @Success      204 "删除成功"
// @Failure      404 {object} problem.Details "店铺不存在"
// @Security     ApiKeyAuth
// @Router       /admin/stores/{id} [delete]
func (h *StoreAdminHandler) Delete(c *gin.Context) {
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

	if err := h.stores.DeleteStore(store.ID); err != nil {
		problem.InternalServerError(err.Error()).Send(c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      创建新店铺
// @Description  管理员直接创建一个已审核通过的店铺。
// @Tags         管理员 - 店铺
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateStoreRequest true "店铺信息"
// @Success      201 {object} dto.StoreResponse "创建成功"
// @Failure      400 {object} problem.Details "请求参数错误"
// @Failure      409 {object} problem.Details "店铺已存在"
// @Security     ApiKeyAuth
// @Router       /admin/stores [post]
func (h *StoreAdminHandler) CreateStore(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req dto.CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.BadRequest("invalid payload").Send(c)
		return
	}

	store, err := h.stores.CreateStore(c.Request.Context(), userID, true, req)
	if err != nil {
		if err.Error() == "store already exists" {
			problem.Conflict("该店铺已存在").Send(c)
			return
		}
		problem.BadRequest(err.Error()).Send(c)
		return
	}

	c.Header("Location", "/api/v1/stores/"+store.ID.String())
	c.JSON(http.StatusCreated, dto.ToStoreResponse(store))
}
