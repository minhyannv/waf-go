import request from '@/utils/request'

// 系统配置数据类型定义
export interface SystemConfig {
  waf: WAFConfig
  server: ServerConfig
  log: LogConfig
}

export interface WAFConfig {
  rate_limit_window: number
  max_requests: number
  enable_rate_limit: boolean
  enable_blacklist: boolean
  enable_whitelist: boolean
}

export interface ServerConfig {
  mode: string
}

export interface LogConfig {
  level: string
}

export interface UpdateConfigRequest {
  waf?: Partial<WAFConfig>
  server?: Partial<ServerConfig>
  log?: Partial<LogConfig>
}

// 获取系统配置
export function getSystemConfig() {
  return request({
    url: '/api/v1/config',
    method: 'get'
  })
}

// 更新系统配置
export function updateSystemConfig(data: UpdateConfigRequest) {
  return request({
    url: '/api/v1/config',
    method: 'put',
    data
  })
}

// 重置系统配置
export function resetSystemConfig() {
  return request({
    url: '/api/v1/config/reset',
    method: 'post'
  })
}

// 获取配置统计信息
export function getConfigStats() {
  return request({
    url: '/api/v1/config/stats',
    method: 'get'
  })
} 