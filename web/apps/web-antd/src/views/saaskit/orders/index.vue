<script lang="ts" setup>
import type { Application, Order } from '#/api';

import { computed, onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Input,
  message,
  Modal,
  Select,
  Space,
  Table,
  Tag,
} from 'ant-design-vue';

import { listApplicationsApi, listOrdersApi, updateOrderStatusApi } from '#/api';
import {
  getCurrentApplicationId,
  setCurrentApplicationId,
} from '#/utils/application-context';

import {
  formatDate,
  formatMoney,
  providerLabels,
  statusColors,
  statusLabels,
} from '../shared';

const loading = ref(false);
const items = ref<Order[]>([]);
const applications = ref<Application[]>([]);
const selectedAppId = ref(getCurrentApplicationId());
const appOptions = computed(() =>
  applications.value
    .filter((a) => a.status === 'active')
    .map((a) => ({ label: a.name, value: a.id })),
);
const query = reactive({ status: '', user_id: '' });
const columns = [
  { dataIndex: 'order_no', key: 'order_no', title: '订单号' },
  { key: 'user', title: '用户' },
  { key: 'amount', title: '金额', width: 130 },
  { dataIndex: 'provider', key: 'provider', title: '支付方式', width: 110 },
  { dataIndex: 'status', key: 'status', title: '状态', width: 100 },
  { dataIndex: 'created_at', key: 'created_at', title: '创建时间' },
  { key: 'actions', title: '操作', width: 120 },
];

async function load() {
  loading.value = true;
  try {
    applications.value = await listApplicationsApi();
    const active = applications.value.filter((a) => a.status === 'active');
    if (!active.some((a) => a.id === selectedAppId.value)) {
      selectedAppId.value = active[0]?.id ?? '';
      setCurrentApplicationId(selectedAppId.value);
    }
    items.value = await listOrdersApi({
      status: query.status || undefined,
      user_id: query.user_id || undefined,
    });
  } finally {
    loading.value = false;
  }
}

function switchApp(id: string) {
  selectedAppId.value = id;
  setCurrentApplicationId(id);
  load();
}

function changeStatus(
  item: Record<string, any>,
  status: 'cancelled' | 'refunded',
) {
  Modal.confirm({
    title:
      status === 'cancelled'
        ? '确认取消该订单？'
        : '确认将该订单标记为已退款？',
    content:
      status === 'refunded'
        ? '此操作会同时终止该订单生成的订阅。实际资金退款仍需在支付平台完成。'
        : '仅待支付订单可以取消。',
    okButtonProps: { danger: true },
    async onOk() {
      await updateOrderStatusApi(item.id, status);
      message.success(status === 'cancelled' ? '订单已取消' : '订单已标记退款');
      await load();
    },
  });
}

onMounted(load);
</script>

<template>
  <Page
    description="查看支付订单并处理取消、退款后的本地业务状态"
    title="订单管理"
  >
    <Card>
      <div class="mb-4 flex flex-wrap items-center gap-3">
        <Select
          :options="appOptions"
          :value="selectedAppId || undefined"
          class="w-44"
          placeholder="请选择应用"
          @change="switchApp"
        />
        <Input
          v-model:value="query.user_id"
          allow-clear
          placeholder="按用户 ID 查询"
          style="width: 260px"
          @press-enter="load"
        />
        <Select
          v-model:value="query.status"
          :options="[
            { label: '全部状态', value: '' },
            { label: '待支付', value: 'pending' },
            { label: '已支付', value: 'paid' },
            { label: '已取消', value: 'cancelled' },
            { label: '已退款', value: 'refunded' },
          ]"
          style="width: 130px"
          @change="load"
        />
        <Button @click="load">查询</Button>
      </div>
      <Table
        row-key="id"
        :columns="columns"
        :data-source="items"
        :loading="loading"
        :pagination="{ pageSize: 20, showSizeChanger: true }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'order_no'">
            <div class="font-mono text-xs">{{ record.order_no }}</div>
            <div v-if="record.provider_trade_no" class="text-xs text-gray-500">
              平台单号：{{ record.provider_trade_no }}
            </div>
          </template>
          <template v-else-if="column.key === 'user'">
            <div>{{ record.user?.name || '-' }}</div>
            <div class="text-xs text-gray-500">
              {{ record.user?.email || record.user_id }}
            </div>
          </template>
          <template v-else-if="column.key === 'amount'">
            {{ formatMoney(record.amount_cents, record.currency) }}
          </template>
          <template v-else-if="column.key === 'provider'">
            {{ providerLabels[record.provider] || record.provider }}
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusColors[record.status]">
              {{ statusLabels[record.status] || record.status }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'created_at'">
            {{ formatDate(record.created_at) }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <Space>
              <Button
                v-if="record.status === 'pending'"
                danger
                size="small"
                type="link"
                @click="changeStatus(record, 'cancelled')"
              >
                取消
              </Button>
              <Button
                v-if="record.status === 'paid'"
                danger
                size="small"
                type="link"
                @click="changeStatus(record, 'refunded')"
              >
                标记退款
              </Button>
              <span
                v-if="!['pending', 'paid'].includes(record.status)"
                class="text-gray-400"
              >
                -
              </span>
            </Space>
          </template>
        </template>
      </Table>
    </Card>
  </Page>
</template>
