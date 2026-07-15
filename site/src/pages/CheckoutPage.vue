<template>
  <div class="page">
    <div class="container">
      <div class="breadcrumb">
        <RouterLink to="/pricing">定价</RouterLink>
        <span>›</span>
        <span>结账</span>
      </div>

      <!-- Loading plan -->
      <div v-if="loadingPlan" class="loading-state">
        <div class="spinner spinner-dark"></div>
        <span>加载套餐信息...</span>
      </div>

      <!-- Plan not found -->
      <div v-else-if="!plan" class="empty-state">
        <p>套餐不存在或已下线。</p>
        <RouterLink to="/pricing" class="btn btn-primary" style="margin-top:16px">查看其他套餐</RouterLink>
      </div>

      <div v-else class="checkout-layout">
        <!-- Order Summary -->
        <div class="order-summary card">
          <h2 class="summary-title">订单摘要</h2>

          <div class="plan-info">
            <div class="plan-icon">📦</div>
            <div>
              <div class="plan-info-name">{{ plan.name }}</div>
              <div class="plan-info-cycle">{{ cycleDesc(plan.billing_cycle) }}</div>
            </div>
          </div>

          <div class="summary-line">
            <span>套餐价格</span>
            <span>{{ formatPrice(plan) }}</span>
          </div>
          <div class="summary-divider"></div>
          <div class="summary-line summary-total">
            <span>合计</span>
            <span class="total-price">{{ formatPrice(plan) }}</span>
          </div>

          <div v-if="plan.description" class="plan-desc-box">
            {{ plan.description }}
          </div>

          <div class="plan-features">
            <div v-if="plan.device_limit > 0" class="feature-line">
              <span class="check">✓</span> {{ plan.device_limit }} 台设备
            </div>
            <div v-if="plan.credit_quota > 0" class="feature-line">
              <span class="check">✓</span> {{ plan.credit_quota.toLocaleString() }} 积分/月
            </div>
            <div class="feature-line"><span class="check">✓</span> 7×24 技术支持</div>
            <div class="feature-line"><span class="check">✓</span> 数据加密存储</div>
          </div>

          <div class="security-note">
            🔒 支付宝安全加密支付 · 7 天无理由退款
          </div>
        </div>

        <!-- Payment Panel -->
        <div class="payment-panel">
          <!-- Step: Confirm -->
          <div v-if="step === 'confirm'" class="pay-card card">
            <h2 class="pay-title">选择支付方式</h2>

            <div class="pay-methods">
              <div class="pay-method active">
                <img src="https://img.alicdn.com/tfs/TB1O6OYmkvoK1RjSZFwXXciCFXa-106-113.png" alt="支付宝" class="pay-logo" />
                <span>支付宝</span>
                <span class="pay-method-check">✓</span>
              </div>
            </div>

            <div v-if="plan.billing_cycle === 'free'" class="free-note alert alert-success">
              🎉 这是免费套餐，无需付款即可激活。
            </div>

            <div v-if="payError" class="alert alert-error" style="margin-bottom:16px">⚠️ {{ payError }}</div>

            <button
              class="btn btn-primary btn-full btn-lg"
              @click="submitOrder"
              :disabled="paying"
            >
              <span v-if="paying" class="spinner"></span>
              {{ paying ? '处理中...' : plan.billing_cycle === 'free' ? '激活免费套餐' : `支付 ${formatPrice(plan)}` }}
            </button>

            <p class="pay-terms">
              点击支付即表示您同意
              <a href="#" class="link">服务条款</a>和<a href="#" class="link">隐私政策</a>
            </p>
          </div>

          <!-- Step: Mock Confirm (simulate Alipay) -->
          <div v-else-if="step === 'mock_confirm'" class="pay-card card mock-pay">
            <div class="mock-header">
              <div class="mock-icon">🏦</div>
              <h2>模拟支付页面</h2>
              <p class="mock-hint">（测试模式 · 仅限开发环境）</p>
            </div>

            <div class="mock-order-info">
              <div class="mock-info-row">
                <span>商品名称</span>
                <strong>{{ plan.name }}</strong>
              </div>
              <div class="mock-info-row">
                <span>支付金额</span>
                <strong class="mock-amount">{{ formatPrice(plan) }}</strong>
              </div>
              <div class="mock-info-row">
                <span>订单号</span>
                <code class="mock-order-no">{{ orderNo }}</code>
              </div>
            </div>

            <div v-if="payError" class="alert alert-error" style="margin-bottom:16px">⚠️ {{ payError }}</div>

            <button
              class="btn btn-primary btn-full btn-lg"
              style="background: #1677ff; border-color: #1677ff;"
              @click="confirmMockPay"
              :disabled="paying"
            >
              <span v-if="paying" class="spinner"></span>
              {{ paying ? '支付处理中...' : '确认付款' }}
            </button>

            <button class="btn btn-ghost btn-full" style="margin-top:10px" @click="step = 'confirm'">
              返回
            </button>
          </div>

          <!-- Step: Redirect (real Alipay) -->
          <div v-else-if="step === 'redirecting'" class="pay-card card">
            <div class="redirect-state">
              <div class="spinner spinner-dark" style="width:36px;height:36px;border-width:3px"></div>
              <h3>正在跳转至支付宝...</h3>
              <p>请在支付宝页面完成付款，完成后将自动返回</p>
              <a :href="paymentUrl" class="btn btn-primary" style="margin-top:16px">
                手动跳转至支付宝
              </a>
            </div>
          </div>

          <!-- Step: Success -->
          <div v-else-if="step === 'success'" class="pay-card card">
            <div class="success-state">
              <div class="success-icon">✅</div>
              <h2>支付成功！</h2>
              <p>您已成功订阅 <strong>{{ plan.name }}</strong></p>
              <RouterLink to="/dashboard" class="btn btn-primary btn-lg" style="margin-top:24px">
                前往控制台
              </RouterLink>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../api/index.js'

