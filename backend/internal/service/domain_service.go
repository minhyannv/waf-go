package service

import (
	"errors"
	"fmt"
	"log"
	"waf-go/internal/models"
	"waf-go/internal/proxy"

	"gorm.io/gorm"
)

type DomainService struct {
	db              *gorm.DB
	securityService *TenantSecurityService
	proxyManager    *proxy.ProxyManager
}

func NewDomainService(db *gorm.DB, proxyManager *proxy.ProxyManager) *DomainService {
	return &DomainService{
		db:              db,
		securityService: NewTenantSecurityService(db),
		proxyManager:    proxyManager,
	}
}

// CreateDomainRequest 创建域名配置请求
type CreateDomainRequest struct {
	TenantID       uint   `json:"tenant_id"`
	Domain         string `json:"domain" binding:"required"`
	Protocol       string `json:"protocol" binding:"required,oneof=http https"`
	Port           int    `json:"port" binding:"required,min=1,max=65535"`
	SSLCertificate string `json:"ssl_certificate"`
	SSLPrivateKey  string `json:"ssl_private_key"`
	BackendURL     string `json:"backend_url" binding:"required"`
	Enabled        bool   `json:"enabled"`
}

// UpdateDomainRequest 更新域名配置请求
type UpdateDomainRequest struct {
	Domain         string `json:"domain"`
	Protocol       string `json:"protocol" binding:"omitempty,oneof=http https"`
	Port           int    `json:"port" binding:"omitempty,min=1,max=65535"`
	SSLCertificate string `json:"ssl_certificate"`
	SSLPrivateKey  string `json:"ssl_private_key"`
	BackendURL     string `json:"backend_url"`
	Enabled        *bool  `json:"enabled"`
}

// DomainListRequest 域名配置列表请求
type DomainListRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
	Domain   string `form:"domain"`
	Protocol string `form:"protocol"`
	Enabled  *bool  `form:"enabled"`
	TenantID uint   `form:"tenant_id"`
}

// DomainPolicyResponse 域名策略关联响应
type DomainPolicyResponse struct {
	ID       uint           `json:"id"`
	DomainID uint           `json:"domain_id"`
	PolicyID uint           `json:"policy_id"`
	Priority int            `json:"priority"`
	Enabled  bool           `json:"enabled"`
	Policy   *models.Policy `json:"policy,omitempty"`
}

// UpdateDomainPoliciesRequest 更新域名策略关联请求
type UpdateDomainPoliciesRequest struct {
	Policies []DomainPolicyItem `json:"policies" binding:"required"`
}

// DomainPolicyItem 域名策略关联项目
type DomainPolicyItem struct {
	PolicyID uint `json:"policy_id" binding:"required"`
	Priority int  `json:"priority"`
	Enabled  bool `json:"enabled"`
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// CreateDomain 创建域名配置
func (s *DomainService) CreateDomain(req *CreateDomainRequest) (*models.Domain, error) {
	// 检查域名是否已存在
	var existingDomain models.Domain
	err := s.db.Where("domain = ?", req.Domain).First(&existingDomain).Error
	if err == nil {
		return nil, fmt.Errorf("域名 %s 已存在", req.Domain)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查域名失败: %v", err)
	}

	// HTTPS协议验证
	if req.Protocol == "https" {
		if req.SSLCertificate == "" || req.SSLPrivateKey == "" {
			return nil, fmt.Errorf("HTTPS协议需要提供SSL证书和私钥")
		}
	}

	// 设置默认端口
	if req.Port == 0 {
		if req.Protocol == "https" {
			req.Port = 443
		} else {
			req.Port = 80
		}
	}

	domain := &models.Domain{
		TenantID:       req.TenantID,
		Domain:         req.Domain,
		Protocol:       req.Protocol,
		Port:           req.Port,
		SSLCertificate: req.SSLCertificate,
		SSLPrivateKey:  req.SSLPrivateKey,
		BackendURL:     req.BackendURL,
		Enabled:        req.Enabled,
	}

	if err := s.db.Create(domain).Error; err != nil {
		return nil, fmt.Errorf("创建域名失败: %v", err)
	}

	// 更新代理配置
	if domain.Enabled {
		if err := s.proxyManager.UpdateDomain(domain); err != nil {
			// 如果代理配置失败，回滚数据库
			s.db.Delete(domain)
			return nil, fmt.Errorf("配置代理失败: %v", err)
		}
	}

	return domain, nil
}

// GetDomains 获取域名列表（带租户隔离）
func (s *DomainService) GetDomains(req *DomainListRequest, userCtx *UserContext) ([]*models.Domain, int64, error) {
	var domains []*models.Domain
	var total int64

	// 构建查询，使用安全服务确保租户隔离
	query := s.db.Model(&models.Domain{})
	query = s.securityService.FilterByTenant(query, userCtx, false) // 不允许查看全局域名

	// 其他过滤条件
	if req.Domain != "" {
		query = query.Where("domain LIKE ?", "%"+req.Domain+"%")
	}
	if req.Protocol != "" {
		query = query.Where("protocol = ?", req.Protocol)
	}
	if req.Enabled != nil {
		query = query.Where("enabled = ?", *req.Enabled)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计域名数量失败: %v", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Preload("Tenant").
		Offset(offset).
		Limit(req.PageSize).
		Order("created_at DESC").
		Find(&domains).Error; err != nil {
		return nil, 0, fmt.Errorf("获取域名列表失败: %v", err)
	}

	return domains, total, nil
}

// GetDomain 获取单个域名详情
func (s *DomainService) GetDomain(id uint) (*models.Domain, error) {
	var domain models.Domain
	if err := s.db.Preload("Tenant").
		First(&domain, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("域名不存在")
		}
		return nil, fmt.Errorf("获取域名失败: %v", err)
	}

	return &domain, nil
}

// GetDomainWithSecurity 获取单个域名详情（带安全检查）
func (s *DomainService) GetDomainWithSecurity(id uint, userCtx *UserContext) (*models.Domain, error) {
	var domain models.Domain

	// 构建查询，使用安全服务确保租户隔离
	query := s.db.Model(&models.Domain{})
	query = s.securityService.FilterByTenant(query, userCtx, false)

	if err := query.Preload("Tenant").
		First(&domain, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("域名不存在")
		}
		return nil, fmt.Errorf("获取域名失败: %v", err)
	}

	return &domain, nil
}

