package service

import (
	"encoding/json"
	"fmt"
	"time"
	"waf-go/internal/models"

	"gorm.io/gorm"
)

type PolicyService struct {
	db *gorm.DB
}

type CreatePolicyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	RuleIDs     []uint `json:"rule_ids"`
	Enabled     bool   `json:"enabled"`
	TenantID    uint   `json:"tenant_id"`
}

type UpdatePolicyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	RuleIDs     []uint `json:"rule_ids"`
	Enabled     *bool  `json:"enabled"`
}

type PolicyListRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
	Name     string `form:"name"`
	Domain   string `form:"domain"`
	Enabled  *bool  `form:"enabled"`
	TenantID uint   `form:"tenant_id"`
}

type PolicyWithRules struct {
	models.Policy
	Rules []models.Rule `json:"rules"`
}

// PolicyResponse 策略响应结构体，将rule_ids转换为数组
type PolicyResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Domain      string    `json:"domain"`
	RuleIDs     []uint    `json:"rule_ids"`
	Enabled     bool      `json:"enabled"`
	TenantID    uint      `json:"tenant_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewPolicyService(db *gorm.DB) *PolicyService {
	return &PolicyService{db: db}
}

// convertToResponse 将Policy转换为PolicyResponse
func (s *PolicyService) convertToResponse(policy *models.Policy) (*PolicyResponse, error) {
	var ruleIDs []uint
	if policy.RuleIDs != "" {
		err := json.Unmarshal([]byte(policy.RuleIDs), &ruleIDs)
		if err != nil {
			return nil, fmt.Errorf("规则ID解析失败: %v", err)
		}
	}

	return &PolicyResponse{
		ID:          policy.ID,
		Name:        policy.Name,
		Description: policy.Description,
		Domain:      policy.Domain,
		RuleIDs:     ruleIDs,
		Enabled:     policy.Enabled,
		TenantID:    policy.TenantID,
		CreatedAt:   policy.CreatedAt,
		UpdatedAt:   policy.UpdatedAt,
	}, nil
}

// convertToResponseList 将Policy列表转换为PolicyResponse列表
func (s *PolicyService) convertToResponseList(policies []models.Policy) ([]PolicyResponse, error) {
	responses := make([]PolicyResponse, len(policies))
	for i, policy := range policies {
		response, err := s.convertToResponse(&policy)
		if err != nil {
			return nil, err
		}
		responses[i] = *response
	}
	return responses, nil
}

// CreatePolicy 创建策略
func (s *PolicyService) CreatePolicy(req *CreatePolicyRequest) (*PolicyResponse, error) {
	// 将规则ID数组转换为JSON字符串
	ruleIDsJSON, err := json.Marshal(req.RuleIDs)
	if err != nil {
		return nil, fmt.Errorf("规则ID序列化失败: %v", err)
	}

	policy := &models.Policy{
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
		RuleIDs:     string(ruleIDsJSON),
		Enabled:     req.Enabled,
		TenantID:    req.TenantID,
	}

	err = s.db.Create(policy).Error
	if err != nil {
		return nil, err
	}

	return s.convertToResponse(policy)
}

// GetPolicyList 获取策略列表
func (s *PolicyService) GetPolicyList(req *PolicyListRequest) ([]PolicyResponse, int64, error) {
	var policies []models.Policy
	var total int64

	query := s.db.Model(&models.Policy{})

	// 添加筛选条件
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Domain != "" {
		query = query.Where("domain LIKE ?", "%"+req.Domain+"%")
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
	err = query.Offset(offset).Limit(req.PageSize).Order("id DESC").Find(&policies).Error
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	responses, err := s.convertToResponseList(policies)
	if err != nil {
		return nil, 0, err
	}

	return responses, total, nil
}

// GetPolicyByID 根据ID获取策略
func (s *PolicyService) GetPolicyByID(id uint) (*PolicyResponse, error) {
	var policy models.Policy
	err := s.db.First(&policy, id).Error
	if err != nil {
		return nil, err
	}

	return s.convertToResponse(&policy)
}

// GetPolicyWithRules 获取策略及其关联的规则
func (s *PolicyService) GetPolicyWithRules(id uint) (*PolicyWithRules, error) {
	var policy models.Policy
	err := s.db.First(&policy, id).Error
	if err != nil {
		return nil, err
	}

	// 解析规则ID
	var ruleIDs []uint
	if policy.RuleIDs != "" {
		err = json.Unmarshal([]byte(policy.RuleIDs), &ruleIDs)
		if err != nil {
			return nil, fmt.Errorf("规则ID解析失败: %v", err)
		}
	}

	// 获取关联的规则
	var rules []models.Rule
	if len(ruleIDs) > 0 {
		err = s.db.Where("id IN ?", ruleIDs).Find(&rules).Error
		if err != nil {
			return nil, err
		}
	}

	return &PolicyWithRules{
		Policy: policy,
		Rules:  rules,
	}, nil
}

// UpdatePolicy 更新策略
func (s *PolicyService) UpdatePolicy(id uint, req *UpdatePolicyRequest) (*PolicyResponse, error) {
	var policy models.Policy
	err := s.db.First(&policy, id).Error
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		policy.Name = req.Name
	}
	if req.Description != "" {
		policy.Description = req.Description
	}
	if req.Domain != "" {
		policy.Domain = req.Domain
	}
	if req.RuleIDs != nil {
		ruleIDsJSON, err := json.Marshal(req.RuleIDs)
		if err != nil {
			return nil, fmt.Errorf("规则ID序列化失败: %v", err)
		}
		policy.RuleIDs = string(ruleIDsJSON)
	}
	if req.Enabled != nil {
		policy.Enabled = *req.Enabled
	}

	err = s.db.Save(&policy).Error
	if err != nil {
		return nil, err
	}

	return s.convertToResponse(&policy)
}

// DeletePolicy 删除策略
func (s *PolicyService) DeletePolicy(id uint) error {
	return s.db.Delete(&models.Policy{}, id).Error
}

// TogglePolicy 切换策略启用状态
func (s *PolicyService) TogglePolicy(id uint) error {
	var policy models.Policy
	err := s.db.First(&policy, id).Error
	if err != nil {
		return err
	}

	policy.Enabled = !policy.Enabled
	return s.db.Save(&policy).Error
}

// GetAvailableRules 获取可用的规则列表
func (s *PolicyService) GetAvailableRules(tenantID uint) ([]models.Rule, error) {
	var rules []models.Rule
	query := s.db.Model(&models.Rule{}).Where("enabled = ?", true)

	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}

	err := query.Order("priority DESC, name ASC").Find(&rules).Error
	return rules, err
}

// BatchDeletePolicies 批量删除策略
func (s *PolicyService) BatchDeletePolicies(ids []uint) error {
	return s.db.Where("id IN ?", ids).Delete(&models.Policy{}).Error
}

// GetPolicyRuleIDs 获取策略的规则ID列表
func (s *PolicyService) GetPolicyRuleIDs(policy *models.Policy) ([]uint, error) {
	var ruleIDs []uint
	if policy.RuleIDs != "" {
		err := json.Unmarshal([]byte(policy.RuleIDs), &ruleIDs)
		if err != nil {
			return nil, fmt.Errorf("规则ID解析失败: %v", err)
		}
	}
	return ruleIDs, nil
}

// ValidateRuleIDs 验证规则ID是否存在
func (s *PolicyService) ValidateRuleIDs(ruleIDs []uint, tenantID uint) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	var count int64
	query := s.db.Model(&models.Rule{}).Where("id IN ?", ruleIDs)
	if tenantID > 0 {
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return err
	}

	if int(count) != len(ruleIDs) {
		return fmt.Errorf("部分规则不存在或无权限访问")
	}

	return nil
}
