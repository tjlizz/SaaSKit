export class SaaSKit {
  constructor({ baseURL, apiKey, apiSecret, fetchImpl = globalThis.fetch }) {
    if (!baseURL || !apiKey || !apiSecret) throw new Error('baseURL, apiKey and apiSecret are required');
    if (!fetchImpl) throw new Error('A fetch implementation is required');
    this.baseURL = baseURL.replace(/\/$/, ''); this.apiKey = apiKey; this.apiSecret = apiSecret; this.fetch = fetchImpl;
  }
  async request(path, options = {}) {
    const response = await this.fetch(`${this.baseURL}/api/client${path}`, { ...options, headers: {'Content-Type':'application/json','X-API-Key':this.apiKey,'X-API-Secret':this.apiSecret,...options.headers} });
    const payload = await response.json(); if (!response.ok || payload.code !== 0) throw new Error(payload.message || `SaaSKit request failed (${response.status})`); return payload.data;
  }
  createOrder({ planCode, userId, provider = 'alipay', channel = 'page', returnURL, quantity = 1, requestId }) {
    return this.request('/orders', {method:'POST',body:JSON.stringify({plan_code:planCode,user_id:userId,provider,channel,return_url:returnURL,quantity,request_id:requestId})});
  }
  getOrder(orderNo) { return this.request(`/orders/${encodeURIComponent(orderNo)}`); }
  getProviderOrderStatus(orderNo) { return this.request(`/orders/${encodeURIComponent(orderNo)}/provider-status`); }
  checkSubscription(userId) { return this.request(`/subscription/check?user_id=${encodeURIComponent(userId)}`); }
}
