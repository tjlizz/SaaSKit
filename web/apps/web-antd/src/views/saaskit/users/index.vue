<script lang="ts" setup>
import type { User, UserInput } from '#/api';

import { h, onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Form,
  FormItem,
  Input,
  message,
  Modal,
  Select,
  Space,
  Table,
  Tag,
} from 'ant-design-vue';

import {
  createUserApi,
  deleteUserApi,
  listUsersApi,
  resetUserPasswordApi,
  setUserStatusApi,
  updateUserApi,
} from '#/api';

import { formatDate, statusColors, statusLabels } from '../shared';

const loading = ref(false);
const items = ref<User[]>([]);
const total = ref(0);
const query = reactive({ page: 1, page_size: 20, q: '', status: '' });
const editorOpen = ref(false);
const editorSaving = ref(false);
const editingId = ref('');
const form = reactive<UserInput>({
  email: '',
  name: '',
  password: '',
  phone: '',
  status: 'active',
});

const columns = [
  { dataIndex: 'name', key: 'name', title: '用户' },
  { dataIndex: 'phone', key: 'phone', title: '手机' },
  { dataIndex: 'status', key: 'status', title: '状态', width: 100 },
  { dataIndex: 'last_login_at', key: 'last_login_at', title: '最近登录' },
  { dataIndex: 'created_at', key: 'created_at', title: '注册时间' },
  { key: 'actions', title: '操作', width: 250 },
];

async function load() {
  loading.value = true;
  try {
    const result = await listUsersApi({
      ...query,
      q: query.q || undefined,
      status: query.status || undefined,
    });
    items.value = result.items;
    total.value = result.total;
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingId.value = '';
  Object.assign(form, {
    email: '',
    name: '',
    password: '',
    phone: '',
    status: 'active',
  });
  editorOpen.value = true;
}

function openEdit(item: Record<string, any>) {
  editingId.value = item.id;
  Object.assign(form, {
    email: item.email,
    name: item.name,
    password: '',
    phone: item.phone || '',
    status: item.status,
  });
  editorOpen.value = true;
}

async function save() {
  if (!form.email || !form.name || (!editingId.value && !form.password)) {
    message.warning('请填写名称、邮箱和初始密码');
    return;
  }
  if (!editingId.value && (form.password?.length || 0) < 8) {
    message.warning('初始密码至少 8 位');
    return;
  }
  editorSaving.value = true;
  try {
    if (editingId.value) {
      await updateUserApi(editingId.value, {
        email: form.email,
        name: form.name,
        phone: form.phone,
      });
      await setUserStatusApi(editingId.value, form.status || 'active');
    } else {
      await createUserApi(form);
    }
    message.success(editingId.value ? '用户已更新' : '用户已创建');
    editorOpen.value = false;
    await load();
  } finally {
    editorSaving.value = false;
  }
}

function resetPassword(item: Record<string, any>) {
  let password = '';
  Modal.confirm({
    title: `重置 ${item.name} 的密码`,
    content: () =>
      h(Input.Password, {
        'onUpdate:value': (value: string) => (password = value),
        placeholder: '输入至少 8 位新密码',
      }),
    async onOk() {
      if (password.length < 8) {
        message.warning('密码至少 8 位');
        throw new Error('invalid password');
      }
      await resetUserPasswordApi(item.id, password);
      message.success('密码已重置');
    },
  });
}

function remove(item: Record<string, any>) {
  Modal.confirm({
    title: `确认删除用户“${item.name}”？`,
    content: '已有订单的用户不能删除，可改为停用。',
    okButtonProps: { danger: true },
    async onOk() {
      await deleteUserApi(item.id);
      message.success('用户已删除');
      await load();
    },
  });
}

async function toggleStatus(item: Record<string, any>) {
  await setUserStatusApi(
    item.id,
    item.status === 'active' ? 'disabled' : 'active',
  );
  message.success(item.status === 'active' ? '用户已停用' : '用户已启用');
  await load();
}

function search() {
  query.page = 1;
  void load();
}

function changePage(page: number, pageSize: number) {
  query.page = page;
  query.page_size = pageSize;
  void load();
}

onMounted(load);
</script>

<template>
  <Page
    description="管理本 SaaS 产品自己的用户、登录状态与账号安全"
    title="用户管理"
  >
    <Card>
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <Space wrap>
          <Input
            v-model:value="query.q"
            allow-clear
            placeholder="搜索名称、邮箱或手机"
            style="width: 260px"
            @press-enter="search"
          />
          <Select
            v-model:value="query.status"
            :options="[
              { label: '全部状态', value: '' },
              { label: '正常', value: 'active' },
              { label: '已停用', value: 'disabled' },
            ]"
            style="width: 130px"
            @change="search"
          />
          <Button @click="search">查询</Button>
        </Space>
        <Button type="primary" @click="openCreate">新建用户</Button>
      </div>

      <Table
        row-key="id"
        :columns="columns"
        :data-source="items"
        :loading="loading"
        :pagination="{
          current: query.page,
          pageSize: query.page_size,
          showSizeChanger: true,
          total,
        }"
        @change="
          (pagination) =>
            changePage(pagination.current || 1, pagination.pageSize || 20)
        "
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <div class="font-medium">{{ record.name }}</div>
            <div class="text-xs text-gray-500">{{ record.email }}</div>
          </template>
          <template v-else-if="column.key === 'phone'">
            {{ record.phone || '-' }}
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusColors[record.status]">
              {{ statusLabels[record.status] }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'last_login_at'">
            {{ formatDate(record.last_login_at) }}
          </template>
          <template v-else-if="column.key === 'created_at'">
            {{ formatDate(record.created_at) }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <Space size="small">
              <Button size="small" type="link" @click="openEdit(record)">
                编辑
              </Button>
              <Button size="small" type="link" @click="toggleStatus(record)">
                {{ record.status === 'active' ? '停用' : '启用' }}
              </Button>
              <Button size="small" type="link" @click="resetPassword(record)">
                重置密码
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
      :confirm-loading="editorSaving"
      :title="editingId ? '编辑用户' : '新建用户'"
      @ok="save"
    >
      <Form layout="vertical">
        <FormItem label="名称" required>
          <Input v-model:value="form.name" />
        </FormItem>
        <FormItem label="邮箱" required>
          <Input v-model:value="form.email" />
        </FormItem>
        <FormItem label="手机"><Input v-model:value="form.phone" /></FormItem>
        <FormItem v-if="!editingId" label="初始密码" required>
          <Input.Password
            v-model:value="form.password"
            placeholder="至少 8 位"
          />
        </FormItem>
        <FormItem label="状态">
          <Select
            v-model:value="form.status"
            :options="[
              { label: '正常', value: 'active' },
              { label: '停用', value: 'disabled' },
            ]"
          />
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>
