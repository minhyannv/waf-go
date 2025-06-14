package service

import (
	"waf-go/internal/models"
	"waf-go/internal/waf"

	"gorm.io/gorm"
)

type RuleService struct {
	db        *gorm.DB
	wafEngine *waf.Engine
}

type CreateRuleRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	MatchType    string `json:"match_type" binding:"required,oneof=uri ip header body user_agent"`
	Pattern      string `json:"pattern" binding:"required"`
	MatchMode    string `json:"match_mode" binding:"required,oneof=exact regex contains"`
	Action       string `json:"action" binding:"required,oneof=block allow log"`
	ResponseCode int    `json:"response_code"`
	ResponseMsg  string `json:"response_msg"`
	Priority     int    `json:"priority"`
	Enabled      bool   `json:"enabled"`
	TenantID     uint   `json:"tenant_id"`
}

type UpdateRuleRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	MatchType    string `json:"match_type" binding:"oneof=uri ip header body user_agent"`
	Pattern      string `json:"pattern"`
	MatchMode    string `json:"match_mode" binding:"oneof=exact regex contains"`
	Action       string `json:"action" binding:"oneof=block allow log"`
	ResponseCode int    `json:"response_code"`
	ResponseMsg  string `json:"response_msg"`
	Priority     int    `json:"priority"`
	Enabled      *bool  `json:"enabled"`
}

type RuleListRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
	Name     string `form:"name"`
	Enabled  *bool  `form:"enabled"`
	TenantID uint   `form:"tenant_id"`
}

func NewRuleService(db *gorm.DB, wafEngine *waf.Engine) *RuleService {
	return &RuleService{
		db:        db,
		wafEngine: wafEngine,
	}
}

// CreateRule 创建规则
func (s *RuleService) CreateRule(req *CreateRuleRequest) (*models.Rule, error) {
	rule := &models.Rule{
		Name:         req.Name,
		Description:  req.Description,
		MatchType:    req.MatchType,
		Pattern:      req.Pattern,
		MatchMode:    req.MatchMode,
		Action:       req.Action,
		ResponseCode: req.ResponseCode,
		ResponseMsg:  req.ResponseMsg,
		Priority:     req.Priority,
		Enabled:      req.Enabled,
		TenantID:     req.TenantID,
	}

	if rule.ResponseCode == 0 {
		rule.ResponseCode = 403
	}

	err := s.db.Create(rule).Error
	if err != nil {
		return nil, err
	}

	// 重新加载WAF规则
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
	err = query.Offset(offset).Limit(req.PageSize).Order("priority DESC, id DESC").Find(&rules).Error
	if err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// GetRuleByID 根据ID获取规则
func (s *RuleService) GetRuleByID(id uint) (*models.Rule, error) {
	var rule models.Rule
	err := s.db.First(&rule, id).Error
	return &rule, err
}

// UpdateRule 更新规则
func (s *RuleService) UpdateRule(id uint, req *UpdateRuleRequest) (*models.Rule, error) {
	var rule models.Rule
	err := s.db.First(&rule, id).Error
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.Description != "" {
		rule.Description = req.Description
	}
	if req.MatchType != "" {
		rule.MatchType = req.MatchType
	}
	if req.Pattern != "" {
		rule.Pattern = req.Pattern
	}
	if req.MatchMode != "" {
		rule.MatchMode = req.MatchMode
	}
	if req.Action != "" {
		rule.Action = req.Action
	}
	if req.ResponseCode > 0 {
		rule.ResponseCode = req.ResponseCode
	}
	if req.ResponseMsg != "" {
		rule.ResponseMsg = req.ResponseMsg
	}
	if req.Priority > 0 {
		rule.Priority = req.Priority
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}

	err = s.db.Save(&rule).Error
	if err != nil {
		return nil, err
	}

	// 重新加载WAF规则
	s.wafEngine.LoadRules()

	return &rule, nil
}

// DeleteRule 删除规则
func (s *RuleService) DeleteRule(id uint) error {
	err := s.db.Delete(&models.Rule{}, id).Error
	if err != nil {
		return err
	}

	// 重新加载WAF规则
	s.wafEngine.LoadRules()

	return nil
}

// ToggleRule 切换规则启用状态
func (s *RuleService) ToggleRule(id uint) error {
	var rule models.Rule
	err := s.db.First(&rule, id).Error
	if err != nil {
		return err
	}

	rule.Enabled = !rule.Enabled
	err = s.db.Save(&rule).Error
	if err != nil {
		return err
	}

	// 重新加载WAF规则
	s.wafEngine.LoadRules()

	return nil
}
