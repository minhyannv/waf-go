import request from '@/utils/request'

// 攻击日志数据类型定义
export interface AttackLog {
  id: number
  request_id: string
  client_ip: string
  user_agent: string
  request_method: string
  request_uri: string
  request_headers: string
  request_body: string
  rule_id: number
  rule_name: string
  match_field: string
  match_value: string
  action: 'block' | 'allow' | 'log'
  response_code: number
  tenant_id: number
  created_at: string
  rule?: {
    id: number
    name: string
    description: string
    match_type: string
    pattern: string
    action: string
  }
}

export interface LogListParams {
  page?: number
  page_size?: number
  client_ip?: string
  request_uri?: string
  rule_name?: string
  action?: string
  start_time?: string
  end_time?: string
}

export interface LogListResponse {
  list: AttackLog[]
  total: number
  page: number
  page_size: number
}

// 获取攻击日志列表
export function getAttackLogs(params: AttackLogListParams) {
  return request.get<AttackLogListResponse>('/api/v1/logs/attacks', { params })
}

// 获取攻击日志详情
export function getAttackLogDetail(id: number) {
  return request.get<AttackLog>(`/api/v1/logs/attacks/${id}`)
}

// 删除攻击日志
export function deleteAttackLog(id: number) {
  return request.delete(`/api/v1/logs/attacks/${id}`)
}

// 批量删除攻击日志
export function batchDeleteAttackLogs(ids: number[]) {
  return request.delete('/api/v1/logs/attacks/batch', { data: { ids } })
}

// 清理旧日志
export const cleanOldLogs = (days: number) => {
  return request.post('/logs/attacks/clean', { days })
}

// 导出攻击日志
export const exportAttackLogs = (ids?: number[]) => {
  const params = new URLSearchParams()
  if (ids && ids.length > 0) {
    ids.forEach(id => params.append('ids', id.toString()))
  }
  
  return request.get('/logs/attacks/export', { 
    params,
    responseType: 'blob'
  })
}

// 获取实时日志列表
export function getRealtimeLogs(params: RealtimeLogListParams) {
  return request.get<RealtimeLogListResponse>('/api/v1/logs/realtime', { params })
}

// 获取实时日志详情
export function getRealtimeLogDetail(id: number) {
  return request.get<RealtimeLog>(`/api/v1/logs/realtime/${id}`)
}

// 删除实时日志
export function deleteRealtimeLog(id: number) {
  return request.delete(`/api/v1/logs/realtime/${id}`)
}

// 批量删除实时日志
export function batchDeleteRealtimeLogs(ids: number[]) {
  return request.delete('/api/v1/logs/realtime/batch', { data: { ids } })
} 