const route = useRoute()

const plan = ref(null)
const loadingPlan = ref(true)
const step = ref('confirm')    // confirm | mock_confirm | redirecting | success
const paying = ref(false)
const payError = ref(null)
const orderNo = ref(null)
const paymentUrl = ref(null)

function formatPrice(p) {
  if (p.billing_cycle === 'free') return '免费'
  const yuan = (p.price_cents / 100).toFixed(2)
  return `¥${yuan}`
}

function cycleDesc(cycle) {
  return { monthly: '按月订阅', yearly: '按年订阅', lifetime: '永久买断', one_time: '一次性', free: '免费永久' }[cycle] || cycle
}

async function loadPlan() {
  try {
    const plans = await api.getPlans()
    plan.value = plans.find(p => p.plan_code === route.params.planCode) || null
  } catch (e) {
    plan.value = null
  } finally {
    loadingPlan.value = false
  }
}

async function submitOrder() {
  payError.value = null
  paying.value = true
  try {
    const returnURL = `${window.location.origin}/payment/return`
    const result = await api.createOrder({
      plan_code: plan.value.plan_code,
      provider: plan.value.billing_cycle === 'free' ? 'free' : 'alipay',
      channel: 'page',
      return_url: returnURL,
    })

    // Free plan → immediate success
    if (result.payment?.type === 'none' || plan.value.billing_cycle === 'free') {
      step.value = 'success'
      return
    }

    orderNo.value = result.order?.order_no

    // Mock payment mode
    if (result.payment?.type === 'mock') {
      step.value = 'mock_confirm'
      return
    }

    // Real Alipay redirect
    if (result.payment?.payment_url) {
      paymentUrl.value = result.payment.payment_url
      step.value = 'redirecting'
      setTimeout(() => { window.location.href = result.payment.payment_url }, 1000)
    }
  } catch (e) {
    payError.value = e.message
  } finally {
    paying.value = false
  }
}

async function confirmMockPay() {
  payError.value = null
  paying.value = true
  try {
    await api.mockPay(orderNo.value)
    step.value = 'success'
  } catch (e) {
    payError.value = e.message
  } finally {
    paying.value = false
  }
}

onMounted(loadPlan)
</script>

<style scoped>
.page { padding: 32px 0 64px; }
.breadcrumb { display: flex; align-items: center; gap: 8px; font-size: 14px; color: var(--muted); margin-bottom: 28px; }
.breadcrumb a { color: var(--primary); text-decoration: none; }
.breadcrumb a:hover { text-decoration: underline; }
.loading-state { display: flex; align-items: center; justify-content: center; gap: 12px; padding: 64px; }
.empty-state { text-align: center; padding: 64px; color: var(--muted); }

