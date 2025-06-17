package handler

import (
	"net/http"
	"strconv"

	"waf-go/internal/service"
	"waf-go/internal/utils"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetDashboardStats 获取仪表盘统计数据
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}

	// 获取租户ID
	var tenantID uint
	role := c.GetString("role")
	if role != "admin" {
		tenantID = c.GetUint("tenant_id")
	}

	stats, err := h.dashboardService.GetDashboardStats(tenantID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// GetRealtimeStats 获取实时攻击统计数据
func (h *DashboardHandler) GetRealtimeStats(c *gin.Context) {
	// 获取租户ID
	var tenantID uint
	role := c.GetString("role")
	if role != "admin" {
		tenantID = c.GetUint("tenant_id")
	}

	stats, err := h.dashboardService.GetRealtimeStats(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// GetOverview 获取仪表盘概览数据
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	if tenantID == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的租户ID")
		return
	}

	overview, err := h.dashboardService.GetOverview(tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取概览数据失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取概览数据成功", overview)
}

// GetAttackTrend 获取攻击趋势数据
func (h *DashboardHandler) GetAttackTrend(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	if tenantID == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的租户ID")
		return
	}

	// 获取查询参数
	timeType := c.DefaultQuery("time_type", "daily")
	days := c.DefaultQuery("days", "7")
	daysInt, err := strconv.Atoi(days)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的天数参数")
		return
	}

	trend, err := h.dashboardService.GetAttackTrend(tenantID, timeType, daysInt)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取攻击趋势数据失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取攻击趋势数据成功", trend)
}

// GetTopRules 获取触发次数最多的规则
func (h *DashboardHandler) GetTopRules(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	if tenantID == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的租户ID")
		return
	}

	rules, err := h.dashboardService.GetTopRules(tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取规则统计数据失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取规则统计数据成功", rules)
}

// GetTopIPs 获取攻击次数最多的IP
func (h *DashboardHandler) GetTopIPs(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	if tenantID == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的租户ID")
		return
	}

	ips, err := h.dashboardService.GetTopIPs(tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取IP统计数据失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取IP统计数据成功", ips)
}

// GetTopURIs 获取攻击次数最多的URI
func (h *DashboardHandler) GetTopURIs(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	if tenantID == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的租户ID")
		return
	}

	uris, err := h.dashboardService.GetTopURIs(tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取URI统计数据失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取URI统计数据成功", uris)
}

// GetTopUserAgents 获取Top攻击User-Agent
func (h *DashboardHandler) GetTopUserAgents(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	if tenantID == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的租户ID")
		return
	}

	results, err := h.dashboardService.GetTopUserAgents(tenantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取User-Agent统计数据失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取User-Agent统计数据成功", results)
}
