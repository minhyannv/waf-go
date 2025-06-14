import request from '@/utils/request'

export interface DashboardStats {
  total_requests: number
  blocked_requests: number
  allowed_requests: number
  top_attack_ips: Array<{ ip: string; count: number }>
  top_attack_uris: Array<{ uri: string; count: number }>
  top_attack_rules: Array<{ rule_name: string; count: number }>
  top_attack_user_agents: Array<{ user_agent: string; count: number }>
  hourly_stats: Array<{ hour: string; count: number }>
  daily_stats: Array<{ date: string; count: number }>
  active_rules: number
  active_policies: number
}

// 获取仪表盘统计数据
export function getDashboardStats(days: number = 7) {
  return request({
    url: '/dashboard/stats',
    method: 'get',
    params: { days }
  })
}

// 获取实时攻击统计数据
export function getRealtimeStats() {
  return request({
    url: '/dashboard/realtime',
    method: 'get'
  })
} 