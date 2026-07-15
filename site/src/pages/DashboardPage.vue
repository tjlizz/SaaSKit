<template>
  <div class="page">
    <div class="container">
      <div class="page-header">
        <div>
          <h1>控制台</h1>
          <p class="page-sub">欢迎回来，{{ auth.user?.name || auth.user?.email || '用户' }}</p>
        </div>
        <RouterLink to="/pricing" class="btn btn-primary">升级套餐</RouterLink>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="loading-state">
        <div class="spinner spinner-dark"></div>
        <span>加载数据中...</span>
      </div>

      <template v-else>
        <!-- Subscription Status -->
        <div class="section">
          <h2 class="section-title">当前订阅</h2>
          <div class="sub-card card">
            <template v-if="subscription?.valid">
              <div class="sub-header">
                <div class="sub-icon active">✓</div>
                <div>
                  <div class="sub-plan-name">{{ subscription.plan?.name }}</div>
                  <div class="sub-status">
                    <span class="badge badge-success">订阅中</span>
                  </div>
                </div>
                <div class="sub-expires">
                  <div class="expires-label">到期时间</div>
                  <div class="expires-date">{{ formatDate(subscription.expires_at) }}</div>
                </div>
              </div>

              <div class="sub-details">
                <div class="detail-item" v-if="subscription.plan?.device_limit > 0">
                  <span class="detail-label">设备数量</span>
                  <span class="detail-value">{{ subscription.plan?.device_limit }} 台</span>
                </div>
                <div class="detail-item" v-if="subscription.subscription?.remaining_credits > 0">
                  <span class="detail-label">剩余积分</span>
                  <span class="detail-value">{{ subscription.subscription?.remaining_credits?.toLocaleString() }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">计费周期</span>
                  <span class="detail-value">{{ cycleDesc(subscription.plan?.billing_cycle) }}</span>
                </div>
              </div>
            </template>

            <template v-else>
              <div class="sub-empty">
                <div class="sub-empty-icon">📦</div>
                <h3>暂无有效订阅</h3>
                <p>选择适合您的套餐，开始体验完整功能</p>
                <RouterLink to="/pricing" class="btn btn-primary">立即选择套餐</RouterLink>
              </div>
            </template>
          </div>
        </div>

        <!-- Profile -->
        <div class="section">
          <h2 class="section-title">账户信息</h2>
          <div class="profile-card card">
            <div class="profile-avatar">{{ avatarLetter }}</div>
            <div class="profile-info">
              <div class="profile-name">{{ profile?.name || '—' }}</div>
              <div class="profile-email">{{ profile?.email }}</div>
              <div class="profile-meta">
                <span class="badge badge-gray">注册用户</span>
                <span class="meta-date">加入于 {{ formatDate(profile?.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Orders -->
        <div class="section">
          <h2 class="section-title">订单记录</h2>
          <div v-if="orders.length === 0" class="empty-orders card">
            <p>暂无订单记录</p>
            <RouterLink to="/pricing" class="btn btn-outline" style="margin-top:12px">去购买套餐</RouterLink>
          </div>
          <div v-else class="orders-list">
            <div v-for="order in orders" :key="order.id" class="order-row card">
              <div class="order-main">
                <div class="order-no">{{ order.order_no }}</div>
                <div class="order-date">{{ formatDate(order.created_at) }}</div>
              </div>
              <div class="order-amount">¥{{ (order.amount_cents / 100).toFixed(2) }}</div>
              <div>
                <span class="badge" :class="statusBadge(order.status)">{{ statusLabel(order.status) }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth.js'
import { api } from '../api/index.js'

const auth = useAuthStore()
const loading = ref(true)
const profile = ref(null)
const subscription = ref(null)
const orders = ref([])

const avatarLetter = computed(() => {
  const name = profile.value?.name || profile.value?.email || '?'
  return name.charAt(0).toUpperCase()
})

function formatDate(d) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })
}

function cycleDesc(cycle) {
  return { monthly: '按月', yearly: '按年', lifetime: '永久', one_time: '一次性', free: '免费' }[cycle] || '—'
}

function statusLabel(s) {
  return { pending: '待支付', paid: '已支付', cancelled: '已取消', refunded: '已退款' }[s] || s
}

function statusBadge(s) {
  return { pending: 'badge-warning', paid: 'badge-success', cancelled: 'badge-gray', refunded: 'badge-danger' }[s] || 'badge-gray'
}

onMounted(async () => {
  try {
    const [p, s, o] = await Promise.all([
      api.getProfile(),
      api.getSubscription(),
      api.getOrders(),
    ])
    profile.value = p
    subscription.value = s
    orders.value = o || []
  } catch (e) {
    // Silently handle - individual sections will show empty states
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.page { padding: 40px 0 64px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 40px; flex-wrap: wrap; gap: 16px; }
.page-header h1 { font-size: 28px; font-weight: 700; margin-bottom: 4px; }
.page-sub { color: var(--muted); font-size: 14px; }
.loading-state { display: flex; align-items: center; gap: 12px; justify-content: center; padding: 80px; }

.section { margin-bottom: 32px; }
.section-title { font-size: 17px; font-weight: 600; margin-bottom: 14px; color: var(--text); }

/* Subscription */
.sub-card { padding: 28px; }
.sub-header { display: flex; align-items: center; gap: 16px; margin-bottom: 20px; flex-wrap: wrap; }
.sub-icon.active { width: 44px; height: 44px; border-radius: 50%; background: #d1fae5; color: #065f46; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 18px; flex-shrink: 0; }
.sub-plan-name { font-size: 18px; font-weight: 700; margin-bottom: 4px; }
.sub-expires { margin-left: auto; text-align: right; }
.expires-label { font-size: 12px; color: var(--muted); }
.expires-date { font-size: 15px; font-weight: 600; margin-top: 2px; }
.sub-details { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 12px; padding-top: 16px; border-top: 1px solid var(--border); }
.detail-item { display: flex; flex-direction: column; gap: 3px; }
.detail-label { font-size: 12px; color: var(--muted); }
.detail-value { font-size: 15px; font-weight: 600; }
.sub-empty { display: flex; flex-direction: column; align-items: center; text-align: center; padding: 24px; gap: 10px; }
.sub-empty-icon { font-size: 40px; }
.sub-empty h3 { font-size: 16px; font-weight: 600; }
.sub-empty p { color: var(--muted); font-size: 14px; }

/* Profile */
.profile-card { padding: 24px; display: flex; align-items: center; gap: 20px; flex-wrap: wrap; }
.profile-avatar { width: 56px; height: 56px; border-radius: 50%; background: linear-gradient(135deg, var(--primary), #7c3aed); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 24px; font-weight: 700; flex-shrink: 0; }
.profile-name { font-size: 17px; font-weight: 600; margin-bottom: 3px; }
.profile-email { color: var(--muted); font-size: 14px; margin-bottom: 8px; }
.profile-meta { display: flex; align-items: center; gap: 10px; }
.meta-date { font-size: 12px; color: var(--muted); }

/* Orders */
.empty-orders { padding: 32px; text-align: center; color: var(--muted); font-size: 14px; }
.orders-list { display: flex; flex-direction: column; gap: 10px; }
.order-row { padding: 16px 20px; display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; transition: box-shadow .15s; }
.order-row:hover { box-shadow: var(--shadow-md); }
.order-no { font-family: monospace; font-size: 13px; font-weight: 600; color: var(--text); }
.order-date { font-size: 12px; color: var(--muted); margin-top: 2px; }
.order-amount { font-size: 16px; font-weight: 700; color: var(--text); }
</style>
