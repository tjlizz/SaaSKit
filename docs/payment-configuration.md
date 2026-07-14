# 支付配置

先将 `.env` 中 `PAYMENT_MOCK=false`，并确保 `PUBLIC_URL` 是支付平台可访问的 HTTPS 域名。支付回调地址由后端自动生成：

- 支付宝：`{PUBLIC_URL}/api/payments/alipay/notify`
- 微信：`{PUBLIC_URL}/api/payments/wechat/notify`

配置接口需要开发者 JWT：

```http
PUT /api/admin/payment-configs/{provider}
Authorization: Bearer <admin-access-token>
Content-Type: application/json
```

## 支付宝

```json
{
  "enabled": true,
  "sandbox": true,
  "config": {
    "app_id": "支付宝应用 ID",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
    "public_key": "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----"
  }
}
```

`private_key` 是应用 RSA2 私钥，`public_key` 是支付宝公钥，不是应用公钥。两者都使用 PEM 格式。网站支付使用订单参数 `channel=page`，当面付使用 `channel=qr`。

## 微信支付 API v3

```json
{
  "enabled": true,
  "sandbox": false,
  "config": {
    "app_id": "公众号或小程序 AppID",
    "mch_id": "商户号",
    "serial_no": "商户 API 证书序列号",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
    "api_v3_key": "32 字节 API v3 Key",
    "platform_public_key": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
  }
}
```

`platform_public_key` 可填写微信支付平台证书 PEM 或平台 RSA 公钥 PEM，用于回调签名校验。当前实现支持 Native 二维码支付。

## 密钥存储

配置提交后使用 `PAYMENT_CONFIG_KEY` 派生的 256 位密钥进行 AES-GCM 加密，管理查询接口只返回 `configured/enabled/sandbox`，不会回传明文。更换 `PAYMENT_CONFIG_KEY` 前必须先重新提交所有支付配置，否则旧密文将无法解密。
