import { api, callApi } from './config';
import type { ApiResponse, ChangeOwnPasswordRequest, UserResponse } from '../types';

/**
 * User Profile API (for current logged-in user)
 */
export const userProfileAPI = {
  /**
   * Change own password
   */
  changeOwnPassword: async (data: ChangeOwnPasswordRequest): Promise<ApiResponse<void>> => {
    return api.put<void>('/user/password', data);
  },

  /**
   * Search users by username (for autocomplete when creating shares)
   */
  searchUsers: async (query: string) => {
    return await callApi(() => api.get<UserResponse[]>(`/user/users/search?q=${encodeURIComponent(query)}`));
  },
};
