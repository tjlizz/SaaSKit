import express from 'express'
import { fileURLToPath } from 'url'
import path from 'path'
import fs from 'fs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// ─── Load .env manually (no dotenv dependency) ───────────────────────────────
const envPath = path.join(__dirname, '.env')
if (fs.existsSync(envPath)) {
  fs.readFileSync(envPath, 'utf-8')
    .split('\n')
    .forEach(line => {
      const m = line.match(/^([^#=]+)=(.*)$/)
      if (m) process.env[m[1].trim()] = m[2].trim()
    })
}

const SAASKIT_URL = process.env.SAASKIT_URL || 'http://localhost:8083'
const APP_KEY     = process.env.APP_KEY     || ''
const API_KEY     = process.env.API_KEY     || ''
const API_SECRET  = process.env.API_SECRET  || ''
const PORT        = Number(process.env.PORT || 3001)

// ─── Validate required config ─────────────────────────────────────────────────
if (!APP_KEY || !API_KEY || !API_SECRET) {
  console.warn('\n⚠️  Warning: APP_KEY / API_KEY / API_SECRET not set in .env')
  console.warn('   Copy .env.example → .env and fill in your SaaSKit credentials.\n')
}

const app = express()
app.use(express.json())

// ─── Utility: fetch from SaaSKit backend ─────────────────────────────────────
async function sk(path, options = {}) {
  const url = `${SAASKIT_URL}/api${path}`
  const res = await fetch(url, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options.headers },
  })
  const body = await res.json()
  return { status: res.status, body }
}

// ─── GET /site-api/plans ─────────────────────────────────────────────────────
// Returns enabled plans for this app (public endpoint, no auth required)
app.get('/site-api/plans', async (req, res) => {
  try {
    const { status, body } = await sk('/public/plans', {
      headers: { 'X-App-Key': APP_KEY },
    })
    res.status(status).json(body)
  } catch (e) {
    res.status(502).json({ code: 502, message: '无法连接到 SaaSKit 平台: ' + e.message })
  }
})

// ─── POST /site-api/auth/register ────────────────────────────────────────────
app.post('/site-api/auth/register', async (req, res) => {
  try {
    const { status, body } = await sk('/user-auth/register', {
      method: 'POST',
      body: JSON.stringify(req.body),
      headers: { 'X-App-Key': APP_KEY },
    })
    res.status(status).json(body)
  } catch (e) {
    res.status(502).json({ code: 502, message: e.message })
  }
})

// ─── POST /site-api/auth/login ───────────────────────────────────────────────
app.post('/site-api/auth/login', async (req, res) => {
  try {
    const { status, body } = await sk('/user-auth/login', {
      method: 'POST',
      body: JSON.stringify(req.body),
      headers: { 'X-App-Key': APP_KEY },
    })
    res.status(status).json(body)
  } catch (e) {
    res.status(502).json({ code: 502, message: e.message })
  }
})

// ─── GET /site-api/account/profile ───────────────────────────────────────────
app.get('/site-api/account/profile', async (req, res) => {
  try {
    const { status, body } = await sk('/account/profile', {
      headers: { Authorization: req.headers.authorization || '' },
    })
    res.status(status).json(body)
  } catch (e) {
    res.status(502).json({ code: 502, message: e.message })
  }
})

// ─── GET /site-api/account/subscription ──────────────────────────────────────
app.get('/site-api/account/subscription', async (req, res) => {
  try {
    const { status, body } = await sk('/account/subscription', {
      headers: { Authorization: req.headers.authorization || '' },
    })
    res.status(status).json(body)
  } catch (e) {
    res.status(502).json({ code: 502, message: e.message })
  }
})

// ─── GET /site-api/account/orders ────────────────────────────────────────────
app.get('/site-api/account/orders', async (req, res) => {
  try {
    const { status, body } = await sk('/account/orders', {
      headers: { Authorization: req.headers.authorization || '' },
    })
    res.status(status).json(body)
  } catch (e) {
    res.status(502).json({ code: 502, message: e.message })
  }
})

// ─── POST /site-api/orders ───────────────────────────────────────────────────
// Creates a payment order. Fetches user_id server-side from the user token.
app.post('/site-api/orders', async (req, res) => {
  const userToken = req.headers.authorization
  if (!userToken) {
    return res.status(401).json({ code: 401, message: '请先登录' })
  }

  try {
    // Verify user token and get real user_id from SaaSKit
    const profileRes = await sk('/account/profile', {
      headers: { Authorization: userToken },
    })
    if (profileRes.status !== 200 || profileRes.body.code !== 0) {
      return res.status(401).json({ code: 401, message: '登录状态已过期，请重新登录' })
    }
    const userId = profileRes.body.data?.id
    if (!userId) {
      return res.status(401).json({ code: 401, message: '无法获取用户信息' })
    }

    // Create order with server-side API credentials (user_id is verified server-side)
    const { status, body } = await sk('/client/orders', {
      method: 'POST',
      body: JSON.stringify({ ...req.body, user_id: userId }),
      headers: {
        'X-API-Key': API_KEY,
        'X-API-Secret': API_SECRET,
      },
    })
    res.status(status).json(body)
  } catch (e) {
    res.status(502).json({ code: 502, message: e.message })
  }
})

// ─── GET /site-api/orders/:orderNo ───────────────────────────────────────────
app.get('/site-api/orders/:orderNo', async (req, res) => {
  try {
    const { status, body } = await sk(`/client/orders/${encodeURIComponent(req.params.orderNo)}`, {
      headers: { 'X-API-Key': API_KEY, 'X-API-Secret': API_SECRET },
    })
    res.status(status).json(body)
  } catch (e) {
    res.status(502).json({ code: 502, message: e.message })
  }
})

// ─── POST /site-api/orders/:orderNo/mock-pay ─────────────────────────────────
// Triggers mock payment notification (only works when PAYMENT_MOCK=true)
app.post('/site-api/orders/:orderNo/mock-pay', async (req, res) => {
  const orderNo = req.params.orderNo
  const notifyURL = `${SAASKIT_URL}/api/payments/alipay/notify?mock=1&out_trade_no=${encodeURIComponent(orderNo)}`
  try {
    const r = await fetch(notifyURL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: '',
    })
    const text = await r.text()
    if (text === 'success') {
      res.json({ code: 0, data: { success: true }, message: '模拟支付成功' })
    } else {
      res.status(500).json({ code: 500, message: '模拟支付通知失败: ' + text })
    }
  } catch (e) {
    res.status(502).json({ code: 502, message: e.message })
  }
})

// ─── Serve built Vue SPA in production ───────────────────────────────────────
if (process.env.NODE_ENV === 'production') {
  app.use(express.static(path.join(__dirname, 'dist')))
  app.get('*', (req, res) => {
    res.sendFile(path.join(__dirname, 'dist', 'index.html'))
  })
}

app.listen(PORT, () => {
  console.log(`\n🚀 CloudSuite Site Backend: http://localhost:${PORT}`)
  console.log(`   SaaSKit Platform: ${SAASKIT_URL}`)
  console.log(`   App Key: ${APP_KEY ? APP_KEY.slice(0, 12) + '...' : '(not set)'}`)
  if (process.env.NODE_ENV !== 'production') {
    console.log(`\n   Vue Dev Server: http://localhost:3000  (run "npm run dev" to start both)\n`)
  }
})
