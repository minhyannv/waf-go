import { createRouter, createWebHistory } from 'vue-router'

// 路由守卫
const requireAuth = (to: any, from: any, next: any) => {
  const token = localStorage.getItem('token')
  if (!token) {
    next('/login')
  } else {
    next()
  }
}

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
          meta: { title: '仪表盘' }
        },
        {
          path: '/rules',
          name: 'rules',
          component: () => import('@/views/RulesView.vue'),
          meta: { title: '规则管理' }
        },
        {
          path: '/logs',
          name: 'logs',
          component: () => import('@/views/LogsView.vue'),
          meta: { title: '攻击日志' }
        },
        {
          path: '/policies',
          name: 'policies',
          component: () => import('@/views/PoliciesView.vue'),
          meta: { title: '策略管理' }
        },
        {
          path: '/settings',
          name: 'settings',
          component: () => import('@/views/SettingsView.vue'),
          meta: { title: '系统设置' }
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
  
  next()
})

export default router
