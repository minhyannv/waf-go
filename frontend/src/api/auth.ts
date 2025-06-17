import request from '@/utils/request'

export interface LoginForm {
  username: string
  password: string
}

export interface UserInfo {
  id: number
  username: string
  role: string
  tenant_id: number
}

// 用户登录
export function login(data: LoginForm) {
  return request({
    url: '/api/v1/auth/login',
    method: 'post',
    data
  })
}

// 获取用户信息
export function getUserInfo() {
  return request({
    url: '/api/v1/auth/userinfo',
    method: 'get'
  })
}

// 退出登录
export function logout() {
  return request({
    url: '/api/v1/auth/logout',
    method: 'post'
  })
} 