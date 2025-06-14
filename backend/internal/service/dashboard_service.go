package service

import (
	"time"

	"waf-go/internal/models"

	"gorm.io/gorm"
)

type DashboardService struct {
	db *gorm.DB
}

type DashboardStats struct {
	TotalRequests       int64              `json:"total_requests"`
	BlockedRequests     int64              `json:"blocked_requests"`
	AllowedRequests     int64              `json:"allowed_requests"`
	TopAttackIPs        []IPStat           `json:"top_attack_ips"`
	TopAttackURIs       []URIStat          `json:"top_attack_uris"`
	TopAttackRules      []RuleStat         `json:"top_attack_rules"`
	TopAttackUserAgents []UserAgentStat    `json:"top_attack_user_agents"`
	HourlyStats         []HourlyAttackStat `json:"hourly_stats"`
	DailyStats          []DailyAttackStat  `json:"daily_stats"`
	ActiveRules         int64              `json:"active_rules"`
	ActivePolicies      int64              `json:"active_policies"`
}

type IPStat struct {
	IP    string `json:"ip"`
	Count int64  `json:"count"`
}

type URIStat struct {
	URI        string `json:"uri"`
	Count      int64  `json:"count"`
	AttackType string `json:"attack_type"`
}

type RuleStat struct {
	RuleName string `json:"rule_name"`
	Count    int64  `json:"count"`
}

type UserAgentStat struct {
	UserAgent string `json:"user_agent"`
	Count     int64  `json:"count"`
}

type HourlyAttackStat struct {
	Hour  string `json:"hour"`
	Count int64  `json:"count"`
}

type DailyAttackStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// GetDashboardStats 获取仪表盘统计数据
func (s *DashboardService) GetDashboardStats(tenantID uint, days int) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 设置时间范围
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	// 查询条件
	query := s.db.Model(&models.AttackLog{}).Where("created_at >= ? AND created_at <= ?", startTime, endTime)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// 总请求数（攻击日志数）
	err := query.Count(&stats.TotalRequests).Error
	if err != nil {
		return nil, err
	}

	// 被阻止的请求数
	err = query.Where("action = ?", "block").Count(&stats.BlockedRequests).Error
	if err != nil {
		return nil, err
	}

	// 允许的请求数（记录但未阻止）
	stats.AllowedRequests = stats.TotalRequests - stats.BlockedRequests

	// 获取Top攻击IP
	topIPs, err := s.getTopAttackIPs(tenantID, startTime, endTime, 10)
	if err != nil {
		return nil, err
	}
	stats.TopAttackIPs = topIPs

	// 获取Top攻击URI
	topURIs, err := s.getTopAttackURIs(tenantID, startTime, endTime, 10)
	if err != nil {
		return nil, err
	}
	stats.TopAttackURIs = topURIs

	// 获取Top攻击规则
	topRules, err := s.getTopAttackRules(tenantID, startTime, endTime, 10)
	if err != nil {
		return nil, err
	}
	stats.TopAttackRules = topRules

	// 获取Top攻击User-Agent
	topUserAgents, err := s.getTopAttackUserAgents(tenantID, startTime, endTime, 10)
	if err != nil {
		return nil, err
	}
	stats.TopAttackUserAgents = topUserAgents

	// 获取每小时统计（最近24小时）
	hourlyStats, err := s.getHourlyStats(tenantID, 24)
	if err != nil {
		return nil, err
	}
	stats.HourlyStats = hourlyStats

	// 获取每日统计
	dailyStats, err := s.getDailyStats(tenantID, days)
	if err != nil {
		return nil, err
	}
	stats.DailyStats = dailyStats

	// 活跃规则数
	ruleQuery := s.db.Model(&models.Rule{}).Where("enabled = ?", true)
	if tenantID > 0 {
		ruleQuery = ruleQuery.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	err = ruleQuery.Count(&stats.ActiveRules).Error
	if err != nil {
		return nil, err
	}

	// 活跃策略数
	policyQuery := s.db.Model(&models.Policy{}).Where("enabled = ?", true)
	if tenantID > 0 {
		policyQuery = policyQuery.Where("tenant_id = ? OR tenant_id = 0", tenantID)
	}
	err = policyQuery.Count(&stats.ActivePolicies).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// getTopAttackIPs 获取Top攻击IP
func (s *DashboardService) getTopAttackIPs(tenantID uint, startTime, endTime time.Time, limit int) ([]IPStat, error) {
	var results []IPStat

	query := s.db.Model(&models.AttackLog{}).
		Select("client_ip as ip, count(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("client_ip").
		Order("count DESC").
		Limit(limit)

	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	err := query.Scan(&results).Error
	return results, err
}

// getTopAttackURIs 获取Top攻击URI
func (s *DashboardService) getTopAttackURIs(tenantID uint, startTime, endTime time.Time, limit int) ([]URIStat, error) {
	var results []URIStat

	// 首先获取Top URI
	type uriCount struct {
		URI   string `json:"uri"`
		Count int64  `json:"count"`
	}
	var uriCounts []uriCount

	query := s.db.Model(&models.AttackLog{}).
		Select("request_uri as uri, count(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("request_uri").
		Order("count DESC").
		Limit(limit)

	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	err := query.Scan(&uriCounts).Error
	if err != nil {
		return nil, err
	}

	// 为每个URI获取最常触发的规则名称作为攻击类型
	for _, uriCount := range uriCounts {
		var topRule string
		ruleQuery := s.db.Model(&models.AttackLog{}).
			Select("rule_name").
			Where("request_uri = ? AND created_at >= ? AND created_at <= ?", uriCount.URI, startTime, endTime).
			Group("rule_name").
			Order("count(*) DESC").
			Limit(1)

		if tenantID > 0 {
			ruleQuery = ruleQuery.Where("tenant_id = ?", tenantID)
		}

		ruleQuery.Scan(&topRule)

		results = append(results, URIStat{
			URI:        uriCount.URI,
			Count:      uriCount.Count,
			AttackType: topRule,
		})
	}

	return results, err
}

// getTopAttackRules 获取Top攻击规则
func (s *DashboardService) getTopAttackRules(tenantID uint, startTime, endTime time.Time, limit int) ([]RuleStat, error) {
	var results []RuleStat

	query := s.db.Model(&models.AttackLog{}).
		Select("rule_name, count(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("rule_name").
		Order("count DESC").
		Limit(limit)

	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	err := query.Scan(&results).Error
	return results, err
}

// getTopAttackUserAgents 获取Top攻击User-Agent
func (s *DashboardService) getTopAttackUserAgents(tenantID uint, startTime, endTime time.Time, limit int) ([]UserAgentStat, error) {
	var results []UserAgentStat

	query := s.db.Model(&models.AttackLog{}).
		Select("user_agent, count(*) as count").
		Where("created_at >= ? AND created_at <= ? AND user_agent != ''", startTime, endTime).
		Group("user_agent").
		Order("count DESC").
		Limit(limit)

	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	err := query.Scan(&results).Error
	return results, err
}

// getHourlyStats 获取每小时统计
func (s *DashboardService) getHourlyStats(tenantID uint, hours int) ([]HourlyAttackStat, error) {
	var results []HourlyAttackStat

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	query := s.db.Model(&models.AttackLog{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00') as hour, count(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("hour").
		Order("hour ASC")

	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	err := query.Scan(&results).Error
	return results, err
}

// getDailyStats 获取每日统计
func (s *DashboardService) getDailyStats(tenantID uint, days int) ([]DailyAttackStat, error) {
	var results []DailyAttackStat

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	query := s.db.Model(&models.AttackLog{}).
		Select("DATE(created_at) as date, count(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("date").
		Order("date ASC")

	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	err := query.Scan(&results).Error
	return results, err
}

// GetRealtimeStats 获取实时攻击统计（最近1小时，按分钟统计）
func (s *DashboardService) GetRealtimeStats(tenantID uint) ([]RealtimeAttackStat, error) {
	var results []RealtimeAttackStat

	// 使用本地时间以匹配数据库时区
	endTime := time.Now()
	startTime := endTime.Add(-time.Hour) // 最近1小时

	query := s.db.Model(&models.AttackLog{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:00') as minute, count(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("minute").
		Order("minute ASC")

	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 填充缺失的分钟数据（补0）
	fullResults := make([]RealtimeAttackStat, 0, 60)
	resultMap := make(map[string]int64)

	for _, result := range results {
		resultMap[result.Minute] = result.Count
	}

	// 生成最近60分钟的完整数据
	for i := 59; i >= 0; i-- {
		minute := endTime.Add(-time.Duration(i) * time.Minute).Format("2006-01-02 15:04:00")
		count := resultMap[minute]
		fullResults = append(fullResults, RealtimeAttackStat{
			Minute: minute,
			Count:  count,
		})
	}

	return fullResults, nil
}

type RealtimeAttackStat struct {
	Minute string `json:"minute"`
	Count  int64  `json:"count"`
}
