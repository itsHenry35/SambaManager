import { api, callApi, callPaginatedApi } from './config';
import type { CreateShareRequest, UpdateShareRequest, ShareResponse, PaginatedResponse, ApiResponse } from '../types';

/**
 * Share Management API (Admin only)
 */
export const shareAPI = {
  /**
   * Get shares with pagination and search
   */
  getShares: async (page?: number, pageSize?: number, search?: string): Promise<PaginatedResponse<ShareResponse[]>> => {
    const params = new URLSearchParams();
    if (page) params.append('page', page.toString());
    if (pageSize) params.append('page_size', pageSize.toString());
    if (search) params.append('search', search);
    const queryString = params.toString();
    return await callPaginatedApi(() => api.get<ShareResponse[]>(`/admin/shares${queryString ? '?' + queryString : ''}`));
  },

  /**
   * Create a new share
   */
  createShare: async (data: CreateShareRequest): Promise<ApiResponse<{ id: string }>> => {
    return await callApi(() => api.post<{ id: string }>('/admin/shares', data));
  },

  /**
   * Update an existing share
   */
  updateShare: async (shareId: string, data: UpdateShareRequest): Promise<ApiResponse<{ id: string }>> => {
    return await callApi(() => api.put<{ id: string }>(`/admin/shares/${shareId}`, data));
  },

  /**
   * Delete a share
   */
  deleteShare: async (shareId: string): Promise<ApiResponse<void>> => {
    return await callApi(() => api.delete<void>(`/admin/shares/${shareId}`));
  },
};
