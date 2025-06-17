import request from '@/utils/request'

// 白名单数据类型定义
export interface WhiteList {
  id?: number
  type: 'ip' | 'uri' | 'user_agent'
  value: string
  comment?: string
  enabled: boolean
  created_at?: string
  updated_at?: string
}

// 白名单列表响应
export interface WhiteListListResponse {
  list: WhiteList[]
  total: number
  page: number
  size: number
}

// 创建白名单
export function createWhiteList(data: Omit<WhiteList, 'id' | 'created_at' | 'updated_at'>) {
  return request({
    url: '/api/v1/whitelists',
    method: 'post',
    data
  })
}

// 获取白名单列表
export function getWhiteListList(params: {
  page?: number
  page_size?: number
  type?: string
  value?: string
  enabled?: boolean
}) {
  return request({
    url: '/api/v1/whitelists',
    method: 'get',
    params
  })
}

// 获取白名单详情
export function getWhiteListById(id: number) {
  return request({
    url: `/api/v1/whitelists/${id}`,
    method: 'get'
  })
}

// 更新白名单
export function updateWhiteList(id: number, data: Partial<WhiteList>) {
  return request({
    url: `/api/v1/whitelists/${id}`,
    method: 'put',
    data
  })
}

// 删除白名单
export function deleteWhiteList(id: number) {
  return request({
    url: `/api/v1/whitelists/${id}`,
    method: 'delete'
  })
}

// 批量删除白名单
export function batchDeleteWhiteList(ids: number[]) {
  return request({
    url: '/api/v1/whitelists/batch',
    method: 'delete',
    data: ids
  })
}

// 切换白名单状态
export function toggleWhiteListStatus(id: number) {
  return request({
    url: `/api/v1/whitelists/${id}/toggle`,
    method: 'patch'
  })
} 