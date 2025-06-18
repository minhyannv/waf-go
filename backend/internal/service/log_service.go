package service

import (
	"time"

	"waf-go/internal/models"

	"gorm.io/gorm"
)

type LogService struct {
	db *gorm.DB
}

type LogListRequest struct {
	Page       int       `form:"page,default=1"`
	PageSize   int       `form:"page_size,default=10"`
	ClientIP   string    `form:"client_ip"`
	RequestURI string    `form:"request_uri"`
	RuleName   string    `form:"rule_name"`
	Domain     string    `form:"domain"`
	Action     string    `form:"action"`
	TenantID   uint      `form:"tenant_id"`
	StartTime  time.Time `form:"start_time"`
	EndTime    time.Time `form:"end_time"`
}

func NewLogService(db *gorm.DB) *LogService {
	return &LogService{db: db}
}

// GetAttackLogList 获取攻击日志列表
func (s *LogService) GetAttackLogList(req *LogListRequest) ([]models.AttackLog, int64, error) {
	var logs []models.AttackLog
	var total int64

	query := s.db.Model(&models.AttackLog{})

	// 添加筛选条件
	if req.ClientIP != "" {
		query = query.Where("client_ip = ?", req.ClientIP)
	}
	if req.RequestURI != "" {
		query = query.Where("request_uri LIKE ?", "%"+req.RequestURI+"%")
	}
	if req.RuleName != "" {
		query = query.Where("rule_name LIKE ?", "%"+req.RuleName+"%")
	}
	if req.Domain != "" {
		query = query.Where("domain LIKE ?", "%"+req.Domain+"%")
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.TenantID > 0 {
		query = query.Where("tenant_id = ?", req.TenantID)
	}
	if !req.StartTime.IsZero() {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if !req.EndTime.IsZero() {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	err = query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetAttackLogByID 根据ID获取攻击日志详情
func (s *LogService) GetAttackLogByID(id uint) (*models.AttackLog, error) {
	var log models.AttackLog
	err := s.db.Preload("Rule").Preload("Tenant").First(&log, id).Error
	return &log, err
}

// DeleteAttackLog 删除攻击日志
func (s *LogService) DeleteAttackLog(id uint) error {
	return s.db.Delete(&models.AttackLog{}, id).Error
}

// DeleteAttackLogs 批量删除攻击日志
func (s *LogService) DeleteAttackLogs(ids []uint) error {
	return s.db.Where("id IN ?", ids).Delete(&models.AttackLog{}).Error
}

// ExportAttackLogs 导出攻击日志
func (s *LogService) ExportAttackLogs(ids []uint) ([]models.AttackLog, error) {
	var logs []models.AttackLog
	query := s.db // 移除Preload，只导出基本信息

	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// CleanOldLogs 清理旧日志（保留指定天数）
func (s *LogService) CleanOldLogs(days int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -days)

	result := s.db.Where("created_at < ?", cutoffTime).Delete(&models.AttackLog{})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// GetLogList 获取日志列表
func (s *LogService) GetLogList(query models.LogQuery) ([]models.AttackLog, int64, error) {
	var logs []models.AttackLog
	var total int64

	db := s.db.Model(&models.AttackLog{})

	// 应用查询条件
	if query.DomainID != "" {
		db = db.Where("domain_id = ?", query.DomainID)
	}
	if !query.StartTime.IsZero() {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if !query.EndTime.IsZero() {
		db = db.Where("created_at <= ?", query.EndTime)
	}
	if query.RequestPath != "" {
		db = db.Where("request_path LIKE ?", "%"+query.RequestPath+"%")
	}
	if query.ClientIP != "" {
		db = db.Where("client_ip = ?", query.ClientIP)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetLogDetail 获取日志详情
func (s *LogService) GetLogDetail(logID string) (*models.AttackLog, error) {
	var log models.AttackLog
	if err := s.db.First(&log, "id = ?", logID).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// DeleteLog 删除日志
func (s *LogService) DeleteLog(logID string) error {
	return s.db.Delete(&models.AttackLog{}, "id = ?", logID).Error
}

// BatchDeleteLogs 批量删除日志
func (s *LogService) BatchDeleteLogs(ids []string) error {
	return s.db.Delete(&models.AttackLog{}, "id IN ?", ids).Error
}
