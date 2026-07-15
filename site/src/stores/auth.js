import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('user_token') || null)
  const user = ref(JSON.parse(localStorage.getItem('user_info') || 'null'))

  const isLoggedIn = computed(() => !!token.value)

  function setAuth(accessToken, userInfo) {
    token.value = accessToken
    user.value = userInfo
    localStorage.setItem('user_token', accessToken)
    localStorage.setItem('user_info', JSON.stringify(userInfo))
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('user_token')
    localStorage.removeItem('user_info')
  }

  return { token, user, isLoggedIn, setAuth, logout }
})