.checkout-layout { display: grid; grid-template-columns: 1fr 1fr; gap: 28px; align-items: start; }
@media (max-width: 768px) { .checkout-layout { grid-template-columns: 1fr; } }

/* Order Summary */
.order-summary { padding: 28px; }
.summary-title { font-size: 18px; font-weight: 700; margin-bottom: 20px; }
.plan-info { display: flex; align-items: center; gap: 14px; padding: 16px; background: var(--bg); border-radius: 8px; margin-bottom: 20px; }
.plan-icon { font-size: 28px; }
.plan-info-name { font-weight: 600; font-size: 16px; }
.plan-info-cycle { color: var(--muted); font-size: 13px; margin-top: 2px; }
.summary-line { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; font-size: 14px; color: var(--muted); }
.summary-divider { border-top: 1px solid var(--border); margin: 8px 0; }
.summary-total { font-weight: 600; font-size: 15px; color: var(--text); }
.total-price { font-size: 20px; font-weight: 800; color: var(--primary); }
.plan-desc-box { background: var(--bg); padding: 12px 14px; border-radius: 6px; font-size: 13px; color: var(--muted); line-height: 1.5; margin: 16px 0; }
.plan-features { display: flex; flex-direction: column; gap: 8px; margin: 16px 0; }
.feature-line { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.check { color: var(--success); font-weight: 700; }
.security-note { font-size: 12px; color: #9ca3af; text-align: center; margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--border); }

/* Payment Panel */
.pay-card { padding: 28px; }
.pay-title { font-size: 18px; font-weight: 700; margin-bottom: 20px; }
.pay-methods { display: flex; flex-direction: column; gap: 10px; margin-bottom: 20px; }
.pay-method { display: flex; align-items: center; gap: 12px; padding: 14px 16px; border: 2px solid var(--border); border-radius: 10px; cursor: pointer; transition: border-color .15s; position: relative; }
.pay-method.active { border-color: var(--primary); background: #fafaf0; }
.pay-logo { width: 32px; height: 32px; object-fit: contain; }
.pay-method span { font-size: 15px; font-weight: 500; }
.pay-method-check { margin-left: auto; color: var(--primary); font-size: 16px; font-weight: 700; }
.free-note { margin-bottom: 16px; }
.pay-terms { text-align: center; font-size: 12px; color: #9ca3af; margin-top: 12px; }
.link { color: var(--primary); text-decoration: none; }
.link:hover { text-decoration: underline; }

/* Mock Pay */
.mock-pay { background: linear-gradient(135deg, #f0f9ff, #e0f2fe); border: 2px solid #bae6fd; }
.mock-header { text-align: center; margin-bottom: 24px; }
.mock-icon { font-size: 48px; margin-bottom: 12px; }
.mock-header h2 { font-size: 20px; font-weight: 700; }
.mock-hint { font-size: 12px; color: var(--muted); margin-top: 4px; }
.mock-order-info { background: rgba(255,255,255,.7); border-radius: 10px; padding: 16px; margin-bottom: 20px; display: flex; flex-direction: column; gap: 10px; }
.mock-info-row { display: flex; justify-content: space-between; align-items: center; font-size: 14px; }
.mock-info-row span:first-child { color: var(--muted); }
.mock-amount { font-size: 20px; font-weight: 800; color: #1677ff; }
.mock-order-no { font-family: monospace; font-size: 12px; color: var(--muted); background: var(--border); padding: 2px 6px; border-radius: 4px; }

/* Redirect State */
.redirect-state { text-align: center; padding: 20px 0; }
.redirect-state h3 { font-size: 18px; font-weight: 600; margin: 16px 0 8px; }
.redirect-state p { color: var(--muted); font-size: 14px; }

/* Success State */
.success-state { text-align: center; padding: 20px 0; }
.success-icon { font-size: 56px; margin-bottom: 16px; }
.success-state h2 { font-size: 24px; font-weight: 700; margin-bottom: 8px; }
.success-state p { color: var(--muted); font-size: 15px; }
</style>
