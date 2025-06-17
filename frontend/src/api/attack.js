import request from '@/utils/request'

// 获取攻击日志列表
export function getAttackLogs(params) {
  return request({
    url: '/api/v1/attack-logs',
    method: 'get',
    params
  })
}

// 获取攻击日志详情
export function getAttackLogDetail(id) {
  return request({
    url: `/api/v1/attack-logs/${id}`,
    method: 'get'
  })
} 