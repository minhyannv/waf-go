package service

import (
	"fmt"
	"waf-go/internal/models"
	"waf-go/internal/waf"

	"gorm.io/gorm"
)

type RuleService struct {
	db        *gorm.DB
	wafEngine *waf.WAFEngine
}

func NewRuleService(db *gorm.DB, wafEngine *waf.WAFEngine) *RuleService {
	return &RuleService{
		db:        db,
		wafEngine: wafEngine,
	}
}

// CreateRuleRequest 创建规则请求
type CreateRuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	MatchType   string `json:"match_type" binding:"required,oneof=uri ip header body user_agent"`
	Pattern     string `json:"pattern" binding:"required"`
	MatchMode   string `json:"match_mode" binding:"required,oneof=exact regex contains"`
	Action      string `json:"action" binding:"required,oneof=block log allow"`
	Priority    int    `json:"priority" binding:"required,min=1,max=1000"`
	Enabled     bool   `json:"enabled"`
	TenantID    uint   `json:"tenant_id"`
}

// UpdateRuleRequest 更新规则请求
type UpdateRuleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	MatchType   string `json:"match_type" binding:"omitempty,oneof=uri ip header body user_agent"`
	Pattern     string `json:"pattern"`
	MatchMode   string `json:"match_mode" binding:"omitempty,oneof=exact regex contains"`
	Action      string `json:"action" binding:"omitempty,oneof=block log allow"`
	Priority    *int   `json:"priority" binding:"omitempty,min=1,max=1000"`
	Enabled     *bool  `json:"enabled"`
}

// RuleListRequest 规则列表请求
type RuleListRequest struct {
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=10"`
	Name      string `form:"name"`
	Action    string `form:"action"`
	Enabled   *bool  `form:"enabled"`
	MatchType string `form:"match_type"`
	TenantID  uint   `form:"tenant_id"`
}

// CreateRule 创建规则
func (s *RuleService) CreateRule(req *CreateRuleRequest) (*models.Rule, error) {
	rule := &models.Rule{
		Name:        req.Name,
		Description: req.Description,
		MatchType:   req.MatchType,
		Pattern:     req.Pattern,
		MatchMode:   req.MatchMode,
		Action:      req.Action,
		Enabled:     req.Enabled,
		TenantID:    req.TenantID,
	}

	if err := s.db.Create(rule).Error; err != nil {
		return nil, fmt.Errorf("创建规则失败: %v", err)
	}

	// 通知WAF引擎重新加载规则
	s.wafEngine.LoadRules()

	return rule, nil
}

// GetRuleList 获取规则列表
func (s *RuleService) GetRuleList(req *RuleListRequest) ([]models.Rule, int64, error) {
	var rules []models.Rule
	var total int64

	query := s.db.Model(&models.Rule{})

	// 添加筛选条件
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.Enabled != nil {
		query = query.Where("enabled = ?", *req.Enabled)
	}
	if req.MatchType != "" {
		query = query.Where("match_type = ?", req.MatchType)
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
		Preload("Tenant").Find(&rules).Error
	if err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// GetRule 获取规则详情
func (s *RuleService) GetRule(id uint) (*models.Rule, error) {
	var rule models.Rule
	if err := s.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetRuleByID 根据ID获取规则（兼容旧方法名）
func (s *RuleService) GetRuleByID(id uint) (*models.Rule, error) {
	return s.GetRule(id)
}

// UpdateRule 更新规则
func (s *RuleService) UpdateRule(id uint, req *UpdateRuleRequest) (*models.Rule, error) {
	var rule models.Rule
	if err := s.db.First(&rule, id).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.MatchType != "" {
		updates["match_type"] = req.MatchType
	}
	if req.Pattern != "" {
		updates["pattern"] = req.Pattern
	}
	if req.MatchMode != "" {
		updates["match_mode"] = req.MatchMode
	}
	if req.Action != "" {
		updates["action"] = req.Action
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := s.db.Model(&rule).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 通知WAF引擎重新加载规则
	s.wafEngine.LoadRules()

	return &rule, nil
}

// DeleteRule 删除规则
func (s *RuleService) DeleteRule(id uint) error {
	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除策略规则关联
		if err := tx.Where("rule_id = ?", id).Delete(&models.PolicyRule{}).Error; err != nil {
			return fmt.Errorf("删除策略规则关联失败: %v", err)
		}

		// 删除规则
		if err := tx.Delete(&models.Rule{}, id).Error; err != nil {
			return fmt.Errorf("删除规则失败: %v", err)
		}

		// 通知WAF引擎重新加载规则
		s.wafEngine.LoadRules()

		return nil
	})
}

// ToggleRule 切换规则启用状态
func (s *RuleService) ToggleRule(id uint) error {
	var rule models.Rule
	if err := s.db.First(&rule, id).Error; err != nil {
		return err
	}

	rule.Enabled = !rule.Enabled
	if err := s.db.Save(&rule).Error; err != nil {
		return err
	}

	// 通知WAF引擎重新加载规则
	s.wafEngine.LoadRules()

	return nil
}

// BatchDeleteRules 批量删除规则
func (s *RuleService) BatchDeleteRules(ids []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除策略规则关联
		if err := tx.Where("rule_id IN ?", ids).Delete(&models.PolicyRule{}).Error; err != nil {
			return fmt.Errorf("删除策略规则关联失败: %v", err)
		}

		// 删除规则
		if err := tx.Where("id IN ?", ids).Delete(&models.Rule{}).Error; err != nil {
			return fmt.Errorf("删除规则失败: %v", err)
		}

		// 通知WAF引擎重新加载规则
		s.wafEngine.LoadRules()

		return nil
	})
}
