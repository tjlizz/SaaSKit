<template>
  <div>
    <!-- Hero -->
    <section class="hero">
      <div class="container">
        <div class="hero-badge"><span class="badge badge-primary">🎉 现已全面上线</span></div>
        <h1 class="hero-title">
          让业务管理<br/>
          <span class="gradient-text">更智能、更高效</span>
        </h1>
        <p class="hero-desc">
          CloudSuite 是一款专为现代企业设计的 SaaS 平台，帮助团队统一管理流程、数据与协作，<br/>
          告别繁琐操作，专注于真正重要的事情。
        </p>
        <div class="hero-cta">
          <RouterLink to="/register" class="btn btn-primary btn-lg">免费开始使用</RouterLink>
          <RouterLink to="/pricing" class="btn btn-ghost btn-lg">查看定价方案</RouterLink>
        </div>
        <p class="hero-note">无需信用卡 · 14 天免费试用 · 随时取消</p>
      </div>
    </section>

    <!-- Features -->
    <section class="features">
      <div class="container">
        <div class="section-header">
          <h2>为何选择 CloudSuite</h2>
          <p>我们提供完整的业务管理解决方案，帮助企业实现数字化转型</p>
        </div>
        <div class="features-grid">
          <div v-for="f in features" :key="f.title" class="feature-card card">
            <div class="feature-icon">{{ f.icon }}</div>
            <h3>{{ f.title }}</h3>
            <p>{{ f.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Pricing Preview -->
    <section class="pricing-preview">
      <div class="container">
        <div class="section-header">
          <h2>简单透明的定价</h2>
          <p>按需选择，随时升级，无隐藏费用</p>
        </div>

        <div v-if="loading" class="plans-loading">
          <div class="spinner spinner-dark"></div>
          <span>加载套餐中...</span>
        </div>

        <div v-else-if="error" class="plans-error alert alert-info">
          ℹ️ 暂时无法加载套餐信息，请稍后重试或
          <RouterLink to="/pricing" class="link">查看定价页面</RouterLink>。
        </div>

        <div v-else class="plans-grid">
          <div
            v-for="plan in plans.slice(0, 3)"
            :key="plan.id"
            class="plan-card card"
            :class="{ 'plan-card--recommended': plan.recommended }"
          >
            <div v-if="plan.recommended" class="plan-badge">推荐</div>
            <h3 class="plan-name">{{ plan.name }}</h3>
            <div class="plan-price">
              <span v-if="plan.billing_cycle === 'free'" class="price-free">免费</span>
              <template v-else>
                <span class="price-amount">¥{{ (plan.price_cents / 100).toFixed(0) }}</span>
                <span class="price-cycle">/ {{ cycleLabel(plan.billing_cycle) }}</span>
              </template>
            </div>
            <p class="plan-desc">{{ plan.description || '适合您的业务需求' }}</p>
            <RouterLink
              :to="auth.isLoggedIn ? `/checkout/${plan.plan_code}` : '/register'"
              class="btn btn-full"
              :class="plan.recommended ? 'btn-primary' : 'btn-outline'"
            >
              {{ plan.billing_cycle === 'free' ? '免费开始' : '立即购买' }}
            </RouterLink>
          </div>
        </div>

        <div class="pricing-cta">
          <RouterLink to="/pricing" class="btn btn-outline btn-lg">查看所有套餐 →</RouterLink>
        </div>
      </div>
    </section>

    <!-- CTA Banner -->
    <section class="cta-banner">
      <div class="container cta-inner">
        <div>
          <h2>准备好开始了吗？</h2>
          <p>加入数千家已在使用 CloudSuite 的企业</p>
        </div>
        <RouterLink to="/register" class="btn btn-primary btn-lg">立即免费注册</RouterLink>
      </div>
    </section>
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

const features = [
  { icon: '🚀', title: '极速部署', desc: '5 分钟完成配置，无需专业运维，即开即用，让您的业务快速上线。' },
  { icon: '🔒', title: '企业级安全', desc: '数据加密存储，权限精细管控，符合行业合规要求，保护您的核心资产。' },
  { icon: '📊', title: '实时数据分析', desc: '可视化仪表盘，多维数据报表，随时掌握业务关键指标和趋势。' },
  { icon: '🤝', title: '团队协作', desc: '多人实时协同，任务分配追踪，让团队沟通更顺畅，效率提升 3 倍。' },
  { icon: '🔗', title: '开放集成', desc: '丰富的 API 接口，轻松对接现有系统，构建专属业务生态。' },
  { icon: '📱', title: '全端支持', desc: '支持 Web、iOS、Android，随时随地访问，工作不受地点限制。' },
]

function cycleLabel(cycle) {
  return { monthly: '月', yearly: '年', lifetime: '永久', one_time: '次' }[cycle] || cycle
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
/* Hero */
.hero { padding: 80px 0 64px; text-align: center; }
.hero-badge { margin-bottom: 20px; }
.hero-title { font-size: clamp(36px, 6vw, 60px); font-weight: 800; line-height: 1.15; margin-bottom: 20px; letter-spacing: -1px; }
.gradient-text { background: linear-gradient(135deg, var(--primary), #7c3aed, #db2777); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.hero-desc { font-size: 18px; color: var(--muted); max-width: 600px; margin: 0 auto 32px; }
.hero-cta { display: flex; gap: 12px; justify-content: center; flex-wrap: wrap; }
.hero-note { margin-top: 16px; font-size: 13px; color: #9ca3af; }

/* Features */
.features { padding: 64px 0; background: var(--card); border-top: 1px solid var(--border); border-bottom: 1px solid var(--border); }
.section-header { text-align: center; margin-bottom: 48px; }
.section-header h2 { font-size: 32px; font-weight: 700; margin-bottom: 12px; }
.section-header p { color: var(--muted); font-size: 16px; }
.features-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 20px; }
.feature-card { padding: 28px; }
.feature-icon { font-size: 32px; margin-bottom: 12px; }
.feature-card h3 { font-size: 17px; font-weight: 600; margin-bottom: 8px; }
.feature-card p { font-size: 14px; color: var(--muted); line-height: 1.6; }

/* Pricing Preview */
.pricing-preview { padding: 64px 0; }
.plans-loading, .plans-error { display: flex; align-items: center; gap: 12px; justify-content: center; padding: 32px; }
.plans-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(260px, 1fr)); gap: 20px; margin-bottom: 32px; }
.plan-card { padding: 28px; position: relative; transition: transform .15s, box-shadow .15s; }
.plan-card:hover { transform: translateY(-2px); box-shadow: var(--shadow-lg); }
.plan-card--recommended { border-color: var(--primary); box-shadow: 0 0 0 2px rgba(79,70,229,.15), var(--shadow-md); }
.plan-badge { position: absolute; top: -12px; left: 50%; transform: translateX(-50%); background: var(--primary); color: #fff; padding: 3px 14px; border-radius: 999px; font-size: 12px; font-weight: 600; white-space: nowrap; }
.plan-name { font-size: 18px; font-weight: 700; margin-bottom: 12px; }
.plan-price { margin-bottom: 12px; display: flex; align-items: baseline; gap: 4px; }
.price-free { font-size: 28px; font-weight: 800; color: var(--success); }
.price-amount { font-size: 32px; font-weight: 800; }
.price-cycle { color: var(--muted); font-size: 14px; }
.plan-desc { font-size: 13px; color: var(--muted); margin-bottom: 20px; line-height: 1.5; }
.pricing-cta { text-align: center; }
.link { color: var(--primary); text-decoration: underline; }

/* CTA Banner */
.cta-banner { background: linear-gradient(135deg, var(--primary), #7c3aed); padding: 56px 0; margin-top: 0; }
.cta-inner { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 24px; }
.cta-banner h2 { font-size: 28px; font-weight: 700; color: #fff; margin-bottom: 6px; }
.cta-banner p { color: rgba(255,255,255,.8); font-size: 15px; }
.cta-banner .btn-primary { background: #fff; color: var(--primary); }
.cta-banner .btn-primary:hover { background: #f0f0ff; }
</style>
