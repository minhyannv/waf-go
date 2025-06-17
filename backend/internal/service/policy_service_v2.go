package service

import (
	"fmt"
	"waf-go/internal/models"

	"gorm.io/gorm"
)

type PolicyServiceV2 struct {
	db *gorm.DB
}

// PolicyRequest 策略请求结构
type PolicyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	RuleIDs     []uint `json:"rule_ids"`
	Enabled     bool   `json:"enabled"`
	TenantID    uint   `json:"tenant_id"`
}

// UpdatePolicyRequestV2 更新策略请求
type UpdatePolicyRequestV2 struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	RuleIDs     *[]uint `json:"rule_ids"`
	Enabled     *bool   `json:"enabled"`
}

// PolicyListRequestV2 策略列表请求
type PolicyListRequestV2 struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
	Name     string `form:"name"`
	Enabled  *bool  `form:"enabled"`
	TenantID uint   `form:"tenant_id"`
}

// PolicyWithRulesResponse 策略及其关联规则响应
type PolicyWithRulesResponse struct {
	models.Policy
	Rules []PolicyRuleResponse `json:"rules"`
}

// PolicyRuleResponse 策略规则关联响应
type PolicyRuleResponse struct {
	ID       uint         `json:"id"`
	PolicyID uint         `json:"policy_id"`
	RuleID   uint         `json:"rule_id"`
	Priority int          `json:"priority"`
	Enabled  bool         `json:"enabled"`
	Rule     *models.Rule `json:"rule,omitempty"`
}

func NewPolicyServiceV2(db *gorm.DB) *PolicyServiceV2 {
	return &PolicyServiceV2{db: db}
}

// CreatePolicy 创建策略
func (s *PolicyServiceV2) CreatePolicy(req *PolicyRequest) (*models.Policy, error) {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建策略
	policy := &models.Policy{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		TenantID:    req.TenantID,
	}

	err := tx.Create(policy).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建策略失败: %v", err)
	}

	// 创建策略规则关联
	if len(req.RuleIDs) > 0 {
		for i, ruleID := range req.RuleIDs {
			policyRule := models.PolicyRule{
				PolicyID: policy.ID,
				RuleID:   ruleID,
				Priority: i + 1, // 根据顺序设置优先级
				Enabled:  true,
			}

			err = tx.Create(&policyRule).Error
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("创建策略规则关联失败: %v", err)
			}
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return policy, nil
}

// GetPolicyList 获取策略列表
func (s *PolicyServiceV2) GetPolicyList(req *PolicyListRequestV2) ([]models.Policy, int64, error) {
	var policies []models.Policy
	var total int64

	query := s.db.Model(&models.Policy{})

	// 添加筛选条件
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Enabled != nil {
		query = query.Where("enabled = ?", *req.Enabled)
	}
	if req.TenantID > 0 {
		query = query.Where("tenant_id = ?", req.TenantID)
	}

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	err = query.Offset(offset).Limit(req.PageSize).Order("id DESC").
		Preload("Tenant").Find(&policies).Error
	if err != nil {
		return nil, 0, err
	}

	return policies, total, nil
}

