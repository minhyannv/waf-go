package service

import (
	"waf-go/internal/models"

	"gorm.io/gorm"
)

type WhiteListService struct {
	db *gorm.DB
}

type CreateWhiteListRequest struct {
	Type     string `json:"type" binding:"required,oneof=ip uri user_agent"`
	Value    string `json:"value" binding:"required"`
	Comment  string `json:"comment"`
	Enabled  bool   `json:"enabled"`
	TenantID uint   `json:"tenant_id"`
}

type UpdateWhiteListRequest struct {
	Type    string `json:"type" binding:"oneof=ip uri user_agent"`
	Value   string `json:"value"`
	Comment string `json:"comment"`
	Enabled *bool  `json:"enabled"`
}

type WhiteListListRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
	Type     string `form:"type"`
	Value    string `form:"value"`
	Enabled  *bool  `form:"enabled"`
	TenantID uint   `form:"tenant_id"`
}

func NewWhiteListService(db *gorm.DB) *WhiteListService {
	return &WhiteListService{db: db}
}

// CreateWhiteList 创建白名单
func (s *WhiteListService) CreateWhiteList(req *CreateWhiteListRequest) (*models.WhiteList, error) {
	whitelist := &models.WhiteList{
		Type:     req.Type,
		Value:    req.Value,
		Comment:  req.Comment,
		Enabled:  req.Enabled,
		TenantID: req.TenantID,
	}

	if err := s.db.Create(whitelist).Error; err != nil {
		return nil, err
	}

	return whitelist, nil
}

// GetWhiteListList 获取白名单列表
func (s *WhiteListService) GetWhiteListList(req *WhiteListListRequest) ([]models.WhiteList, int64, error) {
	var whitelists []models.WhiteList
	var total int64

	query := s.db.Model(&models.WhiteList{})

	// 添加筛选条件
	if req.TenantID > 0 {
		query = query.Where("tenant_id = ?", req.TenantID)
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Value != "" {
		query = query.Where("value LIKE ?", "%"+req.Value+"%")
	}
	if req.Enabled != nil {
		query = query.Where("enabled = ?", *req.Enabled)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).
		Order("created_at DESC").
		Find(&whitelists).Error; err != nil {
		return nil, 0, err
	}

	return whitelists, total, nil
}

// GetWhiteListByID 根据ID获取白名单
func (s *WhiteListService) GetWhiteListByID(id uint, tenantID uint) (*models.WhiteList, error) {
	var whitelist models.WhiteList
	query := s.db.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&whitelist).Error; err != nil {
		return nil, err
	}
	return &whitelist, nil
}

// UpdateWhiteList 更新白名单
func (s *WhiteListService) UpdateWhiteList(id uint, tenantID uint, req *UpdateWhiteListRequest) (*models.WhiteList, error) {
	whitelist, err := s.GetWhiteListByID(id, tenantID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Type != "" {
		whitelist.Type = req.Type
	}
	if req.Value != "" {
		whitelist.Value = req.Value
	}
	if req.Comment != "" {
		whitelist.Comment = req.Comment
	}
	if req.Enabled != nil {
		whitelist.Enabled = *req.Enabled
	}

	if err := s.db.Save(whitelist).Error; err != nil {
		return nil, err
	}

	return whitelist, nil
}

// DeleteWhiteList 删除白名单
func (s *WhiteListService) DeleteWhiteList(id uint, tenantID uint) error {
	query := s.db.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	return query.Delete(&models.WhiteList{}).Error
}

// BatchDeleteWhiteList 批量删除白名单
func (s *WhiteListService) BatchDeleteWhiteList(ids []uint, tenantID uint) error {
	query := s.db.Where("id IN ?", ids)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	return query.Delete(&models.WhiteList{}).Error
}

// ToggleWhiteListStatus 切换白名单状态
func (s *WhiteListService) ToggleWhiteListStatus(id uint, tenantID uint) (*models.WhiteList, error) {
	whitelist, err := s.GetWhiteListByID(id, tenantID)
	if err != nil {
		return nil, err
	}

	whitelist.Enabled = !whitelist.Enabled
	if err := s.db.Save(whitelist).Error; err != nil {
		return nil, err
	}

	return whitelist, nil
}
