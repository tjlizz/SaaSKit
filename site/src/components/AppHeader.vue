<template>
  <header class="header">
    <div class="container header-inner">
      <RouterLink to="/" class="logo">
        <span class="logo-icon">⚡</span>
        <span class="logo-text">CloudSuite</span>
      </RouterLink>

      <nav class="nav">
        <RouterLink to="/pricing" class="nav-link">定价</RouterLink>
        <template v-if="auth.isLoggedIn">
          <RouterLink to="/dashboard" class="nav-link">控制台</RouterLink>
          <button class="btn btn-ghost btn-sm" @click="handleLogout">退出</button>
        </template>
        <template v-else>
          <RouterLink to="/login" class="nav-link">登录</RouterLink>
          <RouterLink to="/register" class="btn btn-primary btn-sm">免费注册</RouterLink>
        </template>
      </nav>
    </div>
  </header>
</template>

<script setup>
import { useAuthStore } from '../stores/auth.js'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

function handleLogout() {
  auth.logout()
  router.push('/')
}
</script>

<style scoped>
.header {
  position: sticky; top: 0; z-index: 100;
  background: rgba(255,255,255,.95);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--border);
}
.header-inner {
  display: flex; align-items: center;
  justify-content: space-between;
  height: 64px;
}
.logo {
  display: flex; align-items: center; gap: 8px;
  text-decoration: none; font-weight: 700; font-size: 18px;
  color: var(--text);
}
.logo-icon { font-size: 22px; }
.logo-text { background: linear-gradient(135deg, var(--primary), #7c3aed); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.nav { display: flex; align-items: center; gap: 8px; }
.nav-link {
  padding: 6px 14px; border-radius: 6px; font-size: 14px; font-weight: 500;
  color: var(--muted); text-decoration: none; transition: color .15s;
}
.nav-link:hover { color: var(--text); }
.nav-link.router-link-active { color: var(--primary); }
.btn-sm { padding: 7px 16px; font-size: 14px; }
</style>
