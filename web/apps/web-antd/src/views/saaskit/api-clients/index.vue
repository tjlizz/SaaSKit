<script lang="ts" setup>
import type { ApiClient } from '#/api';

import { onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Alert,
  Button,
  Card,
  Form,
  FormItem,
  Input,
  InputNumber,
  message,
  Modal,
  Space,
  Switch,
  Table,
  Tag,
} from 'ant-design-vue';

import {
  createApiClientApi,
  deleteApiClientApi,
  listApiClientsApi,
  rotateApiClientSecretApi,
  updateApiClientApi,
} from '#/api';

import { formatDate } from '../shared';

const loading = ref(false);
const saving = ref(false);
const items = ref<ApiClient[]>([]);
const editorOpen = ref(false);
const editingId = ref('');
const secretOpen = ref(false);
const secret = ref('');
const form = reactive({ enabled: true, name: '', rate_limit_per_min: 300 });
const columns = [
  { dataIndex: 'name', key: 'name', title: '凭证名称' },
  { dataIndex: 'client_key', key: 'client_key', title: 'Client Key' },
  {
    dataIndex: 'rate_limit_per_min',
    key: 'limit',
    title: '每分钟限流',
    width: 120,
  },
  { dataIndex: 'enabled', key: 'enabled', title: '状态', width: 90 },
  { dataIndex: 'created_at', key: 'created_at', title: '创建时间' },
  { key: 'actions', title: '操作', width: 220 },
];

async function load() {
  loading.value = true;
  try {
    items.value = await listApiClientsApi();
  } finally {
    loading.value = false;
  }
}

function create() {
  editingId.value = '';
  Object.assign(form, { enabled: true, name: '', rate_limit_per_min: 300 });
  editorOpen.value = true;
}

function edit(item: Record<string, any>) {
  editingId.value = item.id;
  Object.assign(form, {
    enabled: item.enabled,
    name: item.name,
    rate_limit_per_min: item.rate_limit_per_min,
  });
  editorOpen.value = true;
}

function showSecret(value: string) {
  secret.value = value;
  secretOpen.value = true;
}

async function save() {
  if (!form.name) {
    message.warning('请输入凭证名称');
    return;
  }
  saving.value = true;
  try {
    if (editingId.value) {
      await updateApiClientApi(editingId.value, { ...form });
      message.success('API 凭证已更新');
    } else {
      const result = await createApiClientApi({ ...form });
      showSecret(result.client_secret);
      message.success('API 凭证已创建');
    }
    editorOpen.value = false;
    await load();
  } finally {
    saving.value = false;
  }
}

function rotate(item: Record<string, any>) {
  Modal.confirm({
    title: `轮换“${item.name}”的密钥？`,
    content: '旧密钥将立即失效，请确保业务服务能够及时更新。',
    async onOk() {
      const result = await rotateApiClientSecretApi(item.id);
      showSecret(result.client_secret);
    },
  });
}

function remove(item: Record<string, any>) {
  Modal.confirm({
    title: `删除 API 凭证“${item.name}”？`,
    okButtonProps: { danger: true },
    async onOk() {
      await deleteApiClientApi(item.id);
      message.success('API 凭证已删除');
      await load();
    },
  });
}

async function copy(value: string) {
  await navigator.clipboard.writeText(value);
  message.success('已复制');
}

onMounted(load);
</script>

<template>
  <Page
    description="供你的业务后端调用订单和订阅查询接口，密钥不会暴露给浏览器用户"
    title="API 凭证"
  >
    <Card>
      <div class="mb-4 flex justify-end">
        <Button type="primary" @click="create">新建凭证</Button>
      </div>
      <Table
        row-key="id"
        :columns="columns"
        :data-source="items"
        :loading="loading"
        :pagination="false"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <span class="font-medium">{{ record.name }}</span>
          </template>
          <template v-else-if="column.key === 'client_key'">
            <div class="flex items-center gap-2">
              <code>{{ record.client_key }}</code>
              <Button size="small" type="link" @click="copy(record.client_key)">
                复制
              </Button>
            </div>
          </template>
          <template v-else-if="column.key === 'enabled'">
            <Tag :color="record.enabled ? 'green' : 'default'">
              {{ record.enabled ? '启用' : '停用' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'created_at'">
            {{ formatDate(record.created_at) }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <Space>
              <Button size="small" type="link" @click="edit(record)">
                编辑
              </Button>
              <Button size="small" type="link" @click="rotate(record)">
                轮换密钥
              </Button>
              <Button danger size="small" type="link" @click="remove(record)">
                删除
              </Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="editorOpen"
      :confirm-loading="saving"
      :title="editingId ? '编辑 API 凭证' : '新建 API 凭证'"
      @ok="save"
    >
      <Form layout="vertical">
        <FormItem label="名称" required>
          <Input v-model:value="form.name" placeholder="例如：生产业务后端" />
        </FormItem>
        <FormItem label="每分钟请求上限">
          <InputNumber
            v-model:value="form.rate_limit_per_min"
            :min="1"
            class="w-full"
          />
        </FormItem>
        <FormItem label="启用">
          <Switch v-model:checked="form.enabled" />
        </FormItem>
      </Form>
    </Modal>

    <Modal
      v-model:open="secretOpen"
      :footer="null"
      title="请立即保存 Client Secret"
    >
      <Alert
        class="mb-4"
        message="该密钥只展示一次，关闭后无法再次查看，只能重新轮换。"
        show-icon
        type="warning"
      />
      <div
        class="rounded bg-gray-100 p-3 font-mono text-sm break-all dark:bg-gray-800"
      >
        {{ secret }}
      </div>
      <Button block class="mt-4" type="primary" @click="copy(secret)">
        复制密钥
      </Button>
    </Modal>
  </Page>
</template>
