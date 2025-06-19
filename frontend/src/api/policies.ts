import request from '@/utils/request'

// 策略配置
export interface PolicyConfig {
  id: number
  name: string
  description?: string
  enabled: boolean
  created_at?: string
  updated_at?: string
  tenant_id?: number
  domain?: string
  domain_id?: number
  rule_ids?: number[]
  toggling?: boolean
  tenant?: {
    id: number
    name: string
    code: string
  }
}

// 策略详情（包含规则）
export interface PolicyWithRules {
  policy: PolicyConfig
  rules: any[]
}

// 策略列表请求参数
export interface PolicyListRequest {
  page?: number
  page_size?: number
  name?: string
  domain?: string
  enabled?: boolean
  tenant_id?: number
}

// 策略列表响应
export interface PolicyListResponse {
  code: number
  message: string
  data: {
    list: PolicyConfig[]
    total: number
    page: number
    size: number
  }
}

// 创建策略请求参数
export interface CreatePolicyRequest {
  tenant_id?: number
  name: string
  description?: string
  domain_id?: number
  rule_ids?: number[]
  enabled: boolean
}

// 更新策略请求参数
export interface UpdatePolicyRequest {
  name?: string
  description?: string
  domain_id?: number
  rule_ids?: number[]
  enabled?: boolean
}

// API 函数
export const policyApi = {
  // 获取策略列表
  getPolicyList: (params: PolicyListRequest): Promise<PolicyListResponse> => {
    return request.get('/api/v1/policies', { params })
  },

  // 获取策略详情
  getPolicyDetail: (id: number): Promise<{ data: PolicyConfig }> => {
    return request.get(`/api/v1/policies/${id}`)
  },

  // 获取策略详情（包含规则）
  getPolicyWithRules: (id: number): Promise<{ data: PolicyWithRules }> => {
    return request.get(`/api/v1/policies/${id}/with-rules`)
  },

  // 创建策略
  createPolicy: (data: CreatePolicyRequest): Promise<{ data: PolicyConfig }> => {
    return request.post('/api/v1/policies', data)
  },

  // 更新策略
  updatePolicy: (id: number, data: UpdatePolicyRequest): Promise<{ data: PolicyConfig }> => {
    return request.put(`/api/v1/policies/${id}`, data)
  },

  // 删除策略
  deletePolicy: (id: number): Promise<void> => {
    return request.delete(`/api/v1/policies/${id}`)
  },

  // 切换策略状态
  togglePolicy: (id: number): Promise<void> => {
    return request.post(`/api/v1/policies/${id}/toggle`)
  },

  // 批量删除策略
  batchDeletePolicies: (ids: number[]): Promise<void> => {
    return request.delete('/api/v1/policies/batch', { data: { ids } })
  },

  // 获取可用的规则列表
  getAvailableRules: (): Promise<{ data: any[] }> => {
    return request.get('/api/v1/rules')
  }
} 