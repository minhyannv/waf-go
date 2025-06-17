package service

import (
	"waf-go/internal/models"

	"gorm.io/gorm"
)

// TenantService 租户服务
type TenantService struct {
	db *gorm.DB
}

// NewTenantService 创建租户服务
func NewTenantService(db *gorm.DB) *TenantService {
	return &TenantService{
		db: db,
	}
}

// GetTenantByID 根据ID获取租户
func (s *TenantService) GetTenantByID(id uint) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.First(&tenant, id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

// CreateTenant 创建租户
func (s *TenantService) CreateTenant(tenant *models.Tenant) error {
	return s.db.Create(tenant).Error
}

// UpdateTenant 更新租户
func (s *TenantService) UpdateTenant(tenant *models.Tenant) error {
	return s.db.Save(tenant).Error
}

// DeleteTenant 删除租户
func (s *TenantService) DeleteTenant(id uint) error {
	return s.db.Delete(&models.Tenant{}, id).Error
}

// ListTenants 获取租户列表
func (s *TenantService) ListTenants(page, pageSize int) ([]models.Tenant, int64, error) {
	var tenants []models.Tenant
	var total int64

	if err := s.db.Model(&models.Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := s.db.Offset(offset).Limit(pageSize).Find(&tenants).Error; err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}
