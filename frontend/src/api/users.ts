import { api, callApi, callPaginatedApi } from './config';
import type { UserResponse, CreateUserRequest, ChangePasswordRequest, DeleteUserRequest, OrphanedDirectory, PaginatedResponse, ApiResponse } from '../types';

/**
 * User Management API (Admin only)
 */
export const userAPI = {
  /**
   * Get users with pagination and search
   */
  getUsers: async (page?: number, pageSize?: number, search?: string): Promise<PaginatedResponse<UserResponse[]>> => {
    const params = new URLSearchParams();
    if (page) params.append('page', page.toString());
    if (pageSize) params.append('page_size', pageSize.toString());
    if (search) params.append('search', search);
    const queryString = params.toString();
    return await callPaginatedApi(() => api.get<UserResponse[]>(`/admin/users${queryString ? '?' + queryString : ''}`));
  },

  /**
   * Search users by username (for autocomplete)
   */
  searchUsers: async (query: string): Promise<ApiResponse<UserResponse[]>> => {
    return await callApi(() => api.get<UserResponse[]>(`/admin/users/search?q=${encodeURIComponent(query)}`));
  },

  /**
   * Create a new user
   */
  createUser: async (data: CreateUserRequest): Promise<ApiResponse<{ username: string }>> => {
    return await callApi(() => api.post<{ username: string }>('/admin/users', data));
  },

  /**
   * Delete a user with optional home directory deletion
   */
  deleteUser: async (username: string, data: DeleteUserRequest): Promise<ApiResponse<void>> => {
    return await callApi(() => api.delete<void>(`/admin/users/${username}`, data));
  },

  /**
   * Change user password (admin changing other user's password)
   */
  changePassword: async (username: string, data: ChangePasswordRequest): Promise<ApiResponse<void>> => {
    return await callApi(() => api.put<void>(`/admin/users/${username}/password`, data));
  },

  /**
   * Get orphaned directories (directories without corresponding users)
   */
  getOrphanedDirectories: async (): Promise<ApiResponse<OrphanedDirectory[]>> => {
    return await callApi(() => api.get<OrphanedDirectory[]>('/admin/users/orphaned'));
  },

  /**
   * Delete an orphaned directory
   */
  deleteOrphanedDirectory: async (dirName: string): Promise<ApiResponse<void>> => {
    return await callApi(() => api.delete<void>(`/admin/users/orphaned/${dirName}`));
  },
};
