<template>
  <div class="page">
    <div class="container">
      <div class="return-card card">
        <!-- Checking -->
        <div v-if="status === 'checking'" class="state-box">
          <div class="spinner spinner-dark lg"></div>
          <h2>正在确认支付结果...</h2>
          <p>请稍候，我们正在向支付宝确认您的付款状态</p>
        </div>

        <!-- Success -->
        <div v-else-if="status === 'success'" class="state-box">
          <div class="status-icon success">✅</div>
          <h2>支付成功！</h2>
          <p>您的订阅已激活，感谢您的购买。</p>
          <div class="order-info" v-if="orderNo">
            <span class="info-label">订单号：</span>
            <code>{{ orderNo }}</code>
          </div>
          <RouterLink to="/dashboard" class="btn btn-primary btn-lg">前往控制台</RouterLink>
        </div>

        <!-- Pending -->
        <div v-else-if="status === 'pending'" class="state-box">
          <div class="status-icon warning">⏳</div>
          <h2>支付确认中</h2>
          <p>支付结果还在确认中，通常需要几秒钟。如果您已完成支付，请稍后刷新页面或前往控制台查看。</p>
          <div class="actions">
            <button class="btn btn-primary" @click="checkOrderStatus">重新查询</button>
            <RouterLink to="/dashboard" class="btn btn-ghost">前往控制台</RouterLink>
          </div>
        </div>

        <!-- Error / Not Found -->
        <div v-else class="state-box">
          <div class="status-icon error">❌</div>
          <h2>{{ status === 'cancelled' ? '支付已取消' : '支付失败' }}</h2>
          <p>{{ errorMsg || '支付未完成，您可以重试或选择其他套餐。' }}</p>
          <div class="actions">
            <RouterLink to="/pricing" class="btn btn-primary">重新选择套餐</RouterLink>
            <RouterLink to="/dashboard" class="btn btn-ghost">前往控制台</RouterLink>
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
const status = ref('checking')
const orderNo = ref(null)
const errorMsg = ref(null)
let pollCount = 0

async function checkOrderStatus() {
  const no = route.query.out_trade_no || route.query.order_no
  if (!no) {
    status.value = 'error'
    errorMsg.value = '未找到订单号'
    return
  }
  orderNo.value = no
  try {
    const order = await api.getOrder(no)
    if (order.status === 'paid') {
      status.value = 'success'
    } else if (order.status === 'cancelled' || order.status === 'refunded') {
      status.value = order.status
    } else if (order.status === 'pending' && pollCount < 5) {
      pollCount++
      // Poll a few times
      setTimeout(checkOrderStatus, 2000)
    } else {
      status.value = 'pending'
    }
  } catch (e) {
    status.value = 'error'
    errorMsg.value = e.message
  }
}

onMounted(checkOrderStatus)
</script>

<style scoped>
.page { min-height: calc(100vh - 64px - 80px); display: flex; align-items: center; justify-content: center; padding: 40px 20px; }
.return-card { width: 100%; max-width: 480px; padding: 48px 40px; margin: 0 auto; }
.state-box { display: flex; flex-direction: column; align-items: center; text-align: center; gap: 12px; }
.state-box h2 { font-size: 24px; font-weight: 700; }
.state-box p { color: var(--muted); font-size: 15px; line-height: 1.6; max-width: 340px; }
.spinner.lg { width: 40px; height: 40px; border-width: 3px; margin-bottom: 8px; }
.status-icon { font-size: 56px; margin-bottom: 8px; }
.order-info { background: var(--bg); padding: 10px 16px; border-radius: 8px; font-size: 13px; }
.info-label { color: var(--muted); }
.order-info code { font-family: monospace; color: var(--text); }
.actions { display: flex; gap: 12px; justify-content: center; flex-wrap: wrap; margin-top: 8px; }
</style>
