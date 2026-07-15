<template>
  <div id="app-root">
    <AppHeader />
    <RouterView />
    <footer class="footer">
      <div class="footer-inner">
        <span>© 2025 CloudSuite. 由 <strong>SaaSKit</strong> 提供技术支持</span>
      </div>
    </footer>
  </div>
</template>

<script setup>
import AppHeader from './components/AppHeader.vue'
</script>

<style>
/* ── Global Variables ── */
:root {
  --primary: #4f46e5;
  --primary-hover: #4338ca;
  --primary-light: #ede9fe;
  --success: #10b981;
  --danger: #ef4444;
  --warning: #f59e0b;
  --bg: #f9fafb;
  --card: #ffffff;
  --text: #111827;
  --muted: #6b7280;
  --border: #e5e7eb;
  --radius: 12px;
  --shadow: 0 1px 3px rgba(0,0,0,.1), 0 1px 2px rgba(0,0,0,.06);
  --shadow-md: 0 4px 6px -1px rgba(0,0,0,.1), 0 2px 4px -1px rgba(0,0,0,.06);
  --shadow-lg: 0 10px 15px -3px rgba(0,0,0,.1), 0 4px 6px -2px rgba(0,0,0,.05);
}

/* ── Utilities ── */
.container { max-width: 1200px; margin: 0 auto; padding: 0 24px; }
.btn {
  display: inline-flex; align-items: center; justify-content: center;
  gap: 8px; padding: 10px 20px; border-radius: 8px; font-size: 14px;
  font-weight: 600; cursor: pointer; border: none; transition: all .15s;
  text-decoration: none; white-space: nowrap;
}
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover { background: var(--primary-hover); }
.btn-primary:disabled { background: #a5b4fc; cursor: not-allowed; }
.btn-outline { background: transparent; color: var(--primary); border: 1.5px solid var(--primary); }
.btn-outline:hover { background: var(--primary-light); }
.btn-ghost { background: transparent; color: var(--muted); border: 1.5px solid var(--border); }
.btn-ghost:hover { border-color: #9ca3af; color: var(--text); }
.btn-lg { padding: 14px 28px; font-size: 16px; border-radius: 10px; }
.btn-full { width: 100%; }
.card {
  background: var(--card); border-radius: var(--radius);
  box-shadow: var(--shadow); border: 1px solid var(--border);
}
.badge {
  display: inline-block; padding: 3px 10px; border-radius: 999px;
  font-size: 12px; font-weight: 600;
}
.badge-primary { background: var(--primary-light); color: var(--primary); }
.badge-success { background: #d1fae5; color: #065f46; }
.badge-warning { background: #fef3c7; color: #92400e; }
.badge-danger  { background: #fee2e2; color: #991b1b; }
.badge-gray    { background: #f3f4f6; color: #374151; }
.alert {
  padding: 12px 16px; border-radius: 8px; font-size: 14px;
  display: flex; align-items: flex-start; gap: 10px;
}
.alert-error { background: #fee2e2; color: #991b1b; border: 1px solid #fecaca; }
.alert-success { background: #d1fae5; color: #065f46; border: 1px solid #a7f3d0; }
.alert-info { background: #eff6ff; color: #1e40af; border: 1px solid #bfdbfe; }
.spinner {
  width: 20px; height: 20px; border: 2.5px solid rgba(255,255,255,.3);
  border-top-color: #fff; border-radius: 50%; animation: spin .7s linear infinite;
}
.spinner-dark {
  border-color: rgba(79,70,229,.2); border-top-color: var(--primary);
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Footer ── */
.footer {
  border-top: 1px solid var(--border);
  background: var(--card);
  padding: 24px 0;
  margin-top: auto;
  text-align: center;
}
.footer-inner { color: var(--muted); font-size: 14px; }
</style>
