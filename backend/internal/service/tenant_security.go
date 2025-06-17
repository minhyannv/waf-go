package service

import (
	"fmt"
	"waf-go/internal/models"

	"gorm.io/gorm"
)

// TenantSecurityService 租户安全控制服务
type TenantSecurityService struct {
	db *gorm.DB
}

// UserContext 用户上下文信息
type UserContext struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Role       string `json:"role"`      // admin, tenant_admin, viewer
	TenantID   uint   `json:"tenant_id"` // 0表示超级管理员
	TenantCode string `json:"tenant_code"`
}

func NewTenantSecurityService(db *gorm.DB) *TenantSecurityService {
	return &TenantSecurityService{db: db}
}

// ValidateTenantAccess 验证用户是否有权限访问指定租户的资源
func (s *TenantSecurityService) ValidateTenantAccess(userCtx *UserContext, targetTenantID uint) error {
	// 超级管理员可以访问所有租户
	if userCtx.Role == "admin" {
		return nil
	}

	// 普通用户只能访问自己租户的资源
	if userCtx.TenantID != targetTenantID {
		return fmt.Errorf("无权限访问租户ID %d的资源", targetTenantID)
	}

	return nil
}

// FilterByTenant 为查询添加租户过滤条件
func (s *TenantSecurityService) FilterByTenant(query *gorm.DB, userCtx *UserContext, allowGlobal bool) *gorm.DB {
	// 超级管理员可以查看所有数据
	if userCtx.Role == "admin" {
		return query
	}

	// 普通用户只能查看自己租户的数据
	if allowGlobal {
		// 允许查看全局数据（tenant_id = 0）和自己租户的数据
		return query.Where("tenant_id = ? OR tenant_id = ?", 0, userCtx.TenantID)
	} else {
		// 只能查看自己租户的数据
		return query.Where("tenant_id = ?", userCtx.TenantID)
	}
}

// ValidateResourceOwnership 验证资源是否属于用户的租户
func (s *TenantSecurityService) ValidateResourceOwnership(userCtx *UserContext, resourceTenantID uint) error {
	// 超级管理员可以操作所有资源
	if userCtx.Role == "admin" {
		return nil
	}

	// 普通用户只能操作自己租户的资源
	if userCtx.TenantID != resourceTenantID {
		return fmt.Errorf("无权限操作其他租户的资源")
	}

	return nil
}

// ValidateDomainOwnership 验证域名是否属于用户的租户
func (s *TenantSecurityService) ValidateDomainOwnership(userCtx *UserContext, domainID uint) error {
	var domain models.Domain
	err := s.db.Select("tenant_id").First(&domain, domainID).Error
	if err != nil {
		return fmt.Errorf("域名不存在")
	}

	return s.ValidateResourceOwnership(userCtx, domain.TenantID)
}

// ValidatePolicyOwnership 验证策略是否属于用户的租户
func (s *TenantSecurityService) ValidatePolicyOwnership(userCtx *UserContext, policyID uint) error {
	var policy models.Policy
	err := s.db.Select("tenant_id").First(&policy, policyID).Error
	if err != nil {
		return fmt.Errorf("策略不存在")
	}

	return s.ValidateResourceOwnership(userCtx, policy.TenantID)
}

// ValidateRuleOwnership 验证规则是否属于用户的租户
func (s *TenantSecurityService) ValidateRuleOwnership(userCtx *UserContext, ruleID uint) error {
	var rule models.Rule
	err := s.db.Select("tenant_id").First(&rule, ruleID).Error
	if err != nil {
		return fmt.Errorf("规则不存在")
	}

	// 全局规则（tenant_id = 0）可以被所有租户使用
	if rule.TenantID == 0 {
		return nil
	}

	return s.ValidateResourceOwnership(userCtx, rule.TenantID)
}

// ValidateBlackListOwnership 验证黑名单是否属于用户的租户
func (s *TenantSecurityService) ValidateBlackListOwnership(userCtx *UserContext, blackListID uint) error {
	var blackList models.BlackList
	err := s.db.Select("tenant_id").First(&blackList, blackListID).Error
	if err != nil {
		return fmt.Errorf("黑名单不存在")
	}

	// 全局黑名单（tenant_id = 0）可以被所有租户使用
	if blackList.TenantID == 0 {
		return nil
	}

	return s.ValidateResourceOwnership(userCtx, blackList.TenantID)
}

