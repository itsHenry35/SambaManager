import { api } from './config';
import type { ApiResponse, ShareResponse, CreateMyShareRequest, UpdateShareRequest } from '../types';

/**
 * User Share API (for current logged-in user's shares)
 */
export const userShareAPI = {
  /**
   * Get current user's shares
   */
  getMyShares: async (): Promise<ApiResponse<ShareResponse[]>> => {
    return api.get<ShareResponse[]>('/user/shares');
  },

  /**
   * Create a new share for current user (owner is automatically set to current user)
   */
  createMyShare: async (data: CreateMyShareRequest): Promise<ApiResponse<string>> => {
    return api.post<string>('/user/shares', data);
  },

  /**
   * Update current user's share (identified by shareId in URL)
   */
  updateMyShare: async (shareId: string, data: UpdateShareRequest): Promise<ApiResponse<void>> => {
    return api.put<void>(`/user/shares/${shareId}`, data);
  },

  /**
   * Delete current user's share (identified by shareId in URL)
   */
  deleteMyShare: async (shareId: string): Promise<ApiResponse<void>> => {
    return api.delete<void>(`/user/shares/${shareId}`);
  },
};
