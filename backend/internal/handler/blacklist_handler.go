package handler

import (
	"net/http"
	"strconv"

	"waf-go/internal/service"
	"waf-go/internal/utils"

	"github.com/gin-gonic/gin"
)

type BlackListHandler struct {
	blackListService *service.BlackListService
}

func NewBlackListHandler(blackListService *service.BlackListService) *BlackListHandler {
	return &BlackListHandler{
		blackListService: blackListService,
	}
}

// CreateBlackList 创建黑名单
// @Summary 创建黑名单
// @Description 创建新的黑名单条目
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Param blacklist body service.CreateBlackListRequest true "黑名单信息"
// @Success 200 {object} utils.Response{data=models.BlackList}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/blacklists [post]
func (h *BlackListHandler) CreateBlackList(c *gin.Context) {
	var req service.CreateBlackListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 从JWT获取租户ID
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*utils.Claims); ok {
			req.TenantID = userClaims.TenantID
		}
	}

	blacklist, err := h.blackListService.CreateBlackList(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建黑名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "创建黑名单成功", blacklist)
}

// GetBlackListList 获取黑名单列表
// @Summary 获取黑名单列表
// @Description 分页获取黑名单列表
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param type query string false "类型筛选"
// @Param value query string false "值筛选"
// @Param enabled query bool false "状态筛选"
// @Success 200 {object} utils.Response{data=utils.PageData}
// @Failure 500 {object} utils.Response
// @Router /api/v1/blacklists [get]
func (h *BlackListHandler) GetBlackListList(c *gin.Context) {
	var req service.BlackListListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 从JWT获取租户ID
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*utils.Claims); ok {
			req.TenantID = userClaims.TenantID
		}
	}

	blacklists, total, err := h.blackListService.GetBlackListList(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取黑名单列表失败: "+err.Error())
		return
	}

	utils.PageResponse(c, "获取黑名单列表成功", blacklists, total, req.Page, req.PageSize)
}

// GetBlackListByID 获取黑名单详情
// @Summary 获取黑名单详情
// @Description 根据ID获取黑名单详情
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Param id path int true "黑名单ID"
// @Success 200 {object} utils.Response{data=models.BlackList}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/blacklists/{id} [get]
func (h *BlackListHandler) GetBlackListByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var tenantID uint
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*utils.Claims); ok {
			tenantID = userClaims.TenantID
		}
	}

	blacklist, err := h.blackListService.GetBlackListByID(uint(id), tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "黑名单不存在")
		return
	}

	utils.SuccessResponse(c, "获取黑名单详情成功", blacklist)
}

// UpdateBlackList 更新黑名单
// @Summary 更新黑名单
// @Description 更新黑名单信息
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Param id path int true "黑名单ID"
// @Param blacklist body service.UpdateBlackListRequest true "黑名单信息"
// @Success 200 {object} utils.Response{data=models.BlackList}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/blacklists/{id} [put]
func (h *BlackListHandler) UpdateBlackList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req service.UpdateBlackListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	var tenantID uint
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*utils.Claims); ok {
			tenantID = userClaims.TenantID
		}
	}

	blacklist, err := h.blackListService.UpdateBlackList(uint(id), tenantID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新黑名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "更新黑名单成功", blacklist)
}

// DeleteBlackList 删除黑名单
// @Summary 删除黑名单
// @Description 删除黑名单
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Param id path int true "黑名单ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/blacklists/{id} [delete]
func (h *BlackListHandler) DeleteBlackList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var tenantID uint
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*utils.Claims); ok {
			tenantID = userClaims.TenantID
		}
	}

	if err := h.blackListService.DeleteBlackList(uint(id), tenantID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除黑名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "删除黑名单成功", nil)
}

// BatchDeleteBlackList 批量删除黑名单
// @Summary 批量删除黑名单
// @Description 批量删除黑名单
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Param ids body []uint true "黑名单ID列表"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/blacklists/batch [delete]
func (h *BlackListHandler) BatchDeleteBlackList(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if len(ids) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "请选择要删除的黑名单")
		return
	}

	var tenantID uint
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*utils.Claims); ok {
			tenantID = userClaims.TenantID
		}
	}

	if err := h.blackListService.BatchDeleteBlackList(ids, tenantID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "批量删除黑名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "批量删除黑名单成功", nil)
}

// ToggleBlackListStatus 切换黑名单状态
// @Summary 切换黑名单状态
// @Description 启用或禁用黑名单
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Param id path int true "黑名单ID"
// @Success 200 {object} utils.Response{data=models.BlackList}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/blacklists/{id}/toggle [patch]
func (h *BlackListHandler) ToggleBlackListStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var tenantID uint
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*utils.Claims); ok {
			tenantID = userClaims.TenantID
		}
	}

	blacklist, err := h.blackListService.ToggleBlackListStatus(uint(id), tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "切换黑名单状态失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "切换黑名单状态成功", blacklist)
}
