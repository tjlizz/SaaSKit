import { fileURLToPath } from 'node:url';

import { defineConfig } from '@vben/vite-config';

import { loadEnv } from 'vite';

export default defineConfig(async ({ mode }) => {
  const projectRoot = fileURLToPath(new URL('../../../', import.meta.url));
  const rootEnv = loadEnv(mode, projectRoot, '');
  const apiPort = rootEnv.PORT || '8080';

  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api': {
            changeOrigin: true,
            // SaaSKit Go API，本地请求保留 /api 前缀直接转发。
            target: `http://localhost:${apiPort}`,
            ws: true,
          },
        },
      },
    },
  };
});
