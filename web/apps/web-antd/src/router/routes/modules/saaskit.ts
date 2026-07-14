import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: { icon: 'lucide:boxes', order: 5, title: '应用中心' },
    name: 'ApplicationCenter',
    path: '/applications',
    children: [
      {
        component: () => import('#/views/saaskit/applications/index.vue'),
        meta: { icon: 'lucide:app-window', title: '应用管理' },
        name: 'Applications',
        path: '',
      },
    ],
  },
  {
    meta: { icon: 'lucide:users', order: 10, title: '用户中心' },
    name: 'CustomerCenter',
    path: '/customers',
    children: [
      {
        component: () => import('#/views/saaskit/users/index.vue'),
        meta: { icon: 'lucide:user-round-cog', title: '用户管理' },
        name: 'CustomerUsers',
        path: 'users',
      },
    ],
  },
  {
    meta: { icon: 'lucide:shopping-bag', order: 20, title: '商业化' },
    name: 'Commerce',
    path: '/commerce',
    children: [
      {
        component: () => import('#/views/saaskit/plans/index.vue'),
        meta: { icon: 'lucide:badge-dollar-sign', title: '套餐管理' },
        name: 'Plans',
        path: 'plans',
      },
      {
        component: () => import('#/views/saaskit/orders/index.vue'),
        meta: { icon: 'lucide:receipt-text', title: '订单管理' },
        name: 'Orders',
        path: 'orders',
      },
      {
        component: () => import('#/views/saaskit/subscriptions/index.vue'),
        meta: { icon: 'lucide:calendar-check-2', title: '订阅管理' },
        name: 'Subscriptions',
        path: 'subscriptions',
      },
    ],
  },
  {
    meta: { icon: 'lucide:settings', order: 30, title: '系统配置' },
    name: 'SystemSettings',
    path: '/system',
    children: [
      {
        component: () => import('#/views/saaskit/api-clients/index.vue'),
        meta: { icon: 'lucide:key-round', title: 'API 凭证' },
        name: 'ApiClients',
        path: 'api-clients',
      },
      {
        component: () => import('#/views/saaskit/payment-settings/index.vue'),
        meta: { icon: 'lucide:credit-card', title: '支付配置' },
        name: 'PaymentSettings',
        path: 'payment-settings',
      },
    ],
  },
];

export default routes;
