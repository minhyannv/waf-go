import request from '@/utils/request'

// 策略数据类型定义
export interface Policy {
  id?: number
  name: string
  description?: string
  domain?: string
  rule_ids: number[]
  enabled: boolean
  tenant_id?: number
  created_at?: string
  updated_at?: string
  toggling?: boolean // 用于UI状态控制
}

// 策略列表响应类型
export interface PolicyListResponse {
  list: Policy[]
  total: number
  page: number
  size: number
}

// 策略详情响应类型（包含关联规则）
export interface PolicyWithRules extends Policy {
  rules: Array<{
    id: number
    name: string
    description?: string
    match_type: string
    pattern: string
    match_mode: string
    action: string
    priority: number
    enabled: boolean
  }>
}

// 获取策略列表
export const getPolicyList = (params?: {
  page?: number
  page_size?: number
  name?: string
  domain?: string
  enabled?: boolean
}) => {
  return request.get<PolicyListResponse>('/policies', { params })
}

// 获取策略详情
export const getPolicyDetail = (id: number) => {
  return request.get<Policy>(`/policies/${id}`)
}

// 获取策略及其关联规则
export const getPolicyWithRules = (id: number) => {
  return request.get<PolicyWithRules>(`/policies/${id}/rules`)
}

// 创建策略
export const createPolicy = (data: Omit<Policy, 'id' | 'created_at' | 'updated_at'>) => {
  return request.post<Policy>('/policies', data)
}

// 更新策略
export const updatePolicy = (id: number, data: Partial<Policy>) => {
  return request.put<Policy>(`/policies/${id}`, data)
}

// 删除策略
export const deletePolicy = (id: number) => {
  return request.delete(`/policies/${id}`)
}

// 切换策略启用状态
export const togglePolicy = (id: number) => {
  return request.post(`/policies/${id}/toggle`)
}

// 批量删除策略
export const batchDeletePolicies = (ids: number[]) => {
  return request.delete('/policies/batch', { data: { ids } })
}

// 获取可用的规则列表
export const getAvailableRules = () => {
  return request.get<Array<{
    id: number
    name: string
    description?: string
    match_type: string
    pattern: string
    match_mode: string
    action: string
    priority: number
    enabled: boolean
  }>>('/policies/rules/available')
} 