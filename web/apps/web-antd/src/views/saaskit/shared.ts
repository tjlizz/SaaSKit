export const statusColors: Record<string, string> = {
  active: 'green',
  cancelled: 'default',
  disabled: 'default',
  paid: 'green',
  pending: 'orange',
  refunded: 'blue',
};

export const statusLabels: Record<string, string> = {
  active: '正常',
  cancelled: '已取消',
  disabled: '已停用',
  paid: '已支付',
  pending: '待支付',
  refunded: '已退款',
};

export const cycleLabels: Record<string, string> = {
  free: '免费',
  lifetime: '终身',
  monthly: '月付',
  one_time: '一次性',
  yearly: '年付',
};

export const providerLabels: Record<string, string> = {
  alipay: '支付宝',
  free: '免费',
  wechat: '微信支付',
};

export function formatDate(value?: string) {
  if (!value) return '-';
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}

export function formatMoney(cents: number, currency = 'CNY') {
  return new Intl.NumberFormat('zh-CN', {
    currency,
    style: 'currency',
  }).format(cents / 100);
}
