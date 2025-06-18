package handler

import (
	"net/http"
	"strconv"

	"waf-go/internal/middleware"
	"waf-go/internal/models"
	"waf-go/internal/service"
	"waf-go/internal/utils"

	"github.com/gin-gonic/gin"
)

// BlackListHandler 黑名单处理器
type BlackListHandler struct {
	blackListService *service.BlackListService
}

// NewBlackListHandler 创建黑名单处理器实例
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
	var blackList models.BlackList
	if err := c.ShouldBindJSON(&blackList); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 从JWT中获取租户ID
	userCtx := middleware.GetUserContext(c)
	blackList.TenantID = userCtx.TenantID

	if err := h.blackListService.CreateBlackList(&blackList); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建黑名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "创建黑名单成功", blackList)
}

// GetBlackLists 获取黑名单列表
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
func (h *BlackListHandler) GetBlackLists(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	listType := c.Query("type")
	value := c.Query("value")

	// 处理enabled参数
	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if enabledVal, err := strconv.ParseBool(enabledStr); err == nil {
			enabled = &enabledVal
		}
	}

	// 从JWT中获取租户ID
	userCtx := middleware.GetUserContext(c)

	blackLists, total, err := h.blackListService.GetBlackLists(userCtx.TenantID, page, pageSize, search, listType, value, enabled)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取黑名单列表失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取黑名单列表成功", gin.H{
		"list":  blackLists,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
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

	blackList, err := h.blackListService.GetBlackListByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "黑名单不存在")
		return
	}

	// 检查租户权限
	userCtx := middleware.GetUserContext(c)
	if blackList.TenantID != userCtx.TenantID {
		utils.ErrorResponse(c, http.StatusForbidden, "无权限访问此黑名单")
		return
	}

	utils.SuccessResponse(c, "获取黑名单成功", blackList)
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

	// 先获取现有黑名单
	existingBlackList, err := h.blackListService.GetBlackListByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "黑名单不存在")
		return
	}

	// 检查租户权限
	userCtx := middleware.GetUserContext(c)
	if existingBlackList.TenantID != userCtx.TenantID {
		utils.ErrorResponse(c, http.StatusForbidden, "无权限修改此黑名单")
		return
	}

	// 绑定更新数据
	var updateData models.BlackList
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 更新字段
	if updateData.Type != "" {
		existingBlackList.Type = updateData.Type
	}
	if updateData.Value != "" {
		existingBlackList.Value = updateData.Value
	}
	if updateData.Comment != "" {
		existingBlackList.Comment = updateData.Comment
	}
	existingBlackList.Enabled = updateData.Enabled

	if err := h.blackListService.UpdateBlackList(existingBlackList); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新黑名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "更新黑名单成功", existingBlackList)
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

	// 先获取现有黑名单
	existingBlackList, err := h.blackListService.GetBlackListByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "黑名单不存在")
		return
	}

	// 检查租户权限
	userCtx := middleware.GetUserContext(c)
	if existingBlackList.TenantID != userCtx.TenantID {
		utils.ErrorResponse(c, http.StatusForbidden, "无权限删除此黑名单")
		return
	}

	if err := h.blackListService.DeleteBlackList(uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除黑名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "删除黑名单成功", nil)
}

// ToggleBlackList 切换黑名单状态
// @Summary 切换黑名单状态
// @Description 切换黑名单的启用状态
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Param id path int true "黑名单ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/blacklists/{id}/toggle [post]
func (h *BlackListHandler) ToggleBlackList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	// 先获取现有黑名单
	existingBlackList, err := h.blackListService.GetBlackListByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "黑名单不存在")
		return
	}

	// 检查租户权限
	userCtx := middleware.GetUserContext(c)
	if existingBlackList.TenantID != userCtx.TenantID {
		utils.ErrorResponse(c, http.StatusForbidden, "无权限操作此黑名单")
		return
	}

	err = h.blackListService.ToggleBlackList(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "切换黑名单状态失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "切换黑名单状态成功", nil)
}
