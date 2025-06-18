package service

import (
	"errors"

	"waf-go/internal/models"

	"gorm.io/gorm"
)

// BlackListService 黑名单服务
type BlackListService struct {
	db *gorm.DB
}

// NewBlackListService 创建黑名单服务实例
func NewBlackListService(db *gorm.DB) *BlackListService {
	return &BlackListService{db: db}
}

// CreateBlackList 创建黑名单
func (s *BlackListService) CreateBlackList(blackList *models.BlackList) error {
	// 检查是否已存在相同的类型、值和租户
	var existingBlackList models.BlackList
	err := s.db.Where("type = ? AND value = ? AND tenant_id = ?", blackList.Type, blackList.Value, blackList.TenantID).First(&existingBlackList).Error
	if err == nil {
		return errors.New("该黑名单条目已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return s.db.Create(blackList).Error
}

// GetBlackListByID 根据ID获取黑名单
func (s *BlackListService) GetBlackListByID(id uint) (*models.BlackList, error) {
	var blackList models.BlackList
	err := s.db.Preload("Tenant").First(&blackList, id).Error
	if err != nil {
		return nil, err
	}
	return &blackList, nil
}

// GetBlackLists 获取黑名单列表
func (s *BlackListService) GetBlackLists(tenantID uint, page, pageSize int, search string, listType string, value string, enabled *bool) ([]models.BlackList, int64, error) {
	var blackLists []models.BlackList
	var total int64

	query := s.db.Model(&models.BlackList{}).Where("tenant_id = ?", tenantID)

	// 搜索条件
	if search != "" {
		query = query.Where("value LIKE ? OR comment LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 类型筛选
	if listType != "" {
		query = query.Where("type = ?", listType)
	}

	// 值筛选
	if value != "" {
		query = query.Where("value LIKE ?", "%"+value+"%")
	}

	// 状态筛选
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Preload("Tenant").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&blackLists).Error

	return blackLists, total, err
}

// UpdateBlackList 更新黑名单
func (s *BlackListService) UpdateBlackList(blackList *models.BlackList) error {
	// 检查是否已存在相同的类型、值和租户（排除当前记录）
	var existingBlackList models.BlackList
	err := s.db.Where("type = ? AND value = ? AND tenant_id = ? AND id != ?", blackList.Type, blackList.Value, blackList.TenantID, blackList.ID).First(&existingBlackList).Error
	if err == nil {
		return errors.New("该黑名单条目已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return s.db.Save(blackList).Error
}

// DeleteBlackList 删除黑名单
func (s *BlackListService) DeleteBlackList(id uint) error {
	return s.db.Delete(&models.BlackList{}, id).Error
}

// IsIPBlacklisted 检查IP是否在黑名单中
func (s *BlackListService) IsIPBlacklisted(ipAddress string, tenantID uint) (bool, *models.BlackList, error) {
	var blackList models.BlackList
	err := s.db.Where("type = ? AND value = ? AND tenant_id = ? AND enabled = ?", "ip", ipAddress, tenantID, true).First(&blackList).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}

	return true, &blackList, nil
}

// GetBlacklistedIPs 获取租户的所有黑名单IP
func (s *BlackListService) GetBlacklistedIPs(tenantID uint) ([]string, error) {
	var ips []string
	err := s.db.Model(&models.BlackList{}).
		Where("type = ? AND tenant_id = ? AND enabled = ?", "ip", tenantID, true).
		Pluck("value", &ips).Error
	return ips, err
}

// ToggleBlackList 切换黑名单状态
func (s *BlackListService) ToggleBlackList(id uint) error {
	var blackList models.BlackList
	err := s.db.First(&blackList, id).Error
	if err != nil {
		return err
	}

	// 切换启用状态
	blackList.Enabled = !blackList.Enabled
	return s.db.Save(&blackList).Error
}
