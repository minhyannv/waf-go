package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"waf-go/internal/service"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService *service.LogService
}

func NewLogHandler(logService *service.LogService) *LogHandler {
	return &LogHandler{
		logService: logService,
	}
}

// GetAttackLogList 获取攻击日志列表
func (h *LogHandler) GetAttackLogList(c *gin.Context) {
	var req service.LogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	// 设置租户ID
	role := c.GetString("role")
	if role != "admin" {
		req.TenantID = c.GetUint("tenant_id")
	}

	logs, total, err := h.logService.GetAttackLogList(&req)
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
		"data": gin.H{
			"list":  logs,
			"total": total,
			"page":  req.Page,
			"size":  req.PageSize,
		},
	})
}

// GetAttackLogDetail 获取攻击日志详情
func (h *LogHandler) GetAttackLogDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的日志ID",
			"data":    nil,
		})
		return
	}

	log, err := h.logService.GetAttackLogByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "日志不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    log,
	})
}

// DeleteAttackLog 删除攻击日志
func (h *LogHandler) DeleteAttackLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的日志ID",
			"data":    nil,
		})
		return
	}

	err = h.logService.DeleteAttackLog(uint(id))
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
		"message": "删除成功",
		"data":    nil,
	})
}

// BatchDeleteAttackLogs 批量删除攻击日志
func (h *LogHandler) BatchDeleteAttackLogs(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请选择要删除的日志",
			"data":    nil,
		})
		return
	}

	err := h.logService.DeleteAttackLogs(req.IDs)
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
		"message": "批量删除成功",
		"data":    nil,
	})
}

// ExportAttackLogs 导出攻击日志
func (h *LogHandler) ExportAttackLogs(c *gin.Context) {
	var req struct {
		IDs []uint `form:"ids"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	logs, err := h.logService.ExportAttackLogs(req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=attack_logs.json")

	// 直接返回JSON数据
	jsonData, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "导出失败",
			"data":    nil,
		})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonData)
}

// CleanOldLogs 清理旧日志
func (h *LogHandler) CleanOldLogs(c *gin.Context) {
	var req struct {
		Days int `json:"days" binding:"required,min=1,max=365"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	count, err := h.logService.CleanOldLogs(req.Days)
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
		"message": "清理完成",
		"data":    count,
	})
}
