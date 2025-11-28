import type { LoginRequest, LoginResponse, ApiResponse } from '../types';

/**
 * Authentication API
 */
export const authAPI = {
  /**
   * Login with username and password
   * Note: Login endpoint doesn't require authentication
   */
  login: async (credentials: LoginRequest): Promise<ApiResponse<LoginResponse>> => {
    // Use api.post directly without auth since it's the login endpoint
    const headers = new Headers();
    headers.set('Content-Type', 'application/json');

    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        headers,
        body: JSON.stringify(credentials),
      });

      const data: ApiResponse<LoginResponse> = await response.json();
      return data;
    } catch (error) {
      return {
        code: 0,
        message: error instanceof Error ? error.message : 'Login failed',
        data: undefined,
      };
    }
  },
};
