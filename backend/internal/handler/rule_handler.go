package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"waf-go/internal/logger"
	"waf-go/internal/service"
	"waf-go/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RuleHandler struct {
	ruleService *service.RuleService
}

func NewRuleHandler(ruleService *service.RuleService) *RuleHandler {
	return &RuleHandler{
		ruleService: ruleService,
	}
}

// CreateRule 创建规则
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var req service.CreateRuleRequest

	// 读取请求体用于日志记录
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	bodyString := string(bodyBytes)

	// 重新设置请求体，因为ReadAll会消耗掉
	c.Request.Body = io.NopCloser(strings.NewReader(bodyString))

	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用zap记录详细的错误信息
		logger.Error("CreateRule - 参数绑定失败",
			zap.Error(err),
			zap.String("request_body", bodyString),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.GetHeader("User-Agent")),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("请求参数错误: %v", err),
			"data":    nil,
		})
		return
	}

	// 记录成功绑定的请求参数
	logger.Info("CreateRule - 请求参数绑定成功",
		zap.String("name", req.Name),
		zap.String("description", req.Description),
		zap.String("match_type", req.MatchType),
		zap.String("match_mode", req.MatchMode),
		zap.String("pattern", req.Pattern),
		zap.String("action", req.Action),
		zap.Int("priority", req.Priority),
		zap.Bool("enabled", req.Enabled),
		zap.String("client_ip", c.ClientIP()),
	)

	// 设置租户ID
	tenantID := c.GetUint("tenant_id")
	req.TenantID = tenantID

	logger.Debug("CreateRule - 设置租户ID",
		zap.Uint("tenant_id", tenantID),
	)

	rule, err := h.ruleService.CreateRule(&req)
	if err != nil {
		logger.Error("CreateRule - 服务层错误",
			zap.Error(err),
			zap.String("name", req.Name),
			zap.String("match_type", req.MatchType),
			zap.String("pattern", req.Pattern),
			zap.Uint("tenant_id", tenantID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	logger.Info("CreateRule - 创建成功",
		zap.Uint("rule_id", rule.ID),
		zap.String("name", rule.Name),
		zap.String("match_type", rule.MatchType),
		zap.String("pattern", rule.Pattern),
		zap.Uint("tenant_id", rule.TenantID),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    rule,
	})
}

// GetRuleList 获取规则列表
func (h *RuleHandler) GetRuleList(c *gin.Context) {
	var req service.RuleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("GetRuleList - 查询参数绑定失败",
			zap.Error(err),
			zap.String("query_string", c.Request.URL.RawQuery),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("请求参数错误: %v", err),
			"data":    nil,
		})
		return
	}

	// 设置租户ID
	role := c.GetString("role")
	if role != "admin" {
		req.TenantID = c.GetUint("tenant_id")
	}

	logger.Debug("GetRuleList - 查询参数",
		zap.Int("page", req.Page),
		zap.Int("page_size", req.PageSize),
		zap.String("name", req.Name),
		zap.String("match_type", req.MatchType),
		zap.Bool("enabled", req.Enabled != nil && *req.Enabled),
		zap.String("role", role),
		zap.Uint("tenant_id", req.TenantID),
	)

	rules, total, err := h.ruleService.GetRuleList(&req)
	if err != nil {
		logger.Error("GetRuleList - 服务层错误",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("page_size", req.PageSize),
			zap.Uint("tenant_id", req.TenantID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	logger.Debug("GetRuleList - 查询成功",
		zap.Int64("total", total),
		zap.Int("returned_count", len(rules)),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  rules,
			"total": total,
			"page":  req.Page,
			"size":  req.PageSize,
		},
	})
}

// GetRule 获取规则详情
func (h *RuleHandler) GetRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("GetRule - ID参数解析失败",
			zap.Error(err),
			zap.String("id_param", idStr),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则ID",
			"data":    nil,
		})
		return
	}

	rule, err := h.ruleService.GetRule(uint(id))
	if err != nil {
		logger.Error("GetRule - 服务层错误",
			zap.Error(err),
			zap.Uint("rule_id", uint(id)),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "规则不存在",
			"data":    nil,
		})
		return
	}

	logger.Debug("GetRule - 查询成功",
		zap.Uint("rule_id", rule.ID),
		zap.String("name", rule.Name),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    rule,
	})
}

// UpdateRule 更新规则
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("UpdateRule - ID参数解析失败",
			zap.Error(err),
			zap.String("id_param", idStr),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则ID",
			"data":    nil,
		})
		return
	}

	var req service.UpdateRuleRequest

	// 读取请求体用于日志记录
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	bodyString := string(bodyBytes)
	c.Request.Body = io.NopCloser(strings.NewReader(bodyString))

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("UpdateRule - 参数绑定失败",
			zap.Error(err),
			zap.Uint("rule_id", uint(id)),
			zap.String("request_body", bodyString),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("请求参数错误: %v", err),
			"data":    nil,
		})
		return
	}

	logger.Info("UpdateRule - 更新请求",
		zap.Uint("rule_id", uint(id)),
		zap.String("name", req.Name),
		zap.String("match_type", req.MatchType),
		zap.String("pattern", req.Pattern),
		zap.String("client_ip", c.ClientIP()),
	)

	rule, err := h.ruleService.UpdateRule(uint(id), &req)
	if err != nil {
		logger.Error("UpdateRule - 服务层错误",
			zap.Error(err),
			zap.Uint("rule_id", uint(id)),
			zap.String("name", req.Name),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	logger.Info("UpdateRule - 更新成功",
		zap.Uint("rule_id", rule.ID),
		zap.String("name", rule.Name),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    rule,
	})
}

// DeleteRule 删除规则
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则ID",
			"data":    nil,
		})
		return
	}

	err = h.ruleService.DeleteRule(uint(id))
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

// ToggleRule 切换规则状态
func (h *RuleHandler) ToggleRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的规则ID",
			"data":    nil,
		})
		return
	}

	err = h.ruleService.ToggleRule(uint(id))
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
		"message": "操作成功",
		"data":    nil,
	})
}

// BatchDeleteRules 批量删除规则
func (h *RuleHandler) BatchDeleteRules(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	if err := h.ruleService.BatchDeleteRules(ids); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除规则失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "规则删除成功", nil)
}
