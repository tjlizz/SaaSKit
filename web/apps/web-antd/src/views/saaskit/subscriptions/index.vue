<script lang="ts" setup>
import type { Subscription } from '#/api';

import { onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Button, Card, Input, Table, Tag } from 'ant-design-vue';

import { listSubscriptionsApi } from '#/api';

import { formatDate, statusColors, statusLabels } from '../shared';

const loading = ref(false);
const items = ref<Subscription[]>([]);
const userId = ref('');
const columns = [
  { key: 'user', title: '用户' },
  { key: 'plan', title: '当前套餐' },
  {
    dataIndex: 'subscription_status',
    key: 'status',
    title: '状态',
    width: 100,
  },
  { dataIndex: 'current_period_start', key: 'start', title: '开始时间' },
  { dataIndex: 'current_period_end', key: 'end', title: '到期时间' },
  {
    dataIndex: 'remaining_credits',
    key: 'credits',
    title: '剩余额度',
    width: 110,
  },
];

async function load() {
  loading.value = true;
  try {
    items.value = await listSubscriptionsApi({
      user_id: userId.value || undefined,
    });
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <Page description="查看用户当前生效的套餐、有效期和剩余额度" title="订阅管理">
    <Card>
      <div class="mb-4 flex gap-3">
        <Input
          v-model:value="userId"
          allow-clear
          placeholder="按用户 ID 查询"
          style="width: 280px"
          @press-enter="load"
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
          <template v-if="column.key === 'user'">
            <div class="font-medium">{{ record.user?.name || '-' }}</div>
            <div class="text-xs text-gray-500">
              {{ record.user?.email || record.user_id }}
            </div>
          </template>
          <template v-else-if="column.key === 'plan'">
            <div>{{ record.plan?.name || '-' }}</div>
            <div class="text-xs text-gray-500">
              {{ record.plan?.plan_code || record.plan_id }}
            </div>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusColors[record.subscription_status] || 'default'">
              {{
                statusLabels[record.subscription_status] ||
                record.subscription_status
              }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'start'">
            {{ formatDate(record.current_period_start) }}
          </template>
          <template v-else-if="column.key === 'end'">
            {{ formatDate(record.current_period_end) }}
          </template>
        </template>
      </Table>
    </Card>
  </Page>
</template>