// UpdateDomain 更新域名配置
func (s *DomainService) UpdateDomain(id uint, req *UpdateDomainRequest) (*models.Domain, error) {
	var domain models.Domain
	if err := s.db.First(&domain, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("域名不存在")
		}
		return nil, fmt.Errorf("获取域名失败: %v", err)
	}

	// 检查域名是否重复
	if req.Domain != "" && req.Domain != domain.Domain {
		var existingDomain models.Domain
		err := s.db.Where("domain = ? AND id != ?", req.Domain, id).First(&existingDomain).Error
		if err == nil {
			return nil, fmt.Errorf("域名 %s 已存在", req.Domain)
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("检查域名失败: %v", err)
		}
	}

	// HTTPS协议验证
	protocol := req.Protocol
	if protocol == "" {
		protocol = domain.Protocol
	}
	if protocol == "https" {
		sslCert := req.SSLCertificate
		sslKey := req.SSLPrivateKey
		if sslCert == "" {
			sslCert = domain.SSLCertificate
		}
		if sslKey == "" {
			sslKey = domain.SSLPrivateKey
		}
		if sslCert == "" || sslKey == "" {
			return nil, fmt.Errorf("HTTPS协议需要提供SSL证书和私钥")
		}
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Domain != "" {
		updates["domain"] = req.Domain
	}
	if req.Protocol != "" {
		updates["protocol"] = req.Protocol
		if req.Port == 0 {
			if req.Protocol == "https" {
				updates["port"] = 443
			} else {
				updates["port"] = 80
			}
		}
	}
	if req.Port > 0 {
		updates["port"] = req.Port
	}
	if req.SSLCertificate != "" {
		updates["ssl_certificate"] = req.SSLCertificate
	}
	if req.SSLPrivateKey != "" {
		updates["ssl_private_key"] = req.SSLPrivateKey
	}
	if req.BackendURL != "" {
		updates["backend_url"] = req.BackendURL
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := s.db.Model(&domain).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新域名失败: %v", err)
	}

	// 重新加载域名信息
	updatedDomain, err := s.GetDomain(id)
	if err != nil {
		return nil, err
	}

	// 更新代理配置
	if err := s.proxyManager.UpdateDomain(updatedDomain); err != nil {
		return nil, fmt.Errorf("更新代理配置失败: %v", err)
	}

	return updatedDomain, nil
}

// DeleteDomain 删除域名配置
func (s *DomainService) DeleteDomain(id uint) error {
	var domain models.Domain
	if err := s.db.First(&domain, id).Error; err != nil {
		return fmt.Errorf("域名不存在: %v", err)
	}

	// 移除代理配置
	s.proxyManager.RemoveDomain(domain.Domain)

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除域名策略关联
		if err := tx.Where("domain_id = ?", id).Delete(&models.DomainPolicy{}).Error; err != nil {
			return fmt.Errorf("删除域名策略关联失败: %v", err)
		}

		// 删除域名黑名单关联
		if err := tx.Where("domain_id = ?", id).Delete(&models.DomainBlackList{}).Error; err != nil {
			return fmt.Errorf("删除域名黑名单关联失败: %v", err)
		}

		// 删除域名白名单关联
		if err := tx.Where("domain_id = ?", id).Delete(&models.DomainWhiteList{}).Error; err != nil {
			return fmt.Errorf("删除域名白名单关联失败: %v", err)
		}

		// 删除域名
		if err := tx.Delete(&domain).Error; err != nil {
			return fmt.Errorf("删除域名失败: %v", err)
		}

		return nil
	})
}

