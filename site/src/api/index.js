const BASE = '/site-api'

async function request(path, options = {}) {
  const token = localStorage.getItem('user_token')
  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  })
  const data = await res.json()
  if (!res.ok || data.code !== 0) {
    throw new Error(data.message || `请求失败 (${res.status})`)
  }
  return data.data
}

export const api = {
  // Plans (public)
  getPlans: () => request('/plans'),

  // Auth
  register: (body) => request('/auth/register', { method: 'POST', body: JSON.stringify(body) }),
  login: (body) => request('/auth/login', { method: 'POST', body: JSON.stringify(body) }),

  // Account (requires auth)
  getProfile: () => request('/account/profile'),
  getSubscription: () => request('/account/subscription'),
  getOrders: () => request('/account/orders'),

  // Orders
  createOrder: (body) => request('/orders', { method: 'POST', body: JSON.stringify(body) }),
  getOrder: (orderNo) => request(`/orders/${encodeURIComponent(orderNo)}`),
  mockPay: (orderNo) => request(`/orders/${encodeURIComponent(orderNo)}/mock-pay`, { method: 'POST' }),
}