// ValidateWhiteListOwnership 验证白名单是否属于用户的租户
func (s *TenantSecurityService) ValidateWhiteListOwnership(userCtx *UserContext, whiteListID uint) error {
	var whiteList models.WhiteList
	err := s.db.Select("tenant_id").First(&whiteList, whiteListID).Error
	if err != nil {
		return fmt.Errorf("白名单不存在")
	}

	// 全局白名单（tenant_id = 0）可以被所有租户使用
	if whiteList.TenantID == 0 {
		return nil
	}

	return s.ValidateResourceOwnership(userCtx, whiteList.TenantID)
}

// ValidateBatchOwnership 批量验证资源所有权
func (s *TenantSecurityService) ValidateBatchOwnership(userCtx *UserContext, tableName string, ids []uint) error {
	if userCtx.Role == "admin" {
		return nil
	}

	if len(ids) == 0 {
		return nil
	}

	var count int64
	err := s.db.Table(tableName).
		Where("id IN ? AND tenant_id = ?", ids, userCtx.TenantID).
		Count(&count).Error
	if err != nil {
		return fmt.Errorf("验证资源所有权失败: %v", err)
	}

	if int(count) != len(ids) {
		return fmt.Errorf("部分资源不属于当前租户")
	}

	return nil
}

// GetUserTenantID 获取用户应该使用的租户ID
func (s *TenantSecurityService) GetUserTenantID(userCtx *UserContext, requestTenantID uint) uint {
	// 超级管理员可以指定任何租户ID
	if userCtx.Role == "admin" {
		if requestTenantID > 0 {
			return requestTenantID
		}
		return userCtx.TenantID
	}

	// 普通用户只能使用自己的租户ID
	return userCtx.TenantID
}

// ValidateRuleIDs 验证规则ID列表是否都属于用户有权限的范围
func (s *TenantSecurityService) ValidateRuleIDs(userCtx *UserContext, ruleIDs []uint) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	if userCtx.Role == "admin" {
		// 超级管理员可以使用所有规则
		var count int64
		err := s.db.Model(&models.Rule{}).
			Where("id IN ?", ruleIDs).
			Count(&count).Error
		if err != nil {
			return fmt.Errorf("验证规则ID失败: %v", err)
		}
		if int(count) != len(ruleIDs) {
			return fmt.Errorf("部分规则不存在")
		}
		return nil
	}

	// 普通用户只能使用全局规则和自己租户的规则
	var count int64
	err := s.db.Model(&models.Rule{}).
		Where("id IN ? AND (tenant_id = ? OR tenant_id = ?)", ruleIDs, 0, userCtx.TenantID).
		Count(&count).Error
	if err != nil {
		return fmt.Errorf("验证规则ID失败: %v", err)
	}

	if int(count) != len(ruleIDs) {
		return fmt.Errorf("部分规则不存在或无权限访问")
	}

	return nil
}

// ValidatePolicyIDs 验证策略ID列表是否都属于用户有权限的范围
func (s *TenantSecurityService) ValidatePolicyIDs(userCtx *UserContext, policyIDs []uint) error {
	if len(policyIDs) == 0 {
		return nil
	}

	if userCtx.Role == "admin" {
		// 超级管理员可以使用所有策略
		var count int64
		err := s.db.Model(&models.Policy{}).
			Where("id IN ?", policyIDs).
			Count(&count).Error
		if err != nil {
			return fmt.Errorf("验证策略ID失败: %v", err)
		}
		if int(count) != len(policyIDs) {
			return fmt.Errorf("部分策略不存在")
		}
		return nil
	}

	// 普通用户只能使用全局策略和自己租户的策略
	var count int64
	err := s.db.Model(&models.Policy{}).
		Where("id IN ? AND (tenant_id = ? OR tenant_id = ?)", policyIDs, 0, userCtx.TenantID).
		Count(&count).Error
	if err != nil {
		return fmt.Errorf("验证策略ID失败: %v", err)
	}

	if int(count) != len(policyIDs) {
		return fmt.Errorf("部分策略不存在或无权限访问")
	}

	return nil
}
