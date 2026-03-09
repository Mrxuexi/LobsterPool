import axios from 'axios';
import type {
  AdminInstanceSummary,
  AdminOverview,
  AdminUserSummary,
  AuthResponse,
  ClawTemplate,
  ClusterTarget,
  CreateInstanceRequest,
  CreateTemplateRequest,
  Instance,
  UpdateUserMaxInstancesRequest,
  User,
} from '../types';

const TOKEN_STORAGE_KEY = 'token';
const USER_STORAGE_KEY = 'user';
const LOGIN_PATH = '/login';

interface HealthResponse {
  status: string;
}

interface MessageResponse {
  message: string;
}

function clearStoredAuth(): void {
  localStorage.removeItem(TOKEN_STORAGE_KEY);
  localStorage.removeItem(USER_STORAGE_KEY);
}

function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_STORAGE_KEY);
}

function isAuthRequest(url: unknown): boolean {
  return typeof url === 'string' && url.includes('/auth/');
}

const api = axios.create({
  baseURL: '/api/v1',
});

api.interceptors.request.use((config) => {
  const token = getStoredToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 && !isAuthRequest(error.config?.url)) {
      clearStoredAuth();
      window.location.href = LOGIN_PATH;
    }
    return Promise.reject(error);
  },
);

export async function register(username: string, password: string): Promise<AuthResponse> {
  const { data } = await api.post<AuthResponse>('/auth/register', { username, password });
  return data;
}

export async function login(username: string, password: string): Promise<AuthResponse> {
  const { data } = await api.post<AuthResponse>('/auth/login', { username, password });
  return data;
}

export async function getMe(): Promise<User> {
  const { data } = await api.get<User>('/auth/me');
  return data;
}

export async function changePassword(newPassword: string): Promise<User> {
  const { data } = await api.post<User>('/auth/change-password', { new_password: newPassword });
  return data;
}

export async function getHealth(): Promise<HealthResponse> {
  const { data } = await api.get<HealthResponse>('/health');
  return data;
}

export async function listTemplates(): Promise<ClawTemplate[]> {
  const { data } = await api.get<ClawTemplate[]>('/templates');
  return data;
}

export async function getTemplate(id: string): Promise<ClawTemplate> {
  const { data } = await api.get<ClawTemplate>(`/templates/${id}`);
  return data;
}

export async function listClusters(): Promise<ClusterTarget[]> {
  const { data } = await api.get<ClusterTarget[]>('/clusters');
  return data;
}

export async function getAdminOverview(): Promise<AdminOverview> {
  const { data } = await api.get<AdminOverview>('/admin/overview');
  return data;
}

export async function listAdminUsers(): Promise<AdminUserSummary[]> {
  const { data } = await api.get<AdminUserSummary[]>('/admin/users');
  return data;
}

export async function listAdminInstances(): Promise<AdminInstanceSummary[]> {
  const { data } = await api.get<AdminInstanceSummary[]>('/admin/instances');
  return data;
}

export async function updateUserMaxInstances(userID: string, req: UpdateUserMaxInstancesRequest): Promise<User> {
  const { data } = await api.patch<User>(`/admin/users/${userID}/max-instances`, req);
  return data;
}

export async function createTemplate(req: CreateTemplateRequest): Promise<ClawTemplate> {
  const { data } = await api.post<ClawTemplate>('/admin/templates', req);
  return data;
}

export async function listInstances(): Promise<Instance[]> {
  const { data } = await api.get<Instance[]>('/instances');
  return data;
}

export async function createInstance(req: CreateInstanceRequest): Promise<Instance> {
  const { data } = await api.post<Instance>('/instances', req);
  return data;
}

export async function getInstance(id: string): Promise<Instance> {
  const { data } = await api.get<Instance>(`/instances/${id}`);
  return data;
}

export async function deleteInstance(id: string): Promise<MessageResponse> {
  const { data } = await api.delete<MessageResponse>(`/instances/${id}`);
  return data;
}