// GetPolicyByID 根据ID获取策略
func (s *PolicyServiceV2) GetPolicyByID(id uint) (*models.Policy, error) {
	var policy models.Policy
	err := s.db.Preload("Tenant").First(&policy, id).Error
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

// GetPolicyWithRules 获取策略及其关联的规则
func (s *PolicyServiceV2) GetPolicyWithRules(id uint) (*PolicyWithRulesResponse, error) {
	var policy models.Policy
	err := s.db.Preload("Tenant").First(&policy, id).Error
	if err != nil {
		return nil, err
	}

	// 获取策略关联的规则
	var policyRules []models.PolicyRule
	err = s.db.Where("policy_id = ?", id).
		Preload("Rule").
		Order("priority DESC").
		Find(&policyRules).Error
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	ruleResponses := make([]PolicyRuleResponse, len(policyRules))
	for i, pr := range policyRules {
		ruleResponses[i] = PolicyRuleResponse{
			ID:       pr.ID,
			PolicyID: pr.PolicyID,
			RuleID:   pr.RuleID,
			Priority: pr.Priority,
			Enabled:  pr.Enabled,
			Rule:     pr.Rule,
		}
	}

	response := &PolicyWithRulesResponse{
		Policy: policy,
		Rules:  ruleResponses,
	}

	return response, nil
}

// UpdatePolicy 更新策略
func (s *PolicyServiceV2) UpdatePolicy(id uint, req *UpdatePolicyRequestV2) (*models.Policy, error) {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var policy models.Policy
	err := tx.First(&policy, id).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 更新策略基本信息
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		err = tx.Model(&policy).Updates(updates).Error
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新策略失败: %v", err)
		}
	}

	// 更新策略规则关联
	if req.RuleIDs != nil {
		// 删除现有关联
		err = tx.Where("policy_id = ?", id).Delete(&models.PolicyRule{}).Error
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("删除现有策略规则关联失败: %v", err)
		}

		// 创建新关联
		for i, ruleID := range *req.RuleIDs {
			policyRule := models.PolicyRule{
				PolicyID: policy.ID,
				RuleID:   ruleID,
				Priority: i + 1,
				Enabled:  true,
			}

			err = tx.Create(&policyRule).Error
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("创建策略规则关联失败: %v", err)
			}
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	// 重新查询更新后的数据
	err = s.db.Preload("Tenant").First(&policy, id).Error
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

// DeletePolicy 删除策略
func (s *PolicyServiceV2) DeletePolicy(id uint) error {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除策略规则关联
	err := tx.Where("policy_id = ?", id).Delete(&models.PolicyRule{}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除策略规则关联失败: %v", err)
	}

	// 删除域名策略关联
	err = tx.Where("policy_id = ?", id).Delete(&models.DomainPolicy{}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除域名策略关联失败: %v", err)
	}

	// 删除策略
	err = tx.Delete(&models.Policy{}, id).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除策略失败: %v", err)
	}

	return tx.Commit().Error
}

// TogglePolicy 切换策略启用状态
func (s *PolicyServiceV2) TogglePolicy(id uint) (*models.Policy, error) {
	var policy models.Policy
	err := s.db.First(&policy, id).Error
	if err != nil {
		return nil, err
	}

	policy.Enabled = !policy.Enabled
	err = s.db.Save(&policy).Error
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

// GetAvailableRules 获取可用的规则列表
func (s *PolicyServiceV2) GetAvailableRules(tenantID uint) ([]models.Rule, error) {
	var rules []models.Rule

	// 获取全局规则和指定租户的规则
	err := s.db.Where("tenant_id = ? OR tenant_id = ?", 0, tenantID).
		Where("enabled = ?", true).
		Order("priority DESC, id DESC").
		Find(&rules).Error

	return rules, err
}

// BatchDeletePolicies 批量删除策略
func (s *PolicyServiceV2) BatchDeletePolicies(ids []uint) error {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除策略规则关联
	err := tx.Where("policy_id IN ?", ids).Delete(&models.PolicyRule{}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除策略规则关联失败: %v", err)
	}

	// 删除域名策略关联
	err = tx.Where("policy_id IN ?", ids).Delete(&models.DomainPolicy{}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除域名策略关联失败: %v", err)
	}

	// 删除策略
	err = tx.Where("id IN ?", ids).Delete(&models.Policy{}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("批量删除策略失败: %v", err)
	}

	return tx.Commit().Error
}

// ValidateRuleIDs 验证规则ID是否存在且属于指定租户
func (s *PolicyServiceV2) ValidateRuleIDs(ruleIDs []uint, tenantID uint) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	var count int64
	err := s.db.Model(&models.Rule{}).
		Where("id IN ? AND (tenant_id = ? OR tenant_id = ?)", ruleIDs, tenantID, 0).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("验证规则ID失败: %v", err)
	}

	if int(count) != len(ruleIDs) {
		return fmt.Errorf("部分规则不存在或无权限访问")
	}

	return nil
}

// GetPolicyRulesByPolicyID 获取策略的规则关联
func (s *PolicyServiceV2) GetPolicyRulesByPolicyID(policyID uint) ([]models.PolicyRule, error) {
	var policyRules []models.PolicyRule
	err := s.db.Where("policy_id = ?", policyID).
		Preload("Rule").
		Order("priority DESC").
		Find(&policyRules).Error

	return policyRules, err
}
