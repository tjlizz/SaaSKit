<script lang="ts" setup>
import type { Order, PaymentConfigStatus } from '#/api';

import { computed, onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Card, Col, Row, Statistic, Table, Tag } from 'ant-design-vue';

import {
  listApiClientsApi,
  listOrdersApi,
  listPaymentConfigsApi,
  listPlansApi,
  listSubscriptionsApi,
  listUsersApi,
} from '#/api';

import {
  formatDate,
  formatMoney,
  providerLabels,
  statusColors,
  statusLabels,
} from '../../saaskit/shared';

const loading = ref(false);
const metrics = reactive({
  apiClients: 0,
  orders: 0,
  plans: 0,
  subscriptions: 0,
  users: 0,
});
const orders = ref<Order[]>([]);
const paymentConfigs = ref<PaymentConfigStatus[]>([]);
const paidAmount = computed(() =>
  orders.value
    .filter((item) => item.status === 'paid')
    .reduce((sum, item) => sum + item.amount_cents, 0),
);
const columns = [
  { dataIndex: 'order_no', key: 'order_no', title: '订单号' },
  { key: 'user', title: '用户' },
  { key: 'amount', title: '金额' },
  { dataIndex: 'status', key: 'status', title: '状态' },
  { dataIndex: 'created_at', key: 'created_at', title: '创建时间' },
];

async function load() {
  loading.value = true;
  try {
    const [users, plans, allOrders, subscriptions, clients, payments] =
      await Promise.all([
        listUsersApi({ page: 1, page_size: 1 }),
        listPlansApi(),
        listOrdersApi(),
        listSubscriptionsApi(),
        listApiClientsApi(),
        listPaymentConfigsApi(),
      ]);
    metrics.users = users.total;
    metrics.plans = plans.length;
    metrics.orders = allOrders.length;
    metrics.subscriptions = subscriptions.filter(
      (item) => item.subscription_status === 'active',
    ).length;
    metrics.apiClients = clients.filter((item) => item.enabled).length;
    orders.value = allOrders;
    paymentConfigs.value = payments;
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <Page
    description="当前 SaaS 产品的用户、交易与基础能力运行概况"
    title="运营概览"
  >
    <Row :gutter="[16, 16]">
      <Col :lg="6" :sm="12" :xs="24">
        <Card :loading="loading">
          <Statistic title="用户总数" :value="metrics.users" />
        </Card>
      </Col>
      <Col :lg="6" :sm="12" :xs="24">
        <Card :loading="loading">
          <Statistic title="有效订阅" :value="metrics.subscriptions" />
        </Card>
      </Col>
      <Col :lg="6" :sm="12" :xs="24">
        <Card :loading="loading">
          <Statistic title="订单总数" :value="metrics.orders" />
        </Card>
      </Col>
      <Col :lg="6" :sm="12" :xs="24">
        <Card :loading="loading">
          <Statistic
            title="已支付金额"
            :precision="2"
            prefix="¥"
            :value="paidAmount / 100"
          />
        </Card>
      </Col>
    </Row>

    <Row :gutter="[16, 16]" class="mt-4">
      <Col :lg="16" :xs="24">
        <Card title="最近订单">
          <Table
            row-key="id"
            :columns="columns"
            :data-source="orders.slice(0, 8)"
            :loading="loading"
            :pagination="false"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'order_no'">
                <span class="font-mono text-xs">{{ record.order_no }}</span>
              </template>
              <template v-else-if="column.key === 'user'">
                {{ record.user?.name || record.user?.email || record.user_id }}
              </template>
              <template v-else-if="column.key === 'amount'">
                {{ formatMoney(record.amount_cents, record.currency) }}
              </template>
              <template v-else-if="column.key === 'status'">
                <Tag :color="statusColors[record.status]">
                  {{ statusLabels[record.status] || record.status }}
                </Tag>
              </template>
              <template v-else-if="column.key === 'created_at'">
                {{ formatDate(record.created_at) }}
              </template>
            </template>
          </Table>
        </Card>
      </Col>
      <Col :lg="8" :xs="24">
        <Card :loading="loading" title="接入状态">
          <div class="divide-y divide-gray-100 dark:divide-gray-700">
            <div class="flex items-center justify-between py-3">
              <span>销售套餐</span>
              <Tag color="blue">{{ metrics.plans }} 个</Tag>
            </div>
            <div class="flex items-center justify-between py-3">
              <span>可用 API 凭证</span>
              <Tag color="blue">{{ metrics.apiClients }} 个</Tag>
            </div>
            <div
              v-for="item in paymentConfigs"
              :key="item.provider"
              class="flex items-center justify-between py-3"
            >
              <span>{{ providerLabels[item.provider] }}</span>
              <Tag
                :color="item.configured && item.enabled ? 'green' : 'default'"
              >
                {{
                  item.configured
                    ? item.enabled
                      ? '已启用'
                      : '已配置'
                    : '未配置'
                }}
              </Tag>
            </div>
            <div
              v-if="paymentConfigs.length === 0"
              class="py-5 text-center text-gray-400"
            >
              暂未配置支付渠道
            </div>
          </div>
        </Card>
      </Col>
    </Row>
  </Page>
</template>
