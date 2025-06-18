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

// WhiteListHandler 白名单处理器
type WhiteListHandler struct {
	whiteListService *service.WhiteListService
}

// NewWhiteListHandler 创建白名单处理器实例
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
	var whiteList models.WhiteList
	if err := c.ShouldBindJSON(&whiteList); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 从JWT中获取租户ID
	userCtx := middleware.GetUserContext(c)
	whiteList.TenantID = userCtx.TenantID

	if err := h.whiteListService.CreateWhiteList(&whiteList); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建白名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "创建白名单成功", whiteList)
}

// GetWhiteLists 获取白名单列表
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
func (h *WhiteListHandler) GetWhiteLists(c *gin.Context) {
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

	whiteLists, total, err := h.whiteListService.GetWhiteLists(userCtx.TenantID, page, pageSize, search, listType, value, enabled)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取白名单列表失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取白名单列表成功", gin.H{
		"list":  whiteLists,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
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

	whiteList, err := h.whiteListService.GetWhiteListByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "白名单不存在")
		return
	}

	// 检查租户权限
	userCtx := middleware.GetUserContext(c)
	if whiteList.TenantID != userCtx.TenantID {
		utils.ErrorResponse(c, http.StatusForbidden, "无权限访问此白名单")
		return
	}

	utils.SuccessResponse(c, "获取白名单成功", whiteList)
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

	// 先获取现有白名单
	existingWhiteList, err := h.whiteListService.GetWhiteListByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "白名单不存在")
		return
	}

	// 检查租户权限
	userCtx := middleware.GetUserContext(c)
	if existingWhiteList.TenantID != userCtx.TenantID {
		utils.ErrorResponse(c, http.StatusForbidden, "无权限修改此白名单")
		return
	}

	// 绑定更新数据
	var updateData models.WhiteList
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 更新字段
	if updateData.Type != "" {
		existingWhiteList.Type = updateData.Type
	}
	if updateData.Value != "" {
		existingWhiteList.Value = updateData.Value
	}
	if updateData.Comment != "" {
		existingWhiteList.Comment = updateData.Comment
	}
	existingWhiteList.Enabled = updateData.Enabled

	if err := h.whiteListService.UpdateWhiteList(existingWhiteList); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新白名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "更新白名单成功", existingWhiteList)
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

	// 先获取现有白名单
	existingWhiteList, err := h.whiteListService.GetWhiteListByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "白名单不存在")
		return
	}

	// 检查租户权限
	userCtx := middleware.GetUserContext(c)
	if existingWhiteList.TenantID != userCtx.TenantID {
		utils.ErrorResponse(c, http.StatusForbidden, "无权限删除此白名单")
		return
	}

	if err := h.whiteListService.DeleteWhiteList(uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除白名单失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "删除白名单成功", nil)
}

// ToggleWhiteList 切换白名单状态
// @Summary 切换白名单状态
// @Description 切换白名单的启用状态
// @Tags 白名单管理
// @Accept json
// @Produce json
// @Param id path int true "白名单ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/whitelists/{id}/toggle [post]
func (h *WhiteListHandler) ToggleWhiteList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	// 先获取现有白名单
	existingWhiteList, err := h.whiteListService.GetWhiteListByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "白名单不存在")
		return
	}

	// 检查租户权限
	userCtx := middleware.GetUserContext(c)
	if existingWhiteList.TenantID != userCtx.TenantID {
		utils.ErrorResponse(c, http.StatusForbidden, "无权限操作此白名单")
		return
	}

	err = h.whiteListService.ToggleWhiteList(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "切换白名单状态失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "切换白名单状态成功", nil)
}
