package service

import (
	"fmt"
	"time"
	"waf-go/internal/logger"
	"waf-go/internal/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PolicyService struct {
	db *gorm.DB
}

type CreatePolicyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	DomainID    *uint  `json:"domain_id"`
	RuleIDs     []uint `json:"rule_ids"`
	Enabled     bool   `json:"enabled"`
	TenantID    uint   `json:"tenant_id"`
}

type UpdatePolicyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DomainID    *uint  `json:"domain_id"`
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
	Policy PolicyResponse `json:"policy"`
	Rules  []models.Rule  `json:"rules"`
}

// PolicyResponse 策略响应结构体，将rule_ids转换为数组
type PolicyResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Domain      string    `json:"domain,omitempty"`
	DomainID    *uint     `json:"domain_id,omitempty"`
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
	// 获取策略关联的规则ID（通过多对多关联表）
	var ruleIDs []uint
	err := s.db.Table("policy_rules").
		Select("rule_id").
		Where("policy_id = ? AND enabled = ?", policy.ID, true).
		Pluck("rule_id", &ruleIDs).Error
	if err != nil {
		return nil, fmt.Errorf("获取策略规则失败: %v", err)
	}

	// 获取域名信息（通过domain_policies关联表）
	var domainInfo struct {
		DomainID uint   `gorm:"column:domain_id"`
		Domain   string `gorm:"column:domain"`
	}

	domainErr := s.db.Raw(`
		SELECT dp.domain_id, d.domain 
		FROM domain_policies dp 
		LEFT JOIN domains d ON dp.domain_id = d.id 
		WHERE dp.policy_id = ? AND dp.enabled = 1
	`, policy.ID).Scan(&domainInfo).Error

	response := &PolicyResponse{
		ID:          policy.ID,
		Name:        policy.Name,
		Description: policy.Description,
		RuleIDs:     ruleIDs,
		Enabled:     policy.Enabled,
		TenantID:    policy.TenantID,
		CreatedAt:   policy.CreatedAt,
		UpdatedAt:   policy.UpdatedAt,
	}

	// 如果找到域名信息，则添加到响应中
	if domainErr == nil && domainInfo.DomainID > 0 {
		response.DomainID = &domainInfo.DomainID
		response.Domain = domainInfo.Domain
	}

	return response, nil
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
	// 使用事务创建策略和关联规则
	var policy *models.Policy
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 创建策略
		policy = &models.Policy{
			Name:        req.Name,
			Description: req.Description,
			Enabled:     req.Enabled,
			TenantID:    req.TenantID,
		}

		if err := tx.Create(policy).Error; err != nil {
			return err
		}

		// 创建策略规则关联
		if len(req.RuleIDs) > 0 {
			for _, ruleID := range req.RuleIDs {
				policyRule := &models.PolicyRule{
					PolicyID: policy.ID,
					RuleID:   ruleID,
					Priority: 1,
					Enabled:  true,
				}
				if err := tx.Create(policyRule).Error; err != nil {
					return err
				}
			}
		}

		// 创建域名策略关联
		if req.DomainID != nil {
			fmt.Printf("DomainID: %d\n", *req.DomainID)
			if *req.DomainID > 0 {
				domainPolicy := &models.DomainPolicy{
					DomainID: *req.DomainID,
					PolicyID: policy.ID,
					Priority: 1,
					Enabled:  true,
				}
				if err := tx.Create(domainPolicy).Error; err != nil {
					return fmt.Errorf("创建域名策略关联失败: %v", err)
				}
				fmt.Printf("域名策略关联创建成功: DomainID=%d, PolicyID=%d\n", *req.DomainID, policy.ID)
			}
		} else {
			fmt.Println("DomainID is nil")
		}

		return nil
	})

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
		// 通过domain_policies和domains表做关联过滤
		query = query.Joins("JOIN domain_policies dp ON dp.policy_id = policies.id AND dp.enabled = 1").
			Joins("JOIN domains d ON dp.domain_id = d.id").
			Where("d.domain = ?", req.Domain)
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

	// 获取关联的规则（通过多对多关联表）
	var rules []models.Rule
	err = s.db.Table("rules").
		Select("rules.*").
		Joins("JOIN policy_rules ON rules.id = policy_rules.rule_id").
		Where("policy_rules.policy_id = ? AND policy_rules.enabled = ?", policy.ID, true).
		Find(&rules).Error
	if err != nil {
		return nil, err
	}

	// 获取域名信息（通过domain_policies关联表）
	var domainInfo struct {
		DomainID uint   `gorm:"column:domain_id"`
		Domain   string `gorm:"column:domain"`
	}

	domainErr := s.db.Raw(`
		SELECT dp.domain_id, d.domain 
		FROM domain_policies dp 
		LEFT JOIN domains d ON dp.domain_id = d.id 
		WHERE dp.policy_id = ? AND dp.enabled = 1
	`, policy.ID).Scan(&domainInfo).Error

	// 创建包含域名信息的策略响应
	policyResponse := &PolicyResponse{
		ID:          policy.ID,
		Name:        policy.Name,
		Description: policy.Description,
		RuleIDs:     []uint{}, // 将在下面填充
		Enabled:     policy.Enabled,
		TenantID:    policy.TenantID,
		CreatedAt:   policy.CreatedAt,
		UpdatedAt:   policy.UpdatedAt,
	}

	// 如果找到域名信息，则添加到响应中
	if domainErr == nil && domainInfo.DomainID > 0 {
		policyResponse.DomainID = &domainInfo.DomainID
		policyResponse.Domain = domainInfo.Domain
	}

	// 获取规则ID列表
	var ruleIDs []uint
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.ID)
	}
	policyResponse.RuleIDs = ruleIDs

	return &PolicyWithRules{
		Policy: *policyResponse,
		Rules:  rules,
	}, nil
}

