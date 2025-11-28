import type { ApiResponse, PaginatedResponse } from '../types';

const API_BASE_URL = '/api';

/**
 * Enhanced fetch wrapper with interceptors and error handling
 *
 * @param url - API endpoint URL
 * @param options - Fetch options
 * @param requiresAuth - Whether this request requires authentication
 * @returns Promise with API response
 */
const fetchWithInterceptors = async <T>(
  url: string,
  options: RequestInit = {},
  requiresAuth: boolean = true
): Promise<ApiResponse<T>> => {
  // Request interceptor - add auth token
  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');

  if (requiresAuth) {
    const token = localStorage.getItem('token');
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    }
  }

  const config: RequestInit = {
    ...options,
    headers,
  };

  try {
    const response = await fetch(url, config);

    // Try to parse JSON response
    try {
      const data: ApiResponse<T> = await response.json();

      // Return the parsed response (even if HTTP status is error)
      return data;
    } catch {
      // Response is not JSON - construct error response
      return {
        code: response.status,
        message: `Server error: ${response.statusText}`,
        data: undefined,
      };
    }
  } catch (error) {
    // Network error or request failed
    if (error instanceof TypeError) {
      return {
        code: 0,
        message: 'Network connection failed. Please check your network.',
        data: undefined,
      };
    }

    return {
      code: 0,
      message: error instanceof Error ? error.message : 'Unknown error occurred',
      data: undefined,
    };
  }
};

/**
 * API client for authenticated requests
 */
export const api = {
  get: async <T>(url: string): Promise<ApiResponse<T>> => {
    return fetchWithInterceptors<T>(`${API_BASE_URL}${url}`, {
      method: 'GET',
    });
  },

  post: async <T>(url: string, data?: unknown): Promise<ApiResponse<T>> => {
    return fetchWithInterceptors<T>(`${API_BASE_URL}${url}`, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  },

  put: async <T>(url: string, data?: unknown): Promise<ApiResponse<T>> => {
    return fetchWithInterceptors<T>(`${API_BASE_URL}${url}`, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  },

  delete: async <T>(url: string, data?: unknown): Promise<ApiResponse<T>> => {
    return fetchWithInterceptors<T>(`${API_BASE_URL}${url}`, {
      method: 'DELETE',
      body: data ? JSON.stringify(data) : undefined,
    });
  },
};

/**
 * Helper function to make API calls
 * Wraps the API call for consistent handling
 */
export const callApi = async <T>(
  apiCall: () => Promise<ApiResponse<T>>
): Promise<ApiResponse<T>> => {
  return await apiCall();
};

/**
 * Helper function to make paginated API calls
 * Wraps the API call for consistent handling of paginated responses
 */
export const callPaginatedApi = async <T>(
  apiCall: () => Promise<ApiResponse<T>>
): Promise<PaginatedResponse<T>> => {
  const response = await apiCall();
  
  // If the response already has pagination properties, cast it directly
  if ('total' in response && 'page' in response && 'size' in response) {
    return response as PaginatedResponse<T>;
  }
  
  // Otherwise, assume it's a simple ApiResponse and create pagination metadata
  // This typically happens when the backend returns pagination info in headers or metadata
  return {
    ...response,
    total: Array.isArray(response.data) ? response.data.length : 0,
    page: 1,
    size: Array.isArray(response.data) ? response.data.length : 0
  } as PaginatedResponse<T>;
};
