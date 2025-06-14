import request from '@/utils/request'

// 黑名单数据类型定义
export interface BlackList {
  id?: number
  type: 'ip' | 'uri' | 'user_agent'
  value: string
  comment?: string
  enabled: boolean
  created_at?: string
  updated_at?: string
}

// 黑名单列表响应
export interface BlackListListResponse {
  list: BlackList[]
  total: number
  page: number
  size: number
}

// 创建黑名单
export function createBlackList(data: Omit<BlackList, 'id' | 'created_at' | 'updated_at'>) {
  return request({
    url: '/blacklists',
    method: 'post',
    data
  })
}

// 获取黑名单列表
export function getBlackListList(params: {
  page?: number
  page_size?: number
  type?: string
  value?: string
  enabled?: boolean
}) {
  return request({
    url: '/blacklists',
    method: 'get',
    params
  })
}

// 获取黑名单详情
export function getBlackListById(id: number) {
  return request({
    url: `/blacklists/${id}`,
    method: 'get'
  })
}

// 更新黑名单
export function updateBlackList(id: number, data: Partial<BlackList>) {
  return request({
    url: `/blacklists/${id}`,
    method: 'put',
    data
  })
}

// 删除黑名单
export function deleteBlackList(id: number) {
  return request({
    url: `/blacklists/${id}`,
    method: 'delete'
  })
}

// 批量删除黑名单
export function batchDeleteBlackList(ids: number[]) {
  return request({
    url: '/blacklists/batch',
    method: 'delete',
    data: ids
  })
}

// 切换黑名单状态
export function toggleBlackListStatus(id: number) {
  return request({
    url: `/blacklists/${id}/toggle`,
    method: 'patch'
  })
} 