package service

import (
	"waf-go/internal/models"

	"gorm.io/gorm"
)

type BlackListService struct {
	db *gorm.DB
}

type CreateBlackListRequest struct {
	Type     string `json:"type" binding:"required,oneof=ip uri user_agent"`
	Value    string `json:"value" binding:"required"`
	Comment  string `json:"comment"`
	Enabled  bool   `json:"enabled"`
	TenantID uint   `json:"tenant_id"`
}

type UpdateBlackListRequest struct {
	Type    string `json:"type" binding:"oneof=ip uri user_agent"`
	Value   string `json:"value"`
	Comment string `json:"comment"`
	Enabled *bool  `json:"enabled"`
}

type BlackListListRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
	Type     string `form:"type"`
	Value    string `form:"value"`
	Enabled  *bool  `form:"enabled"`
	TenantID uint   `form:"tenant_id"`
}

func NewBlackListService(db *gorm.DB) *BlackListService {
	return &BlackListService{db: db}
}

// CreateBlackList 创建黑名单
func (s *BlackListService) CreateBlackList(req *CreateBlackListRequest) (*models.BlackList, error) {
	blacklist := &models.BlackList{
		Type:     req.Type,
		Value:    req.Value,
		Comment:  req.Comment,
		Enabled:  req.Enabled,
		TenantID: req.TenantID,
	}

	if err := s.db.Create(blacklist).Error; err != nil {
		return nil, err
	}

	return blacklist, nil
}

// GetBlackListList 获取黑名单列表
func (s *BlackListService) GetBlackListList(req *BlackListListRequest) ([]models.BlackList, int64, error) {
	var blacklists []models.BlackList
	var total int64

	query := s.db.Model(&models.BlackList{})

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
		Find(&blacklists).Error; err != nil {
		return nil, 0, err
	}

	return blacklists, total, nil
}

// GetBlackListByID 根据ID获取黑名单
func (s *BlackListService) GetBlackListByID(id uint, tenantID uint) (*models.BlackList, error) {
	var blacklist models.BlackList
	query := s.db.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&blacklist).Error; err != nil {
		return nil, err
	}
	return &blacklist, nil
}

// UpdateBlackList 更新黑名单
func (s *BlackListService) UpdateBlackList(id uint, tenantID uint, req *UpdateBlackListRequest) (*models.BlackList, error) {
	blacklist, err := s.GetBlackListByID(id, tenantID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Type != "" {
		blacklist.Type = req.Type
	}
	if req.Value != "" {
		blacklist.Value = req.Value
	}
	if req.Comment != "" {
		blacklist.Comment = req.Comment
	}
	if req.Enabled != nil {
		blacklist.Enabled = *req.Enabled
	}

	if err := s.db.Save(blacklist).Error; err != nil {
		return nil, err
	}

	return blacklist, nil
}

// DeleteBlackList 删除黑名单
func (s *BlackListService) DeleteBlackList(id uint, tenantID uint) error {
	query := s.db.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	return query.Delete(&models.BlackList{}).Error
}

// BatchDeleteBlackList 批量删除黑名单
func (s *BlackListService) BatchDeleteBlackList(ids []uint, tenantID uint) error {
	query := s.db.Where("id IN ?", ids)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	return query.Delete(&models.BlackList{}).Error
}

// ToggleBlackListStatus 切换黑名单状态
func (s *BlackListService) ToggleBlackListStatus(id uint, tenantID uint) (*models.BlackList, error) {
	blacklist, err := s.GetBlackListByID(id, tenantID)
	if err != nil {
		return nil, err
	}

	blacklist.Enabled = !blacklist.Enabled
	if err := s.db.Save(blacklist).Error; err != nil {
		return nil, err
	}

	return blacklist, nil
}
