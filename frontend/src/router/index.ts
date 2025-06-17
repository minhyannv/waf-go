import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { title: '登录' }
    },
    {
      path: '/',
      redirect: '/dashboard',
      component: () => import('@/layouts/AdminLayout.vue'),
      children: [
        {
          path: '/dashboard',
          name: 'dashboard',
          component: () => import('@/views/DashboardView.vue'),
          meta: { title: '仪表盘', requiresAuth: true }
        },
        {
          path: '/rules',
          name: 'rules',
          component: () => import('@/views/RulesView.vue'),
          meta: { title: '规则管理', requiresAuth: true }
        },
        {
          path: '/logs',
          name: 'logs',
          component: () => import('@/views/LogsView.vue'),
          meta: { title: '攻击日志', requiresAuth: true }
        },
        {
          path: '/policies',
          name: 'policies',
          component: () => import('@/views/PoliciesView.vue'),
          meta: { title: '策略管理', requiresAuth: true }
        },
        {
          path: '/domains',
          name: 'domains',
          component: () => import('@/views/DomainsView.vue'),
          meta: { title: '域名管理', requiresAuth: true }
        },
        {
          path: '/settings',
          name: 'settings',
          component: () => import('@/views/SettingsView.vue'),
          meta: { title: '系统设置', requiresAuth: true }
        }
      ]
    }
  ]
})

// 全局路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta?.title) {
    document.title = `${to.meta.title} - WAF管理系统`
  }

  // 检查是否需要登录
  if (to.meta?.requiresAuth) {
    const token = localStorage.getItem('token')
    if (!token) {
      next({ name: 'login', query: { redirect: to.fullPath } })
      return
    }
  }
  
  // 如果已登录且访问登录页，重定向到首页
  if (to.name === 'login' && localStorage.getItem('token')) {
    next({ name: 'dashboard' })
    return
  }
  
  next()
})

export default router
