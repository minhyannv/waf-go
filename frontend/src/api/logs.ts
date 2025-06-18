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
  domain_id: number
  domain: string
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
export function getAttackLogs(params: LogListParams) {
  return request.get<LogListResponse>('/api/v1/logs/attacks', { params })
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
  return request.post('/api/v1/logs/attacks/clean', { days })
}

// 导出攻击日志
export const exportAttackLogs = (ids?: number[]) => {
  const params = new URLSearchParams()
  if (ids && ids.length > 0) {
    ids.forEach(id => params.append('ids', id.toString()))
  }
  
  return request.get('/api/v1/logs/attacks/export', { 
    params,
    responseType: 'blob'
  })
}

