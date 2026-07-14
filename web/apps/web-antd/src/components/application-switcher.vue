<script lang="ts" setup>
import type { Application } from '#/api';

import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Select } from 'ant-design-vue';

import { listApplicationsApi } from '#/api';
import {
  getCurrentApplicationId,
  setCurrentApplicationId,
} from '#/utils/application-context';

const route = useRoute();
const router = useRouter();
const applications = ref<Application[]>([]);
const selectedId = ref(getCurrentApplicationId());

const options = computed(() =>
  applications.value
    .filter((item) => item.status === 'active')
    .map((item) => ({ label: item.name, value: item.id })),
);

async function load() {
  applications.value = await listApplicationsApi();
  const active = applications.value.filter((item) => item.status === 'active');
  if (!active.some((item) => item.id === selectedId.value)) {
    selectedId.value = active[0]?.id ?? '';
    setCurrentApplicationId(selectedId.value);
  }
  if (active.length === 0 && route.name !== 'Applications') {
    await router.replace({ name: 'Applications' });
  }
}

function change(value: unknown) {
  if (typeof value !== 'string') return;
  const id = value;
  selectedId.value = id;
  setCurrentApplicationId(id);
  router.go(0);
}

onMounted(load);
</script>

<template>
  <div class="flex items-center px-2">
    <Select
      :options="options"
      :value="selectedId || undefined"
      class="w-44"
      placeholder="请选择应用"
      @change="change"
    />
  </div>
</template>
