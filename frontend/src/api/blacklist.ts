import request from '@/utils/request'

export interface BlackList {
  id?: number
  type: 'ip' | 'uri' | 'user_agent'
  value: string
  comment?: string
  enabled?: boolean
}

// 创建黑名单
export function createBlackList(data: BlackList) {
  return request({
    url: '/api/v1/blacklists',
    method: 'post',
    data
  })
}

// 获取黑名单列表
export function getBlackList(params?: any) {
  return request({
    url: '/api/v1/blacklists',
    method: 'get',
    params
  })
}

// 更新黑名单
export function updateBlackList(id: number, data: BlackList) {
  return request({
    url: `/api/v1/blacklists/${id}`,
    method: 'put',
    data
  })
}

// 删除黑名单
export function deleteBlackList(id: number) {
  return request({
    url: `/api/v1/blacklists/${id}`,
    method: 'delete'
  })
}

// 批量删除黑名单
export function batchDeleteBlackList(ids: number[]) {
  return request({
    url: '/api/v1/blacklists/batch',
    method: 'delete',
    data: { ids }
  })
}

// 切换黑名单状态
export function toggleBlackList(id: number) {
  return request({
    url: `/api/v1/blacklists/${id}/toggle`,
    method: 'patch'
  })
} 