<template>
  <div class="page">
    <div class="container">
      <div class="page-header">
        <h1>透明定价，按需选择</h1>
        <p>无隐藏收费，无复杂合同，按需付费，随时升级或降级</p>
      </div>

      <div v-if="loading" class="loading-state">
        <div class="spinner spinner-dark"></div>
        <span>加载套餐中...</span>
      </div>

      <div v-else-if="error" class="alert alert-error">⚠️ {{ error }}</div>

      <div v-else-if="plans.length === 0" class="empty-state">
        暂无可用套餐，请联系管理员配置套餐。
      </div>

      <div v-else class="plans-grid">
        <div
          v-for="plan in plans"
          :key="plan.id"
          class="plan-card card"
          :class="{ 'plan-card--recommended': plan.recommended }"
        >
          <div v-if="plan.recommended" class="plan-badge">🏆 最受欢迎</div>

          <div class="plan-header">
            <h2 class="plan-name">{{ plan.name }}</h2>
            <div class="plan-price">
              <template v-if="plan.billing_cycle === 'free'">
                <span class="price-main free">免费</span>
              </template>
              <template v-else>
                <span class="price-currency">¥</span>
                <span class="price-main">{{ (plan.price_cents / 100).toFixed(0) }}</span>
                <span class="price-cycle">/ {{ cycleLabel(plan.billing_cycle) }}</span>
              </template>
            </div>
          </div>

          <p class="plan-desc">{{ plan.description || '适合您的业务需求' }}</p>

          <div class="plan-meta">
            <div v-if="plan.device_limit > 0" class="meta-item">
              <span class="meta-icon">💻</span>
              {{ plan.device_limit }} 台设备
            </div>
            <div v-if="plan.credit_quota > 0" class="meta-item">
              <span class="meta-icon">⚡</span>
              {{ plan.credit_quota.toLocaleString() }} 积分/月
            </div>
            <div class="meta-item">
              <span class="meta-icon">🔄</span>
              {{ cycleDesc(plan.billing_cycle) }}
            </div>
          </div>

          <RouterLink
            :to="auth.isLoggedIn ? `/checkout/${plan.plan_code}` : `/login?redirect=/checkout/${plan.plan_code}`"
            class="btn btn-full"
            :class="plan.recommended ? 'btn-primary' : 'btn-outline'"
          >
            {{ plan.billing_cycle === 'free' ? '免费开始使用' : '立即购买' }}
          </RouterLink>

          <p v-if="plan.billing_cycle !== 'free'" class="plan-note">
            支付宝安全支付 · 7 天无理由退款
          </p>
        </div>
      </div>

      <!-- FAQ -->
      <div class="faq">
        <h2>常见问题</h2>
        <div class="faq-list">
          <div v-for="q in faqs" :key="q.q" class="faq-item card">
            <h3>{{ q.q }}</h3>
            <p>{{ q.a }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth.js'
import { api } from '../api/index.js'

const auth = useAuthStore()
const plans = ref([])
const loading = ref(true)
const error = ref(null)

const faqs = [
  { q: '可以随时取消订阅吗？', a: '是的，您可以随时在控制台取消订阅，不收取任何违约金。年付套餐支持按剩余时间比例退款。' },
  { q: '支持哪些支付方式？', a: '目前支持支付宝支付。后续将陆续接入微信支付等更多方式。' },
  { q: '套餐可以升级或降级吗？', a: '可以随时升级套餐，升级时按剩余天数折算差价。降级将在当前计费周期结束后生效。' },
  { q: '数据安全如何保障？', a: '所有数据均经过加密存储，定期备份，符合国内数据安全法规要求，请放心使用。' },
]

function cycleLabel(cycle) {
  return { monthly: '月', yearly: '年', lifetime: '永久', one_time: '次' }[cycle] || cycle
}
function cycleDesc(cycle) {
  return { monthly: '按月订阅', yearly: '按年订阅（更优惠）', lifetime: '永久买断', one_time: '一次性付款', free: '免费永久使用' }[cycle] || cycle
}

onMounted(async () => {
  try {
    plans.value = await api.getPlans()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.page { padding: 48px 0 64px; }
.page-header { text-align: center; margin-bottom: 48px; }
.page-header h1 { font-size: 40px; font-weight: 800; margin-bottom: 12px; letter-spacing: -0.5px; }
.page-header p { color: var(--muted); font-size: 17px; }
.loading-state { display: flex; align-items: center; justify-content: center; gap: 12px; padding: 48px; }
.empty-state { text-align: center; color: var(--muted); padding: 48px; font-size: 15px; }

.plans-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 24px; margin-bottom: 64px; }
.plan-card { padding: 32px; position: relative; display: flex; flex-direction: column; gap: 0; transition: transform .15s, box-shadow .15s; }
.plan-card:hover { transform: translateY(-3px); box-shadow: var(--shadow-lg); }
.plan-card--recommended { border: 2px solid var(--primary); }
.plan-badge { position: absolute; top: -14px; left: 50%; transform: translateX(-50%); background: var(--primary); color: #fff; padding: 4px 16px; border-radius: 999px; font-size: 12px; font-weight: 600; white-space: nowrap; }
.plan-header { margin-bottom: 16px; }
.plan-name { font-size: 22px; font-weight: 700; margin-bottom: 10px; }
.plan-price { display: flex; align-items: baseline; gap: 2px; }
.price-currency { font-size: 20px; font-weight: 700; color: var(--text); }
.price-main { font-size: 40px; font-weight: 800; line-height: 1; }
.price-main.free { font-size: 32px; color: var(--success); }
.price-cycle { font-size: 14px; color: var(--muted); margin-left: 4px; }
.plan-desc { color: var(--muted); font-size: 14px; line-height: 1.6; margin-bottom: 20px; flex: 1; }
.plan-meta { display: flex; flex-direction: column; gap: 8px; margin-bottom: 24px; }
.meta-item { display: flex; align-items: center; gap: 8px; font-size: 14px; color: var(--text); }
.meta-icon { width: 20px; text-align: center; }
.plan-note { text-align: center; font-size: 12px; color: #9ca3af; margin-top: 10px; }

.faq { margin-top: 0; }
.faq h2 { font-size: 28px; font-weight: 700; text-align: center; margin-bottom: 32px; }
.faq-list { display: grid; grid-template-columns: repeat(auto-fit, minmax(340px, 1fr)); gap: 16px; }
.faq-item { padding: 24px; }
.faq-item h3 { font-size: 15px; font-weight: 600; margin-bottom: 8px; }
.faq-item p { font-size: 14px; color: var(--muted); line-height: 1.6; }
</style>
