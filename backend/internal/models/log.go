package models

import "time"

// LogQuery 日志查询参数
type LogQuery struct {
	DomainID    string    `form:"domain_id"`
	StartTime   time.Time `form:"start_time"`
	EndTime     time.Time `form:"end_time"`
	RequestPath string    `form:"request_path"`
	ClientIP    string    `form:"client_ip"`
	Page        int       `form:"page,default=1"`
	PageSize    int       `form:"page_size,default=10"`
}
