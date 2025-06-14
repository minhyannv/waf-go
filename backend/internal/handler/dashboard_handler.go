package handler

import (
	"net/http"
	"strconv"

	"waf-go/internal/service"

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
