<template>
  <div class="auth-page">
    <div class="auth-card card">
      <div class="auth-logo">
        <span class="logo-icon">⚡</span>
        <span class="logo-text">CloudSuite</span>
      </div>
      <h1 class="auth-title">欢迎回来</h1>
      <p class="auth-sub">登录您的账户继续使用</p>

      <div v-if="error" class="alert alert-error mb">⚠️ {{ error }}</div>

      <form @submit.prevent="handleLogin" class="auth-form">
        <div class="form-group">
          <label class="form-label">邮箱地址</label>
          <input
            v-model="form.email"
            type="email"
            class="form-input"
            placeholder="you@example.com"
            required
            autofocus
          />
        </div>
        <div class="form-group">
          <label class="form-label">密码</label>
          <input
            v-model="form.password"
            type="password"
            class="form-input"
            placeholder="输入密码"
            required
          />
        </div>

        <button type="submit" class="btn btn-primary btn-full" :disabled="loading">
          <span v-if="loading" class="spinner"></span>
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>

      <div class="auth-footer">
        还没有账户？<RouterLink :to="registerLink" class="auth-link">免费注册</RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'
import { api } from '../api/index.js'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const form = ref({ email: '', password: '' })
const loading = ref(false)
const error = ref(null)

const registerLink = computed(() => {
  const redirect = route.query.redirect
  return redirect ? `/register?redirect=${encodeURIComponent(redirect)}` : '/register'
})

async function handleLogin() {
  error.value = null
  loading.value = true
  try {
    const data = await api.login({ email: form.value.email, password: form.value.password })
    auth.setAuth(data.access_token, data.user)
    const redirect = route.query.redirect || '/dashboard'
    router.push(redirect)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page { min-height: calc(100vh - 64px - 80px); display: flex; align-items: center; justify-content: center; padding: 40px 20px; }
.auth-card { width: 100%; max-width: 420px; padding: 40px; }
.auth-logo { display: flex; align-items: center; justify-content: center; gap: 8px; margin-bottom: 24px; font-weight: 700; font-size: 20px; }
.logo-icon { font-size: 24px; }
.logo-text { background: linear-gradient(135deg, var(--primary), #7c3aed); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.auth-title { font-size: 24px; font-weight: 700; text-align: center; margin-bottom: 6px; }
.auth-sub { color: var(--muted); text-align: center; font-size: 14px; margin-bottom: 24px; }
.mb { margin-bottom: 16px; }
.auth-form { display: flex; flex-direction: column; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-label { font-size: 14px; font-weight: 500; }
.form-input {
  padding: 10px 14px; border: 1.5px solid var(--border); border-radius: 8px;
  font-size: 14px; outline: none; transition: border-color .15s; background: #fff;
}
.form-input:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(79,70,229,.1); }
.auth-footer { text-align: center; margin-top: 20px; font-size: 14px; color: var(--muted); }
.auth-link { color: var(--primary); text-decoration: none; font-weight: 500; }
.auth-link:hover { text-decoration: underline; }
</style>
