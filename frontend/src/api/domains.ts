import request from '@/utils/request'

// 后端服务器配置
export interface BackendServer {
  id?: number
  domain_id?: number
  host: string
  port: number
  weight: number
  enabled: boolean
  created_at?: string
  updated_at?: string
}

// 域名数据类型定义
export interface DomainConfig {
  id?: number
  tenant_id?: number
  domain: string
  protocol: 'http' | 'https'
  port: number
  ssl_certificate?: string
  ssl_private_key?: string
  backend_url: string
  enabled: boolean
  created_at?: string
  updated_at?: string
  tenant?: {
    id: number
    name: string
  }
}

// 扩展域名配置类型
export interface ExtendedDomainConfig extends DomainConfig {
  statusLoading?: boolean
  health_check_enabled?: boolean
  health_check_path?: string
  health_check_interval?: number
  policy_ids?: number[]
}

// 域名列表请求参数
export interface DomainListRequest {
  page: number
  page_size: number
  domain?: string
  protocol?: string
  enabled?: boolean
}

// 域名列表响应
export interface DomainListResponse {
  list: DomainConfig[]
  total: number
}

// 创建域名请求参数
export interface CreateDomainRequest {
  tenant_id?: number
  domain: string
  protocol: 'http' | 'https'
  port?: number
  ssl_certificate?: string
  ssl_private_key?: string
  backend_url: string
  backend_servers?: BackendServer[]
  enabled: boolean
}

// 更新域名请求参数
export interface UpdateDomainRequest {
  domain?: string
  protocol?: 'http' | 'https'
  port?: number
  ssl_certificate?: string
  ssl_private_key?: string
  backend_url?: string
  backend_servers?: BackendServer[]
  enabled?: boolean
}

// 域名策略关联
export interface DomainPolicy {
  id: number
  domain_id: number
  policy_id: number
  priority: number
  enabled: boolean
  policy?: {
    id: number
    name: string
    description?: string
  }
}

// 域名策略关联项
export interface DomainPolicyItem {
  policy_id: number
  priority: number
  enabled: boolean
}

// API 函数
export const domainApi = {
  // 获取域名列表
  list(params: DomainListRequest) {
    return request.get<DomainListResponse>('/api/v1/domains', { params })
  },

  // 获取域名详情
  get: (id: number): Promise<{ data: DomainConfig }> => {
    return request.get(`/api/v1/domains/${id}`)
  },

  // 创建域名
  create(data: DomainConfig) {
    return request.post('/api/v1/domains', data)
  },

  // 更新域名
  update(id: number, data: DomainConfig) {
    return request.put(`/api/v1/domains/${id}`, data)
  },

  // 删除域名
  delete(id: number) {
    return request.delete(`/api/v1/domains/${id}`)
  },

  // 切换域名状态
  toggle(id: number) {
    return request.post(`/api/v1/domains/${id}/toggle`)
  },

  // 批量删除域名
  batchDelete: (ids: number[]): Promise<void> => {
    return request.delete('/api/v1/domains/batch', { data: { ids } })
  },

  // 获取域名关联的策略
  getPolicies: (id: number): Promise<{ data: DomainPolicy[] }> => {
    return request.get(`/api/v1/domains/${id}/policies`)
  },

  // 更新域名关联的策略
  updatePolicies: (id: number, policies: DomainPolicyItem[]): Promise<void> => {
    return request.put(`/api/v1/domains/${id}/policies`, { policies })
  }
} 