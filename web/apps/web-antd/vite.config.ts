import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api': {
            changeOrigin: true,
            // SaaSKit Go API，本地请求保留 /api 前缀直接转发。
            target: 'http://localhost:8080',
            ws: true,
          },
        },
      },
    },
  };
});