// UpdatePolicy 更新策略
func (s *PolicyService) UpdatePolicy(id uint, req *UpdatePolicyRequest) (*PolicyResponse, error) {
	var result *PolicyResponse
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var policy models.Policy
		err := tx.First(&policy, id).Error
		if err != nil {
			return err
		}

		// 更新策略基本信息
		updates := make(map[string]interface{})
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}
		if req.Enabled != nil {
			updates["enabled"] = *req.Enabled
		}

		if len(updates) > 0 {
			err = tx.Model(&policy).Updates(updates).Error
			if err != nil {
				return err
			}
		}

		// 更新规则关联
		if req.RuleIDs != nil {
			// 删除现有关联
			err = tx.Where("policy_id = ?", id).Delete(&models.PolicyRule{}).Error
			if err != nil {
				return err
			}

			// 创建新关联
			for _, ruleID := range req.RuleIDs {
				policyRule := &models.PolicyRule{
					PolicyID: policy.ID,
					RuleID:   ruleID,
					Priority: 1,
					Enabled:  true,
				}
				if err := tx.Create(policyRule).Error; err != nil {
					return err
				}
			}
		}

		// 更新域名关联
		if req.DomainID != nil {
			// 删除现有域名关联
			err = tx.Where("policy_id = ?", id).Delete(&models.DomainPolicy{}).Error
			if err != nil {
				return err
			}

			// 如果指定了新的域名ID，创建新关联
			if *req.DomainID > 0 {
				domainPolicy := &models.DomainPolicy{
					DomainID: *req.DomainID,
					PolicyID: policy.ID,
					Priority: 1,
					Enabled:  true,
				}
				if err := tx.Create(domainPolicy).Error; err != nil {
					return err
				}
			}
		}

		// 获取更新后的策略响应
		result, err = s.convertToResponse(&policy)
		return err
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeletePolicy 删除策略
func (s *PolicyService) DeletePolicy(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除策略规则关联
		if err := tx.Where("policy_id = ?", id).Delete(&models.PolicyRule{}).Error; err != nil {
			return err
		}

		// 删除域名策略关联
		if err := tx.Where("policy_id = ?", id).Delete(&models.DomainPolicy{}).Error; err != nil {
			return err
		}

		// 删除策略
		return tx.Delete(&models.Policy{}, id).Error
	})
}

// TogglePolicy 切换策略启用状态
func (s *PolicyService) TogglePolicy(id uint) error {
	var policy models.Policy
	err := s.db.First(&policy, id).Error
	if err != nil {
		return err
	}

	newStatus := !policy.Enabled
	return s.db.Model(&policy).Update("enabled", newStatus).Error
}

// GetAvailableRules 获取可用的规则列表
func (s *PolicyService) GetAvailableRules(tenantID uint, role string) ([]models.Rule, error) {
	var rules []models.Rule
	query := s.db.Where("enabled = ?", true)

	// 强制日志输出，用于调试
	fmt.Printf("DEBUG: GetAvailableRules called with role='%s', tenantID=%d\n", role, tenantID)

	// 如果是超级管理员，可以看到所有规则
	if role == "admin" {
		// 超级管理员可以看到所有规则，不添加额外的where条件
		fmt.Printf("DEBUG: Admin branch taken\n")
		logger.Debug("GetAvailableRules - Admin user accessing all rules", zap.String("role", role), zap.Uint("tenant_id", tenantID))
		// 不添加任何tenant_id限制，让admin看到所有规则
	} else {
		// 普通用户只能看到自己租户的规则和全局规则
		fmt.Printf("DEBUG: Regular user branch taken\n")
		logger.Debug("GetAvailableRules - Regular user accessing limited rules", zap.String("role", role), zap.Uint("tenant_id", tenantID))
		query = query.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}

	err := query.Find(&rules).Error
	fmt.Printf("DEBUG: Found %d rules\n", len(rules))
	logger.Debug("GetAvailableRules - Query result", zap.Int("rules_count", len(rules)), zap.String("role", role))
	return rules, err
}

// BatchDeletePolicies 批量删除策略
func (s *PolicyService) BatchDeletePolicies(ids []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除策略规则关联
		if err := tx.Where("policy_id IN ?", ids).Delete(&models.PolicyRule{}).Error; err != nil {
			return err
		}

		// 删除域名策略关联
		if err := tx.Where("policy_id IN ?", ids).Delete(&models.DomainPolicy{}).Error; err != nil {
			return err
		}

		// 删除策略
		return tx.Where("id IN ?", ids).Delete(&models.Policy{}).Error
	})
}

// GetPolicyRules 获取策略规则列表
func (s *PolicyService) GetPolicyRules(policyID uint) ([]models.Rule, error) {
	var rules []models.Rule
	if err := s.db.Table("rules").
		Joins("JOIN policy_rules ON rules.id = policy_rules.rule_id").
		Where("policy_rules.policy_id = ?", policyID).
		Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// UpdatePolicyRules 更新策略规则
func (s *PolicyService) UpdatePolicyRules(policyID uint, ruleIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧的关联
		if err := tx.Where("policy_id = ?", policyID).Delete(&models.PolicyRule{}).Error; err != nil {
			return err
		}

		// 添加新的关联
		for _, ruleID := range ruleIDs {
			policyRule := models.PolicyRule{
				PolicyID: policyID,
				RuleID:   ruleID,
			}
			if err := tx.Create(&policyRule).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
