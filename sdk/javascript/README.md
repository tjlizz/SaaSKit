# SaaSKit JavaScript SDK

仅用于服务端 Node.js；不要把 `apiSecret` 暴露到浏览器。

```js
import { SaaSKit } from '@saaskit-community/sdk';

const client = new SaaSKit({
  baseURL: 'https://billing.example.com',
  apiKey: process.env.SAASKIT_API_KEY,
  apiSecret: process.env.SAASKIT_API_SECRET,
});

const subscription = await client.checkSubscription('internal-user-uuid');
```
