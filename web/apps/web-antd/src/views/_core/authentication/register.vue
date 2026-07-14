<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';
import type { Recordable } from '@vben/types';

import { computed } from 'vue';

import { AuthenticationRegister, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { useAuthStore } from '#/store';

defineOptions({ name: 'Register' });

const authStore = useAuthStore();

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: '请输入管理员名称',
      },
      fieldName: 'name',
      label: '管理员名称',
      rules: z.string().min(1, { message: '请输入管理员名称' }),
    },
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: 'admin@example.com',
      },
      fieldName: 'email',
      label: '管理员邮箱',
      rules: z
        .string()
        .min(1, { message: '请输入管理员邮箱' })
        .email('请输入有效的邮箱地址'),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        passwordStrength: true,
        placeholder: $t('authentication.password'),
      },
      fieldName: 'password',
      label: $t('authentication.password'),
      renderComponentContent() {
        return {
          strengthText: () => $t('authentication.passwordStrength'),
        };
      },
      rules: z.string().min(8, { message: '密码至少 8 位' }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: $t('authentication.confirmPassword'),
      },
      dependencies: {
        rules(values) {
          const { password } = values;
          return z
            .string({ required_error: $t('authentication.passwordTip') })
            .min(8, { message: '密码至少 8 位' })
            .refine((value) => value === password, {
              message: $t('authentication.confirmPasswordTip'),
            });
        },
        triggerFields: ['password'],
      },
      fieldName: 'confirmPassword',
      label: $t('authentication.confirmPassword'),
    },
  ];
});

async function handleSubmit(value: Recordable<any>) {
  await authStore.registerFirstAdmin(value);
}
</script>

<template>
  <AuthenticationRegister
    :form-schema="formSchema"
    :loading="authStore.loginLoading"
    submit-button-text="创建超级管理员"
    sub-title="这是首次部署初始化。该账号拥有系统全部管理权限。"
    title="注册首位用户"
    @submit="handleSubmit"
  />
</template>
