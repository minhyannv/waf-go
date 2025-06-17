package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"waf-go/internal/models"
	"waf-go/internal/service"
	"waf-go/internal/utils"

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

	// 处理时间范围
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			req.StartTime = t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
			req.EndTime = t
		}
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

// GetLogList 获取日志列表
func (h *LogHandler) GetLogList(c *gin.Context) {
	var query models.LogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的查询参数")
		return
	}

	logs, total, err := h.logService.GetLogList(query)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取日志列表失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取日志列表成功", gin.H{
		"total": total,
		"list":  logs,
	})
}

// GetLogDetail 获取日志详情
func (h *LogHandler) GetLogDetail(c *gin.Context) {
	logID := c.Param("id")
	if logID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "日志ID不能为空")
		return
	}

	log, err := h.logService.GetLogDetail(logID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取日志详情失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取日志详情成功", log)
}

// DeleteLog 删除日志
func (h *LogHandler) DeleteLog(c *gin.Context) {
	logID := c.Param("id")
	if logID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "日志ID不能为空")
		return
	}

	if err := h.logService.DeleteLog(logID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除日志失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "删除日志成功", nil)
}

// BatchDeleteLogs 批量删除日志
func (h *LogHandler) BatchDeleteLogs(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	if err := h.logService.BatchDeleteLogs(ids); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "批量删除日志失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "批量删除日志成功", nil)
}
