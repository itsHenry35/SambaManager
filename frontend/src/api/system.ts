import { api, callApi } from './config';
import type { SystemCheckResponse, SambaConfigResponse, UpdateSambaConfigRequest, SambaConfigFileResponse, UpdateSambaConfigFileRequest, SambaStatusResponse, ApiResponse } from '../types';

/**
 * System Management API (Admin only)
 */
export const systemAPI = {
  /**
   * Check system environment
   */
  checkEnvironment: async () => {
    return await callApi(() => api.get<SystemCheckResponse>('/admin/system/check'));
  },

  /**
   * Get Samba configuration
   */
  getSambaConfig: async () => {
    return await callApi(() => api.get<SambaConfigResponse>('/admin/system/config'));
  },

  /**
   * Update Samba configuration
   */
  updateSambaConfig: async (data: UpdateSambaConfigRequest) => {
    return await callApi(() => api.put<void>('/admin/system/config', data));
  },

  /**
   * Get raw smb.conf file content
   */
  getSambaConfigFile: async () => {
    return await callApi(() => api.get<SambaConfigFileResponse>('/admin/system/config/file'));
  },

  /**
   * Update raw smb.conf file content
   */
  updateSambaConfigFile: async (data: UpdateSambaConfigFileRequest): Promise<ApiResponse<void>> => {
    return await callApi(() => api.put<void>('/admin/system/config/file', data));
  },

  /**
   * Get Samba status (smbstatus output)
   */
  getSambaStatus: async (): Promise<ApiResponse<SambaStatusResponse>> => {
    return await callApi(() => api.get<SambaStatusResponse>('/admin/system/status'));
  },
};
