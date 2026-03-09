export interface ClawTemplate {
  id: string;
  name: string;
  description: string;
  image: string;
  version: string;
  default_port: number;
  created_at: string;
}

export interface Instance {
  id: string;
  name: string;
  template_id: string;
  user_id: string;
  cluster: string;
  namespace: string;
  deployment_name: string;
  service_name: string;
  status: string;
  endpoint: string;
  created_at: string;
}

export interface CreateInstanceRequest {
  name: string;
  template_id: string;
  cluster?: string;
  api_key: string;
  mm_bot_token: string;
}

export interface ClusterTarget {
  name: string;
  display_name: string;
  namespace: string;
  default: boolean;
}

export interface User {
  id: string;
  username: string;
  role: 'admin' | 'member';
  max_instances: number;
  must_change_password: boolean;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface AdminUserSummary {
  id: string;
  username: string;
  role: 'admin' | 'member';
  max_instances: number;
  instance_count: number;
  created_at: string;
}

export interface AdminInstanceSummary {
  id: string;
  name: string;
  template_id: string;
  user_id: string;
  username: string;
  cluster: string;
  namespace: string;
  deployment_name: string;
  service_name: string;
  status: string;
  endpoint: string;
  created_at: string;
}

export interface AdminOverview {
  total_users: number;
  admin_users: number;
  total_instances: number;
  running_instances: number;
  total_templates: number;
  recent_users: AdminUserSummary[];
  recent_instances: AdminInstanceSummary[];
}

export interface CreateTemplateRequest {
  id: string;
  name: string;
  description: string;
  image: string;
  version?: string;
  default_port?: number;
}

export interface UpdateUserMaxInstancesRequest {
  max_instances: number;
}
