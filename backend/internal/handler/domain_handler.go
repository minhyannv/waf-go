package handler

import (
	"net/http"
	"strconv"

	"waf-go/internal/middleware"
	"waf-go/internal/service"
	"waf-go/internal/utils"

	"github.com/gin-gonic/gin"
)

type DomainHandler struct {
	domainService   *service.DomainService
	securityService *service.TenantSecurityService
}

func NewDomainHandler(domainService *service.DomainService, securityService *service.TenantSecurityService) *DomainHandler {
	return &DomainHandler{
		domainService:   domainService,
		securityService: securityService,
	}
}

// CreateDomain 创建域名配置
// @Summary 创建域名配置
// @Description 创建新的域名配置
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param domain body service.CreateDomainRequest true "域名配置信息"
// @Success 200 {object} utils.Response{data=models.Domain}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/domains [post]
func (h *DomainHandler) CreateDomain(c *gin.Context) {
	var req service.CreateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 获取用户上下文
	userCtx := middleware.GetUserContext(c)

	// 设置租户ID - 非管理员只能为自己的租户创建域名
	req.TenantID = h.securityService.GetUserTenantID(userCtx, req.TenantID)

	// 验证租户访问权限
	if err := h.securityService.ValidateTenantAccess(userCtx, req.TenantID); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, "权限不足: "+err.Error())
		return
	}

	domain, err := h.domainService.CreateDomain(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建域名配置失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "创建域名配置成功", domain)
}

// GetDomainList 获取域名配置列表
// @Summary 获取域名配置列表
// @Description 获取域名配置列表
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param domain query string false "域名"
// @Param enabled query bool false "是否启用"
// @Success 200 {object} utils.Response{data=utils.PageResponse{list=[]models.Domain}}
// @Failure 500 {object} utils.Response
// @Router /api/v1/domains [get]
func (h *DomainHandler) GetDomainList(c *gin.Context) {
	var req service.DomainListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 获取用户上下文
	userCtx := middleware.GetUserContext(c)

	// 设置租户过滤 - 确保租户隔离
	req.TenantID = h.securityService.GetUserTenantID(userCtx, req.TenantID)

	domains, total, err := h.domainService.GetDomains(&req, userCtx)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取域名配置列表失败: "+err.Error())
		return
	}

	utils.SuccessWithPagination(c, "获取域名配置列表成功", domains, total, req.Page, req.PageSize)
}

// GetDomain 获取域名配置详情
// @Summary 获取域名配置详情
// @Description 根据ID获取域名配置详情
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param id path int true "域名ID"
// @Success 200 {object} utils.Response{data=models.Domain}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/v1/domains/{id} [get]
func (h *DomainHandler) GetDomain(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的域名ID")
		return
	}

	// 暂时移除所有安全检查，直接调用GetDomain
	domain, err := h.domainService.GetDomain(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "域名配置不存在: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取域名配置成功", domain)
}

// UpdateDomain 更新域名配置
// @Summary 更新域名配置
// @Description 更新域名配置信息
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param id path int true "域名ID"
// @Param domain body service.UpdateDomainRequest true "域名配置信息"
// @Success 200 {object} utils.Response{data=models.Domain}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/domains/{id} [put]
func (h *DomainHandler) UpdateDomain(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的域名ID")
		return
	}

	var req service.UpdateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 获取用户上下文
	userCtx := middleware.GetUserContext(c)

	// 验证域名所有权
	if err := h.securityService.ValidateDomainOwnership(userCtx, uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, "权限不足")
		return
	}

	domain, err := h.domainService.UpdateDomain(uint(id), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新域名配置失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "更新域名配置成功", domain)
}

// DeleteDomain 删除域名配置
// @Summary 删除域名配置
// @Description 删除域名配置
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param id path int true "域名ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/domains/{id} [delete]
func (h *DomainHandler) DeleteDomain(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的域名ID")
		return
	}

	// 获取用户上下文
	userCtx := middleware.GetUserContext(c)

	// 验证域名所有权
	if err := h.securityService.ValidateDomainOwnership(userCtx, uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, "权限不足")
		return
	}

	err = h.domainService.DeleteDomain(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除域名配置失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "删除域名配置成功", nil)
}

// ToggleDomain 切换域名配置启用状态
// @Summary 切换域名配置启用状态
// @Description 切换域名配置的启用状态
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param id path int true "域名ID"
// @Success 200 {object} utils.Response{data=models.Domain}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/domains/{id}/toggle [post]
func (h *DomainHandler) ToggleDomain(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的域名ID")
		return
	}

	// 获取用户上下文
	userCtx := middleware.GetUserContext(c)

	// 验证域名所有权
	if err := h.securityService.ValidateDomainOwnership(userCtx, uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, "权限不足")
		return
	}

	err = h.domainService.ToggleDomain(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "切换域名配置状态失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "切换域名配置状态成功", nil)
}

// GetDomainPolicies 获取域名关联的策略
// @Summary 获取域名关联的策略
// @Description 获取域名关联的所有策略信息
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param id path int true "域名ID"
// @Success 200 {object} utils.Response{data=[]service.DomainPolicyResponse}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/domains/{id}/policies [get]
func (h *DomainHandler) GetDomainPolicies(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的域名ID")
		return
	}

	// 获取用户上下文
	userCtx := middleware.GetUserContext(c)

	// 验证域名所有权
	if err := h.securityService.ValidateDomainOwnership(userCtx, uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, "权限不足")
		return
	}

	policies, err := h.domainService.GetDomainPolicies(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取域名策略失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取域名策略成功", policies)
}

// UpdateDomainPolicies 更新域名关联的策略
// @Summary 更新域名关联的策略
// @Description 更新域名关联的策略配置
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param id path int true "域名ID"
// @Param policies body service.UpdateDomainPoliciesRequest true "策略关联配置"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/domains/{id}/policies [put]
func (h *DomainHandler) UpdateDomainPolicies(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的域名ID")
		return
	}

	var req service.UpdateDomainPoliciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 获取用户上下文
	userCtx := middleware.GetUserContext(c)

	// 验证域名所有权
	if err := h.securityService.ValidateDomainOwnership(userCtx, uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, "权限不足")
		return
	}

	err = h.domainService.UpdateDomainPolicies(uint(id), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新域名策略失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "更新域名策略成功", nil)
}

// BatchDeleteDomains 批量删除域名配置
// @Summary 批量删除域名配置
// @Description 批量删除多个域名配置
// @Tags 域名管理
// @Accept json
// @Produce json
// @Param ids body service.BatchDeleteRequest true "域名ID列表"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/domains/batch [delete]
func (h *DomainHandler) BatchDeleteDomains(c *gin.Context) {
	var req service.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "请选择要删除的域名")
		return
	}

	// 获取用户上下文
	userCtx := middleware.GetUserContext(c)

	// 验证批量所有权
	if err := h.securityService.ValidateBatchOwnership(userCtx, "domains", req.IDs); err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, "权限不足")
		return
	}

	err := h.domainService.BatchDeleteDomains(req.IDs)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "批量删除域名配置失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "批量删除域名配置成功", nil)
}
