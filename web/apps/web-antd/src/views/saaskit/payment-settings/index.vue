<script lang="ts" setup>
import type { PaymentConfigStatus } from '#/api';

import { onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Alert,
  Button,
  Card,
  Form,
  FormItem,
  Input,
  message,
  Space,
  Switch,
  Tag,
} from 'ant-design-vue';

import { listPaymentConfigsApi, savePaymentConfigApi } from '#/api';

import { formatDate } from '../shared';

type Provider = PaymentConfigStatus['provider'];

const loading = ref(false);
const saving = reactive<Record<Provider, boolean>>({
  alipay: false,
  wechat: false,
});
const statuses = reactive<Record<Provider, PaymentConfigStatus>>({
  alipay: {
    configured: false,
    enabled: false,
    provider: 'alipay',
    sandbox: true,
  },
  wechat: {
    configured: false,
    enabled: false,
    provider: 'wechat',
    sandbox: true,
  },
});
const forms = reactive({
  alipay: {
    app_id: '',
    gateway: '',
    private_key: '',
    public_key: '',
  },
  wechat: {
    api_v3_key: '',
    app_id: '',
    mch_id: '',
    platform_public_key: '',
    private_key: '',
    serial_no: '',
  },
});

async function load() {
  loading.value = true;
  try {
    const result = await listPaymentConfigsApi();
    for (const item of result) Object.assign(statuses[item.provider], item);
  } finally {
    loading.value = false;
  }
}

async function save(provider: Provider) {
  const status = statuses[provider];
  const allFields = { ...forms[provider] };
  const config = Object.fromEntries(
    Object.entries(allFields).filter(([, value]) => value.trim() !== ''),
  );
  if (
    status.enabled &&
    !status.configured &&
    Object.values(allFields).some((value) => !value)
  ) {
    const optionalKeys = provider === 'alipay' ? ['gateway'] : [];
    const hasMissingRequired = Object.entries(allFields).some(
      ([key, value]) => !value && !optionalKeys.includes(key),
    );
    if (hasMissingRequired) {
      message.warning('启用支付前请填写完整商户配置');
      return;
    }
  }
  saving[provider] = true;
  try {
    const result = await savePaymentConfigApi(provider, {
      config,
      enabled: status.enabled,
      sandbox: status.sandbox,
    });
    Object.assign(statuses[provider], result);
    message.success(
      provider === 'alipay' ? '支付宝配置已保存' : '微信支付配置已保存',
    );
    Object.keys(allFields).forEach((key) => {
      forms[provider][key as keyof (typeof forms)[Provider]] = '';
    });
    await load();
  } finally {
    saving[provider] = false;
  }
}

onMounted(load);
</script>

<template>
  <Page
    description="商户密钥加密存储在当前实例数据库中，不会发送到 SaaSKit 之外"
    title="支付配置"
  >
    <Alert
      class="mb-4"
      message="出于安全原因，已保存的密钥不会回显。留空字段会保留原值，只有重新填写的字段才会更新。"
      show-icon
      type="info"
    />
    <div class="grid grid-cols-1 gap-5 xl:grid-cols-2">
      <Card :loading="loading" title="支付宝">
        <template #extra>
          <Space>
            <Tag :color="statuses.alipay.configured ? 'green' : 'default'">
              {{ statuses.alipay.configured ? '已配置' : '未配置' }}
            </Tag>
            <span
              v-if="statuses.alipay.updated_at"
              class="text-xs text-gray-500"
            >
              {{ formatDate(statuses.alipay.updated_at) }}
            </span>
          </Space>
        </template>
        <Form layout="vertical">
          <div class="mb-4 flex flex-wrap gap-6">
            <Space>
              启用 <Switch v-model:checked="statuses.alipay.enabled" />
            </Space>
            <Space>
              沙箱环境 <Switch v-model:checked="statuses.alipay.sandbox" />
            </Space>
          </div>
          <FormItem label="App ID" required>
            <Input v-model:value="forms.alipay.app_id" />
          </FormItem>
          <FormItem label="应用私钥" required>
            <Input.TextArea
              v-model:value="forms.alipay.private_key"
              :rows="5"
              placeholder="PKCS#8 PEM 私钥"
            />
          </FormItem>
          <FormItem label="支付宝公钥" required>
            <Input.TextArea v-model:value="forms.alipay.public_key" :rows="5" />
          </FormItem>
          <FormItem label="自定义网关（可选）">
            <Input
              v-model:value="forms.alipay.gateway"
              placeholder="留空使用官方网关"
            />
          </FormItem>
          <Button
            block
            :loading="saving.alipay"
            type="primary"
            @click="save('alipay')"
          >
            保存支付宝配置
          </Button>
        </Form>
      </Card>

      <Card :loading="loading" title="微信支付">
        <template #extra>
          <Space>
            <Tag :color="statuses.wechat.configured ? 'green' : 'default'">
              {{ statuses.wechat.configured ? '已配置' : '未配置' }}
            </Tag>
            <span
              v-if="statuses.wechat.updated_at"
              class="text-xs text-gray-500"
            >
              {{ formatDate(statuses.wechat.updated_at) }}
            </span>
          </Space>
        </template>
        <Form layout="vertical">
          <div class="mb-4 flex flex-wrap gap-6">
            <Space>
              启用 <Switch v-model:checked="statuses.wechat.enabled" />
            </Space>
            <Space>
              沙箱标记 <Switch v-model:checked="statuses.wechat.sandbox" />
            </Space>
          </div>
          <div class="grid grid-cols-1 gap-x-4 md:grid-cols-2">
            <FormItem label="App ID" required>
              <Input v-model:value="forms.wechat.app_id" />
            </FormItem>
            <FormItem label="商户号" required>
              <Input v-model:value="forms.wechat.mch_id" />
            </FormItem>
            <FormItem label="证书序列号" required>
              <Input v-model:value="forms.wechat.serial_no" />
            </FormItem>
            <FormItem label="API v3 Key" required>
              <Input.Password v-model:value="forms.wechat.api_v3_key" />
            </FormItem>
          </div>
          <FormItem label="商户 API 私钥" required>
            <Input.TextArea
              v-model:value="forms.wechat.private_key"
              :rows="5"
            />
          </FormItem>
          <FormItem label="微信支付平台公钥" required>
            <Input.TextArea
              v-model:value="forms.wechat.platform_public_key"
              :rows="5"
            />
          </FormItem>
          <Button
            block
            :loading="saving.wechat"
            type="primary"
            @click="save('wechat')"
          >
            保存微信支付配置
          </Button>
        </Form>
      </Card>
    </div>
  </Page>
</template>
