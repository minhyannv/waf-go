package handler

import (
	"net/http"
	"strconv"

	"waf-go/internal/logger"
	"waf-go/internal/service"
	"waf-go/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PolicyHandler struct {
	policyService *service.PolicyService
}

func NewPolicyHandler(policyService *service.PolicyService) *PolicyHandler {
	return &PolicyHandler{
		policyService: policyService,
	}
}

// CreatePolicy 创建策略
func (h *PolicyHandler) CreatePolicy(c *gin.Context) {
	var req service.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("CreatePolicy - 参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	// 设置租户ID
	tenantID := c.GetUint("tenant_id")
	req.TenantID = tenantID

	// 规则ID验证现在在服务层处理

	policy, err := h.policyService.CreatePolicy(&req)
	if err != nil {
		logger.Error("CreatePolicy - 服务层错误", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	logger.Info("CreatePolicy - 创建成功", zap.Uint("policy_id", policy.ID), zap.String("name", policy.Name))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    policy,
	})
}

// GetPolicyList 获取策略列表
func (h *PolicyHandler) GetPolicyList(c *gin.Context) {
	var req service.PolicyListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("GetPolicyList - 查询参数绑定失败", zap.Error(err))
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

	policies, total, err := h.policyService.GetPolicyList(&req)
	if err != nil {
		logger.Error("GetPolicyList - 服务层错误", zap.Error(err))
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
			"list":  policies,
			"total": total,
			"page":  req.Page,
			"size":  req.PageSize,
		},
	})
}

// GetPolicy 获取策略详情
func (h *PolicyHandler) GetPolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("GetPolicy - ID参数解析失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的策略ID",
			"data":    nil,
		})
		return
	}

	policy, err := h.policyService.GetPolicyByID(uint(id))
	if err != nil {
		logger.Error("GetPolicy - 服务层错误", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "策略不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    policy,
	})
}

// GetPolicyWithRules 获取策略及其关联的规则
func (h *PolicyHandler) GetPolicyWithRules(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("GetPolicyWithRules - ID参数解析失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的策略ID",
			"data":    nil,
		})
		return
	}

	policyWithRules, err := h.policyService.GetPolicyWithRules(uint(id))
	if err != nil {
		logger.Error("GetPolicyWithRules - 服务层错误", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "策略不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    policyWithRules,
	})
}

// UpdatePolicy 更新策略
func (h *PolicyHandler) UpdatePolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("UpdatePolicy - ID参数解析失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的策略ID",
			"data":    nil,
		})
		return
	}

	var req service.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("UpdatePolicy - 参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	// 规则ID验证现在在服务层处理

	policy, err := h.policyService.UpdatePolicy(uint(id), &req)
	if err != nil {
		logger.Error("UpdatePolicy - 服务层错误", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	logger.Info("UpdatePolicy - 更新成功", zap.Uint("policy_id", policy.ID))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    policy,
	})
}

// DeletePolicy 删除策略
func (h *PolicyHandler) DeletePolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("DeletePolicy - ID参数解析失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的策略ID",
			"data":    nil,
		})
		return
	}

	err = h.policyService.DeletePolicy(uint(id))
	if err != nil {
		logger.Error("DeletePolicy - 服务层错误", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	logger.Info("DeletePolicy - 删除成功", zap.Uint("policy_id", uint(id)))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    nil,
	})
}

// TogglePolicy 切换策略启用状态
func (h *PolicyHandler) TogglePolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("TogglePolicy - ID参数解析失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的策略ID",
			"data":    nil,
		})
		return
	}

	err = h.policyService.TogglePolicy(uint(id))
	if err != nil {
		logger.Error("TogglePolicy - 服务层错误", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	logger.Info("TogglePolicy - 状态切换成功", zap.Uint("policy_id", uint(id)))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "操作成功",
		"data":    nil,
	})
}

// GetAvailableRules 获取可用的规则列表
func (h *PolicyHandler) GetAvailableRules(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	role := c.GetString("role")

	logger.Debug("GetAvailableRules - Handler received", zap.Uint("tenant_id", tenantID), zap.String("role", role))

	rules, err := h.policyService.GetAvailableRules(tenantID, role)
	if err != nil {
		logger.Error("GetAvailableRules - 服务层错误", zap.Error(err))
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
		"data":    rules,
	})
}

// BatchDeletePolicies 批量删除策略
func (h *PolicyHandler) BatchDeletePolicies(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("BatchDeletePolicies - 参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	err := h.policyService.BatchDeletePolicies(req.IDs)
	if err != nil {
		logger.Error("BatchDeletePolicies - 服务层错误", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	logger.Info("BatchDeletePolicies - 批量删除成功", zap.Any("policy_ids", req.IDs))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "批量删除成功",
		"data":    nil,
	})
}

// GetPolicyRules 获取策略规则列表
func (h *PolicyHandler) GetPolicyRules(c *gin.Context) {
	policyIDStr := c.Param("id")
	if policyIDStr == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "策略ID不能为空")
		return
	}

	policyID, err := strconv.ParseUint(policyIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的策略ID")
		return
	}

	rules, err := h.policyService.GetPolicyRules(uint(policyID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取策略规则失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取策略规则成功", rules)
}

// UpdatePolicyRules 更新策略规则
func (h *PolicyHandler) UpdatePolicyRules(c *gin.Context) {
	policyIDStr := c.Param("id")
	if policyIDStr == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "策略ID不能为空")
		return
	}

	policyID, err := strconv.ParseUint(policyIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的策略ID")
		return
	}

	var ruleIDs []uint
	if err := c.ShouldBindJSON(&ruleIDs); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	if err := h.policyService.UpdatePolicyRules(uint(policyID), ruleIDs); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新策略规则失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "更新策略规则成功", nil)
}
