import request from '@/utils/request'

// 规则数据类型定义
export interface Rule {
  id?: number
  name: string
  description?: string
  match_type: 'uri' | 'ip' | 'header' | 'body' | 'user_agent'
  match_mode: 'exact' | 'regex' | 'contains'
  pattern: string
  action: 'block' | 'allow' | 'log'
  priority: number
  enabled: boolean
  created_at?: string
  updated_at?: string
}

export interface RuleListParams {
  page?: number
  page_size?: number
  name?: string
  enabled?: boolean
  match_type?: string
}

export interface RuleListResponse {
  list: Rule[]
  total: number
  page: number
  page_size: number
}

// 获取规则列表
export const getRuleList = (params: RuleListParams) => {
  return request.get<RuleListResponse>('/api/v1/rules', { params })
}

// 获取规则详情
export const getRuleDetail = (id: number) => {
  return request.get<Rule>(`/api/v1/rules/${id}`)
}

// 创建规则
export const createRule = (data: Omit<Rule, 'id' | 'created_at' | 'updated_at'>) => {
  return request.post<Rule>('/api/v1/rules', data)
}

// 更新规则
export const updateRule = (id: number, data: Partial<Rule>) => {
  return request.put<Rule>(`/api/v1/rules/${id}`, data)
}

// 删除规则
export const deleteRule = (id: number) => {
  return request.delete(`/api/v1/rules/${id}`)
}

// 切换规则状态
export const toggleRuleStatus = (id: number, enabled: boolean) => {
  return request.post(`/api/v1/rules/${id}/toggle`)
}

// 批量操作规则
export const batchUpdateRules = (ids: number[], data: Partial<Rule>) => {
  return request.patch('/api/v1/rules/batch', { ids, ...data })
}

// 导入规则
export const importRules = (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/api/v1/rules/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 导出规则
export const exportRules = (params?: { ids?: number[] }) => {
  return request.get('/api/v1/rules/export', { 
    params,
    responseType: 'blob'
  })
} 