import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'

const routes = [
  { path: '/', component: () => import('../pages/HomePage.vue') },
  { path: '/pricing', component: () => import('../pages/PricingPage.vue') },
  { path: '/register', component: () => import('../pages/RegisterPage.vue') },
  { path: '/login', component: () => import('../pages/LoginPage.vue') },
  {
    path: '/checkout/:planCode',
    component: () => import('../pages/CheckoutPage.vue'),
    meta: { requiresAuth: true },
  },
  { path: '/payment/return', component: () => import('../pages/PaymentReturnPage.vue') },
  {
    path: '/dashboard',
    component: () => import('../pages/DashboardPage.vue'),
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
})

export default router