// ToggleDomain 切换域名启用状态
func (s *DomainService) ToggleDomain(id uint) error {
	var domain models.Domain
	if err := s.db.First(&domain, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("域名不存在")
		}
		return fmt.Errorf("获取域名失败: %v", err)
	}

	newStatus := !domain.Enabled
	if err := s.db.Model(&domain).Update("enabled", newStatus).Error; err != nil {
		return fmt.Errorf("更新域名状态失败: %v", err)
	}

	// 更新代理配置
	domain.Enabled = newStatus
	if err := s.proxyManager.UpdateDomain(&domain); err != nil {
		return fmt.Errorf("更新代理配置失败: %v", err)
	}

	return nil
}

// GetDomainPolicies 获取域名关联的策略
func (s *DomainService) GetDomainPolicies(domainID uint) ([]DomainPolicyResponse, error) {
	var domainPolicies []models.DomainPolicy
	err := s.db.Where("domain_id = ?", domainID).
		Preload("Policy").Find(&domainPolicies).Error
	if err != nil {
		return nil, err
	}

	responses := make([]DomainPolicyResponse, len(domainPolicies))
	for i, dp := range domainPolicies {
		responses[i] = DomainPolicyResponse{
			ID:       dp.ID,
			DomainID: dp.DomainID,
			PolicyID: dp.PolicyID,
			Priority: dp.Priority,
			Enabled:  dp.Enabled,
			Policy:   dp.Policy,
		}
	}

	return responses, nil
}

// UpdateDomainPolicies 更新域名策略关联
func (s *DomainService) UpdateDomainPolicies(domainID uint, req *UpdateDomainPoliciesRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有关联
		if err := tx.Where("domain_id = ?", domainID).Delete(&models.DomainPolicy{}).Error; err != nil {
			return fmt.Errorf("删除现有策略关联失败: %v", err)
		}

		// 创建新关联
		for _, policy := range req.Policies {
			domainPolicy := models.DomainPolicy{
				DomainID: domainID,
				PolicyID: policy.PolicyID,
				Priority: policy.Priority,
				Enabled:  policy.Enabled,
			}
			if err := tx.Create(&domainPolicy).Error; err != nil {
				return fmt.Errorf("创建策略关联失败: %v", err)
			}
		}

		return nil
	})
}

// BatchDeleteDomains 批量删除域名配置
func (s *DomainService) BatchDeleteDomains(ids []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除相关联的策略关联
		if err := tx.Where("domain_id IN ?", ids).Delete(&models.DomainPolicy{}).Error; err != nil {
			return fmt.Errorf("删除域名策略关联失败: %v", err)
		}

		// 删除相关联的黑名单关联
		if err := tx.Where("domain_id IN ?", ids).Delete(&models.DomainBlackList{}).Error; err != nil {
			return fmt.Errorf("删除域名黑名单关联失败: %v", err)
		}

		// 删除相关联的白名单关联
		if err := tx.Where("domain_id IN ?", ids).Delete(&models.DomainWhiteList{}).Error; err != nil {
			return fmt.Errorf("删除域名白名单关联失败: %v", err)
		}

		// 删除域名配置
		if err := tx.Where("id IN ?", ids).Delete(&models.Domain{}).Error; err != nil {
			return fmt.Errorf("批量删除域名配置失败: %v", err)
		}

		return nil
	})
}

// LoadAllDomains 加载所有启用的域名到代理管理器
func (s *DomainService) LoadAllDomains() error {
	var domains []models.Domain
	if err := s.db.Where("enabled = ?", true).Find(&domains).Error; err != nil {
		return fmt.Errorf("加载域名失败: %v", err)
	}

	for _, domain := range domains {
		if err := s.proxyManager.UpdateDomain(&domain); err != nil {
			log.Printf("加载域名 %s 失败: %v", domain.Domain, err)
			continue
		}
	}

	return nil
}

// GetProxyManager 获取代理管理器
func (s *DomainService) GetProxyManager() *proxy.ProxyManager {
	return s.proxyManager
}

// GetDomainConfig 根据域名获取配置信息
func (s *DomainService) GetDomainConfig(domain string) *models.Domain {
	return s.proxyManager.GetDomainConfig(domain)
}

// HasDomainPolicies 检查域名是否关联了策略
func (s *DomainService) HasDomainPolicies(domain string) bool {
	domainConfig := s.GetDomainConfig(domain)
	if domainConfig == nil {
		return false
	}

	// 查询域名是否关联了策略
	var count int64
	err := s.db.Model(&models.DomainPolicy{}).
		Where("domain_id = ? AND enabled = ?", domainConfig.ID, true).
		Count(&count).Error

	if err != nil {
		return false
	}

	return count > 0
}
