package handler

import (
	"net/http"
	"strconv"

	"waf-go/internal/service"
	"waf-go/internal/utils"

	"github.com/gin-gonic/gin"
)

type WhiteListHandler struct {
	whiteListService *service.WhiteListService
}

func NewWhiteListHandler(whiteListService *service.WhiteListService) *WhiteListHandler {
	return &WhiteListHandler{
		whiteListService: whiteListService,
	}
}

// CreateWhiteList 创建白名单
// @Summary 创建白名单
// @Description 创建新的白名单条目
// @Tags 白名单管理
// @Accept json
// @Produce json
// @Param whitelist body service.CreateWhiteListRequest true "白名单信息"
// @Success 200 {object} utils.Response{data=models.WhiteList}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/whitelists [post]
func (h *WhiteListHandler) CreateWhiteList(c *gin.Context) {
	var req service.CreateWhiteListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 从JWT获取租户ID
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*service.JWTClaims); ok {
			req.TenantID = userClaims.TenantID
		}
	}

	whitelist, err := h.whiteListService.CreateWhiteList(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建白名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "创建白名单成功", whitelist)
}

// GetWhiteListList 获取白名单列表
// @Summary 获取白名单列表
// @Description 分页获取白名单列表
// @Tags 白名单管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param type query string false "类型筛选"
// @Param value query string false "值筛选"
// @Param enabled query bool false "状态筛选"
// @Success 200 {object} utils.Response{data=utils.PageResponse}
// @Failure 500 {object} utils.Response
// @Router /api/v1/whitelists [get]
func (h *WhiteListHandler) GetWhiteListList(c *gin.Context) {
	var req service.WhiteListListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 从JWT获取租户ID
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*service.JWTClaims); ok {
			req.TenantID = userClaims.TenantID
		}
	}

	whitelists, total, err := h.whiteListService.GetWhiteListList(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取白名单列表失败: "+err.Error())
		return
	}

	utils.PageResponse(c, "获取白名单列表成功", whitelists, total, req.Page, req.PageSize)
}

// GetWhiteListByID 获取白名单详情
// @Summary 获取白名单详情
// @Description 根据ID获取白名单详情
// @Tags 白名单管理
// @Accept json
// @Produce json
// @Param id path int true "白名单ID"
// @Success 200 {object} utils.Response{data=models.WhiteList}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/whitelists/{id} [get]
func (h *WhiteListHandler) GetWhiteListByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var tenantID uint
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*service.JWTClaims); ok {
			tenantID = userClaims.TenantID
		}
	}

	whitelist, err := h.whiteListService.GetWhiteListByID(uint(id), tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "白名单不存在")
		return
	}

	utils.SuccessResponse(c, "获取白名单详情成功", whitelist)
}

// UpdateWhiteList 更新白名单
// @Summary 更新白名单
// @Description 更新白名单信息
// @Tags 白名单管理
// @Accept json
// @Produce json
// @Param id path int true "白名单ID"
// @Param whitelist body service.UpdateWhiteListRequest true "白名单信息"
// @Success 200 {object} utils.Response{data=models.WhiteList}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/whitelists/{id} [put]
func (h *WhiteListHandler) UpdateWhiteList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req service.UpdateWhiteListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	var tenantID uint
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*service.JWTClaims); ok {
			tenantID = userClaims.TenantID
		}
	}

	whitelist, err := h.whiteListService.UpdateWhiteList(uint(id), tenantID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新白名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "更新白名单成功", whitelist)
}

// DeleteWhiteList 删除白名单
// @Summary 删除白名单
// @Description 删除白名单
// @Tags 白名单管理
// @Accept json
// @Produce json
// @Param id path int true "白名单ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/whitelists/{id} [delete]
func (h *WhiteListHandler) DeleteWhiteList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var tenantID uint
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*service.JWTClaims); ok {
			tenantID = userClaims.TenantID
		}
	}

	if err := h.whiteListService.DeleteWhiteList(uint(id), tenantID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除白名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "删除白名单成功", nil)
}

// BatchDeleteWhiteList 批量删除白名单
// @Summary 批量删除白名单
// @Description 批量删除白名单
// @Tags 白名单管理
// @Accept json
// @Produce json
// @Param ids body []uint true "白名单ID列表"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/whitelists/batch [delete]
func (h *WhiteListHandler) BatchDeleteWhiteList(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if len(ids) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "请选择要删除的白名单")
		return
	}

	var tenantID uint
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*service.JWTClaims); ok {
			tenantID = userClaims.TenantID
		}
	}

	if err := h.whiteListService.BatchDeleteWhiteList(ids, tenantID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "批量删除白名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "批量删除白名单成功", nil)
}

// ToggleWhiteListStatus 切换白名单状态
// @Summary 切换白名单状态
// @Description 启用或禁用白名单
// @Tags 白名单管理
// @Accept json
// @Produce json
// @Param id path int true "白名单ID"
// @Success 200 {object} utils.Response{data=models.WhiteList}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/whitelists/{id}/toggle [patch]
func (h *WhiteListHandler) ToggleWhiteListStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var tenantID uint
	claims, exists := c.Get("claims")
	if exists {
		if userClaims, ok := claims.(*service.JWTClaims); ok {
			tenantID = userClaims.TenantID
		}
	}

	whitelist, err := h.whiteListService.ToggleWhiteListStatus(uint(id), tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "切换白名单状态失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "切换白名单状态成功", whitelist)
}
