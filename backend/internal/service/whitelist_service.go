package service

import (
	"errors"

	"waf-go/internal/models"

	"gorm.io/gorm"
)

// WhiteListService 白名单服务
type WhiteListService struct {
	db *gorm.DB
}

// NewWhiteListService 创建白名单服务实例
func NewWhiteListService(db *gorm.DB) *WhiteListService {
	return &WhiteListService{db: db}
}

// CreateWhiteList 创建白名单
func (s *WhiteListService) CreateWhiteList(whiteList *models.WhiteList) error {
	// 检查是否已存在相同的类型、值和租户
	var existingWhiteList models.WhiteList
	err := s.db.Where("type = ? AND value = ? AND tenant_id = ?", whiteList.Type, whiteList.Value, whiteList.TenantID).First(&existingWhiteList).Error
	if err == nil {
		return errors.New("该白名单条目已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return s.db.Create(whiteList).Error
}

// GetWhiteListByID 根据ID获取白名单
func (s *WhiteListService) GetWhiteListByID(id uint) (*models.WhiteList, error) {
	var whiteList models.WhiteList
	err := s.db.Preload("Tenant").First(&whiteList, id).Error
	if err != nil {
		return nil, err
	}
	return &whiteList, nil
}

// GetWhiteLists 获取白名单列表
func (s *WhiteListService) GetWhiteLists(tenantID uint, page, pageSize int, search string, listType string, value string, enabled *bool) ([]models.WhiteList, int64, error) {
	var whiteLists []models.WhiteList
	var total int64

	query := s.db.Model(&models.WhiteList{}).Where("tenant_id = ?", tenantID)

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
		Find(&whiteLists).Error

	return whiteLists, total, err
}

// UpdateWhiteList 更新白名单
func (s *WhiteListService) UpdateWhiteList(whiteList *models.WhiteList) error {
	// 检查是否已存在相同的类型、值和租户（排除当前记录）
	var existingWhiteList models.WhiteList
	err := s.db.Where("type = ? AND value = ? AND tenant_id = ? AND id != ?", whiteList.Type, whiteList.Value, whiteList.TenantID, whiteList.ID).First(&existingWhiteList).Error
	if err == nil {
		return errors.New("该白名单条目已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return s.db.Save(whiteList).Error
}

// DeleteWhiteList 删除白名单
func (s *WhiteListService) DeleteWhiteList(id uint) error {
	return s.db.Delete(&models.WhiteList{}, id).Error
}

// IsIPWhitelisted 检查IP是否在白名单中
func (s *WhiteListService) IsIPWhitelisted(ipAddress string, tenantID uint) (bool, *models.WhiteList, error) {
	var whiteList models.WhiteList
	err := s.db.Where("type = ? AND value = ? AND tenant_id = ? AND enabled = ?", "ip", ipAddress, tenantID, true).First(&whiteList).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}

	return true, &whiteList, nil
}

// GetWhitelistedIPs 获取租户的所有白名单IP
func (s *WhiteListService) GetWhitelistedIPs(tenantID uint) ([]string, error) {
	var ips []string
	err := s.db.Model(&models.WhiteList{}).
		Where("type = ? AND tenant_id = ? AND enabled = ?", "ip", tenantID, true).
		Pluck("value", &ips).Error
	return ips, err
}

// ToggleWhiteList 切换白名单状态
func (s *WhiteListService) ToggleWhiteList(id uint) error {
	var whiteList models.WhiteList
	err := s.db.First(&whiteList, id).Error
	if err != nil {
		return err
	}

	// 切换启用状态
	whiteList.Enabled = !whiteList.Enabled
	return s.db.Save(&whiteList).Error
}
