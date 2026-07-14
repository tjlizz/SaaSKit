import { requestClient } from '#/api/request';

export interface User {
  id: string;
  email: string;
  phone?: string;
  name: string;
  avatar_url?: string;
  status: 'active' | 'disabled';
  email_verified: boolean;
  phone_verified: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

export interface PageResult<T> {
  items: T[];
  page: number;
  page_size: number;
  total: number;
}

export interface Plan {
  id: string;
  plan_code: string;
  name: string;
  description: string;
  billing_cycle: 'free' | 'lifetime' | 'monthly' | 'one_time' | 'yearly';
  price_cents: number;
  currency: string;
  auto_renew_supported: boolean;
  device_limit: number;
  credit_quota: number;
  seat_limit: number;
  recommended: boolean;
  enabled: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: string;
  order_no: string;
  user_id: string;
  user?: User;
  plan_id: string;
  provider: string;
  provider_trade_no?: string;
  status: 'cancelled' | 'paid' | 'pending' | 'refunded';
  amount_cents: number;
  currency: string;
  quantity: number;
  paid_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Subscription {
  id: string;
  user_id: string;
  user?: User;
  plan_id: string;
  order_id: string;
  subscription_status: string;
  current_period_start: string;
  current_period_end: string;
  auto_renew: boolean;
  remaining_credits: number;
  plan?: Plan;
  created_at: string;
  updated_at: string;
}

export interface ApiClient {
  id: string;
  name: string;
  client_key: string;
  enabled: boolean;
  rate_limit_per_min: number;
  created_at: string;
  updated_at: string;
}

export interface PaymentConfigStatus {
  provider: 'alipay' | 'wechat';
  enabled: boolean;
  sandbox: boolean;
  configured: boolean;
  updated_at?: string;
}

export interface UserQuery {
  page?: number;
  page_size?: number;
  q?: string;
  status?: string;
}

export type UserInput = Pick<User, 'email' | 'name'> & {
  password?: string;
  phone?: string;
  status?: User['status'];
};

export type PlanInput = Omit<Plan, 'created_at' | 'id' | 'updated_at'>;

export function listUsersApi(params: UserQuery = {}) {
  return requestClient.get<PageResult<User>>('/admin/users', { params });
}

export function createUserApi(data: UserInput) {
  return requestClient.post<User>('/admin/users', data);
}

export function updateUserApi(id: string, data: Partial<UserInput>) {
  return requestClient.put<User>(`/admin/users/${id}`, data);
}

export function setUserStatusApi(id: string, status: User['status']) {
  return requestClient.put<User>(`/admin/users/${id}/status`, { status });
}

export function resetUserPasswordApi(id: string, password: string) {
  return requestClient.put(`/admin/users/${id}/password`, { password });
}

export function deleteUserApi(id: string) {
  return requestClient.delete(`/admin/users/${id}`);
}

export function listPlansApi() {
  return requestClient.get<Plan[]>('/admin/plans');
}

export function createPlanApi(data: PlanInput) {
  return requestClient.post<Plan>('/admin/plans', data);
}

export function updatePlanApi(id: string, data: PlanInput) {
  return requestClient.put<Plan>(`/admin/plans/${id}`, data);
}

export function deletePlanApi(id: string) {
  return requestClient.delete(`/admin/plans/${id}`);
}

export function listOrdersApi(
  params: { status?: string; user_id?: string } = {},
) {
  return requestClient.get<Order[]>('/admin/orders', { params });
}

export function updateOrderStatusApi(
  id: string,
  status: 'cancelled' | 'refunded',
) {
  return requestClient.put<Order>(`/admin/orders/${id}/status`, { status });
}

export function listSubscriptionsApi(params: { user_id?: string } = {}) {
  return requestClient.get<Subscription[]>('/admin/subscriptions', { params });
}

export function listApiClientsApi() {
  return requestClient.get<ApiClient[]>('/admin/api-clients');
}

export function createApiClientApi(data: {
  enabled: boolean;
  name: string;
  rate_limit_per_min: number;
}) {
  return requestClient.post<{ client: ApiClient; client_secret: string }>(
    '/admin/api-clients',
    data,
  );
}

export function updateApiClientApi(
  id: string,
  data: { enabled?: boolean; name?: string; rate_limit_per_min?: number },
) {
  return requestClient.put<ApiClient>(`/admin/api-clients/${id}`, data);
}

export function deleteApiClientApi(id: string) {
  return requestClient.delete(`/admin/api-clients/${id}`);
}

export function rotateApiClientSecretApi(id: string) {
  return requestClient.post<{ client_secret: string }>(
    `/admin/api-clients/${id}/rotate-secret`,
  );
}

export function listPaymentConfigsApi() {
  return requestClient.get<PaymentConfigStatus[]>('/admin/payment-configs');
}

export function savePaymentConfigApi(
  provider: PaymentConfigStatus['provider'],
  data: { config: Record<string, string>; enabled: boolean; sandbox: boolean },
) {
  return requestClient.put<PaymentConfigStatus>(
    `/admin/payment-configs/${provider}`,
    data,
  );
}
