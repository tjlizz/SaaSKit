<script lang="ts" setup>
import type { Application } from '#/api';

import { onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Form,
  FormItem,
  Input,
  message,
  Modal,
  Space,
  Switch,
  Table,
  Tag,
} from 'ant-design-vue';

import {
  createApplicationApi,
  deleteApplicationApi,
  listApplicationsApi,
  updateApplicationApi,
} from '#/api';
import {
  getCurrentApplicationId,
  setCurrentApplicationId,
} from '#/utils/application-context';

import { formatDate } from '../shared';

const loading = ref(false);
const saving = ref(false);
const items = ref<Application[]>([]);
const editorOpen = ref(false);
const editingId = ref('');
const form = reactive({ description: '', enabled: true, name: '' });
const columns = [
  { dataIndex: 'name', key: 'name', title: '应用名称' },
  { dataIndex: 'app_key', key: 'app_key', title: 'App Key' },
  { dataIndex: 'status', key: 'status', title: '状态', width: 90 },
  { dataIndex: 'created_at', key: 'created_at', title: '创建时间' },
  { key: 'actions', title: '操作', width: 180 },
];

async function load() {
  loading.value = true;
  try {
    items.value = await listApplicationsApi();
    const current = getCurrentApplicationId();
    if (
      !items.value.some(
        (item) => item.id === current && item.status === 'active',
      )
    ) {
      setCurrentApplicationId(
        items.value.find((item) => item.status === 'active')?.id ?? '',
      );
    }
  } finally {
    loading.value = false;
  }
}

function create() {
  editingId.value = '';
  Object.assign(form, { description: '', enabled: true, name: '' });
  editorOpen.value = true;
}

function edit(item: Record<string, any>) {
  editingId.value = item.id;
  Object.assign(form, {
    description: item.description,
    enabled: item.status === 'active',
    name: item.name,
  });
  editorOpen.value = true;
}

async function save() {
  if (!form.name.trim()) {
    message.warning('请输入应用名称');
    return;
  }
  saving.value = true;
  try {
    const data = {
      description: form.description,
      name: form.name,
      status: (form.enabled ? 'active' : 'disabled') as Application['status'],
    };
    const item = editingId.value
      ? await updateApplicationApi(editingId.value, data)
      : await createApplicationApi(data);
    if (!getCurrentApplicationId() && item.status === 'active') {
      setCurrentApplicationId(item.id);
    }
    editorOpen.value = false;
    message.success(editingId.value ? '应用已更新' : '应用已创建');
    await load();
  } finally {
    saving.value = false;
  }
}

function remove(item: Record<string, any>) {
  Modal.confirm({
    content: '仅没有用户的应用可以删除；已有数据的应用请改为停用。',
    okButtonProps: { danger: true },
    title: `删除应用“${item.name}”？`,
    async onOk() {
      await deleteApplicationApi(item.id);
      if (getCurrentApplicationId() === item.id) setCurrentApplicationId('');
      message.success('应用已删除');
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
    description="一个部署可管理多个 SaaS 应用，用户和商业数据按应用隔离。"
    title="应用管理"
  >
    <Card>
      <div class="mb-4 flex justify-end">
        <Button type="primary" @click="create">新建应用</Button>
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
            <div>
              <div class="font-medium">{{ record.name }}</div>
              <div class="text-muted-foreground text-xs">
                {{ record.description }}
              </div>
            </div>
          </template>
          <template v-else-if="column.key === 'app_key'">
            <Space>
              <code>{{ record.app_key }}</code>
              <Button size="small" type="link" @click="copy(record.app_key)">
                复制
              </Button>
            </Space>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="record.status === 'active' ? 'green' : 'default'">
              {{ record.status === 'active' ? '启用' : '停用' }}
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
      :title="editingId ? '编辑应用' : '新建应用'"
      @ok="save"
    >
      <Form layout="vertical">
        <FormItem label="应用名称" required>
          <Input v-model:value="form.name" placeholder="例如：在线协作工具" />
        </FormItem>
        <FormItem label="说明">
          <Input.TextArea v-model:value="form.description" :rows="3" />
        </FormItem>
        <FormItem label="启用">
          <Switch v-model:checked="form.enabled" />
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>
