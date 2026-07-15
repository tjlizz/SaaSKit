<template>
  <div class="auth-page">
    <div class="auth-card card">
      <div class="auth-logo">
        <span class="logo-icon">⚡</span>
        <span class="logo-text">CloudSuite</span>
      </div>
      <h1 class="auth-title">创建您的账户</h1>
      <p class="auth-sub">免费注册，立即体验所有功能</p>

      <div v-if="error" class="alert alert-error mb">⚠️ {{ error }}</div>

      <form @submit.prevent="handleRegister" class="auth-form">
        <div class="form-group">
          <label class="form-label">姓名</label>
          <input
            v-model="form.name"
            type="text"
            class="form-input"
            placeholder="您的姓名"
          />
        </div>
        <div class="form-group">
          <label class="form-label">邮箱地址 <span class="required">*</span></label>
          <input
            v-model="form.email"
            type="email"
            class="form-input"
            placeholder="you@example.com"
            required
          />
        </div>
        <div class="form-group">
          <label class="form-label">密码 <span class="required">*</span></label>
          <input
            v-model="form.password"
            type="password"
            class="form-input"
            placeholder="至少 8 个字符"
            required
            minlength="8"
          />
          <p class="form-hint">至少 8 个字符</p>
        </div>

        <button type="submit" class="btn btn-primary btn-full" :disabled="loading">
          <span v-if="loading" class="spinner"></span>
          {{ loading ? '注册中...' : '创建账户' }}
        </button>
      </form>

      <div class="auth-footer">
        已有账户？<RouterLink to="/login" class="auth-link">立即登录</RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'
import { api } from '../api/index.js'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const form = ref({ name: '', email: '', password: '' })
const loading = ref(false)
const error = ref(null)

async function handleRegister() {
  error.value = null
  loading.value = true
  try {
    const data = await api.register({
      email: form.value.email,
      password: form.value.password,
      name: form.value.name,
    })
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
.form-label { font-size: 14px; font-weight: 500; color: var(--text); }
.required { color: var(--danger); }
.form-input {
  padding: 10px 14px; border: 1.5px solid var(--border); border-radius: 8px;
  font-size: 14px; outline: none; transition: border-color .15s;
  background: #fff;
}
.form-input:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(79,70,229,.1); }
.form-hint { font-size: 12px; color: var(--muted); }
.auth-footer { text-align: center; margin-top: 20px; font-size: 14px; color: var(--muted); }
.auth-link { color: var(--primary); text-decoration: none; font-weight: 500; }
.auth-link:hover { text-decoration: underline; }
</style>
