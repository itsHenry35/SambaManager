// ===== User Types =====

export interface User {
  username: string;
  password: string;
}

export interface UserResponse {
  username: string;
  home_dir: string;
}

export interface CreateUserRequest {
  username: string;
  password: string;
}

export interface ChangePasswordRequest {
  password: string;
}

export interface ChangeOwnPasswordRequest {
  old_password: string;
  new_password: string;
}

export type UserRole = 'admin' | 'user';

export interface OrphanedDirectory {
  name: string;
  path: string;
  size: number;
}

export interface DeleteUserRequest {
  delete_home_dir: boolean;
}

// ===== Share Types =====

export interface Share {
  owner: string;
  shared_with: string[];
  read_only: boolean;
  comment: string;
  sub_path?: string;
}

export interface ShareResponse {
  id: string;
  owner: string;
  path: string;
  shared_with: string[];
  read_only: boolean;
  comment: string;
  sub_path?: string;
}

export interface CreateShareRequest {
  name?: string;
  owner: string;
  shared_with: string[];
  read_only: boolean;
  comment: string;
  sub_path?: string;
}

export interface UpdateShareRequest {
  shared_with: string[];
  read_only: boolean;
  comment: string;
  sub_path?: string;
}

export interface CreateMyShareRequest {
  name?: string;
  shared_with: string[];
  read_only: boolean;
  comment: string;
  sub_path?: string;
}

// ===== Auth Types =====

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  username: string;
  role: UserRole;
  expires_at: number;
}

// ===== API Response Types =====

export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

export interface PaginatedResponse<T = unknown> extends ApiResponse<T> {
  total: number;
  page: number;
  size: number;
}

// ===== System Types =====

export interface CheckResult {
  id: string;
  status: 'pass' | 'fail' | 'warning';
}

export interface SystemCheckResponse {
  checks: CheckResult[];
}

export interface SambaGlobalConfig {
  workgroup: string;
  server_string: string;
  security: string;
  passdb_backend: string;
  map_to_guest: string;
  access_based_share_enum: string;
}

export interface SambaHomesConfig {
  comment: string;
  browseable: string;
  writable: string;
  valid_users: string;
  force_user: string;
  force_group: string;
  create_mask: string;
  directory_mask: string;
}

export interface SambaConfigResponse {
  global: SambaGlobalConfig;
  homes: SambaHomesConfig;
}

export interface UpdateSambaConfigRequest {
  global?: SambaGlobalConfig;
  homes?: SambaHomesConfig;
}

export interface SambaConfigFileResponse {
  content: string;
  path: string;
}

export interface UpdateSambaConfigFileRequest {
  content: string;
}

export interface SambaStatusResponse {
  raw_output: string;
}
