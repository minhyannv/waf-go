import request from '@/utils/request'

export interface DashboardStats {
  total_requests: number
  blocked_requests: number
  allowed_requests: number
  top_attack_ips: Array<{ ip: string; count: number }>
  top_attack_uris: Array<{ request_path: string; count: number }>
  top_attack_rules: Array<{ rule_name: string; count: number }>
  top_attack_user_agents: Array<{ user_agent: string; count: number }>
  hourly_stats: Array<{ hour: string; count: number }>
  daily_stats: Array<{ date: string; count: number }>
  active_rules: number
  active_policies: number
}

// 仪表盘概览数据
export interface DashboardOverview {
  total_attack_logs: number
  today_attack_logs: number
  blocked_requests: number
  passed_requests: number
  total_domains: number
  total_policies: number
  total_rules: number
  top_attack_domains: Array<{
    domain_id: string
    domain: string
    count: number
  }>
}

// 攻击趋势数据
export interface AttackTrend {
  time: string
  count: number
}

// 规则统计数据
export interface TopRule {
  rule_id: number
  rule_name: string
  description: string
  count: number
}

// IP统计数据
export interface TopIP {
  client_ip: string
  count: number
}

// URI统计数据
export interface TopURI {
  request_uri: string
  count: number
}

// User-Agent统计数据
export interface TopUserAgent {
  user_agent: string
  count: number
}

// API响应类型
export interface APIResponse<T> {
  code: number
  message: string
  data: T
}

// 获取仪表盘概览数据
export function getDashboardOverview(): Promise<APIResponse<DashboardOverview>> {
  return request({
    url: '/api/v1/dashboard/overview',
    method: 'get'
  })
}

// 获取攻击趋势数据
export function getAttackTrend(days: number = 7, timeType: 'hourly' | 'daily' = 'daily'): Promise<APIResponse<AttackTrend[]>> {
  return request({
    url: '/api/v1/dashboard/attack_trend',
    method: 'get',
    params: { days, time_type: timeType }
  })
}

// 获取 Top 规则数据
export function getTopRules(): Promise<APIResponse<TopRule[]>> {
  return request({
    url: '/api/v1/dashboard/top_rules',
    method: 'get'
  })
}

// 获取 Top IP 数据
export function getTopIPs(): Promise<APIResponse<TopIP[]>> {
  return request({
    url: '/api/v1/dashboard/top_ips',
    method: 'get'
  })
}

// 获取 Top URI 数据
export function getTopURIs(): Promise<APIResponse<TopURI[]>> {
  return request({
    url: '/api/v1/dashboard/top_uris',
    method: 'get'
  })
}

// 获取 Top User-Agent 数据
export function getTopUserAgents(): Promise<APIResponse<TopUserAgent[]>> {
  return request({
    url: '/api/v1/dashboard/top_user_agents',
    method: 'get'
  })
} 