// Standard API response structure
export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

// Paginated response structure
export interface PaginatedResponse<T = unknown> extends ApiResponse<T> {
  total: number;
  page: number;
  size: number;
}

// Pagination info for callbacks
export interface PaginationInfo {
  total: number;
  page: number;
  size: number;
}

// Options for response handling
export interface HandleRespOptions {
  auth?: boolean;           // Check for 401 and redirect to login (default: true)
  notifyError?: boolean;    // Show error message notification (default: true)
  notifySuccess?: boolean;  // Show success message notification (default: false)
}

// Global handlers (set by App component)
let globalLogout: (() => void) | null = null;
let globalNavigate: ((path: string) => void) | null = null;
let globalShowSnackbar: ((message: string, severity: 'success' | 'error' | 'warning' | 'info') => void) | null = null;

export const setGlobalHandlers = (
  logout: () => void,
  navigate: (path: string) => void,
  showSnackbar: (message: string, severity: 'success' | 'error' | 'warning' | 'info') => void
) => {
  globalLogout = logout;
  globalNavigate = navigate;
  globalShowSnackbar = showSnackbar;
};

/**
 * Core response handler with flexible options
 *
 * @param resp - API response
 * @param success - Success callback with data and optional pagination
 * @param fail - Failure callback with error message and code
 * @param options - Handling options
 */
const handleRespCore = <T>(
  resp: ApiResponse<T> | PaginatedResponse<T> | ApiResponse<unknown>,
  success?: (data: T, pagination?: PaginationInfo) => void,
  fail?: (message: string, code: number) => void,
  options: HandleRespOptions = {}
) => {
  const {
    auth = true,
    notifyError = true,
    notifySuccess = false
  } = options;

  // Handle success (code 200 or 201)
  if (resp.code === 200 || resp.code === 201) {
    if (notifySuccess && resp.message && globalShowSnackbar) {
      globalShowSnackbar(resp.message, 'success');
    }

    if (success) {
      // Check if it's a paginated response
      const paginatedResp = resp as PaginatedResponse<T>;
      if (paginatedResp.total !== undefined) {
        success(resp.data as T, {
          total: paginatedResp.total,
          page: paginatedResp.page,
          size: paginatedResp.size
        });
      } else {
        success(resp.data as T);
      }
    }
    return;
  }

  // Handle error cases - show error notification first
  if (notifyError && resp.message && globalShowSnackbar) {
    globalShowSnackbar(resp.message, 'error');
  }

  // Then handle 401 Unauthorized after showing error
  if (auth && resp.code === 401) {
    if (globalLogout) globalLogout();
    if (globalNavigate) globalNavigate('/login');
    return;
  }

  // Finally call fail callback
  if (fail) {
    fail(resp.message || 'Unknown error', resp.code);
  }
};

/**
 * Standard response handler
 * - Checks auth (401 redirects to login)
 * - Shows error notifications
 * - No success notifications
 */
export const handleResp = <T>(
  resp: ApiResponse<T> | PaginatedResponse<T> | ApiResponse<unknown>,
  success?: (data: T, pagination?: PaginationInfo) => void,
  fail?: (message: string, code: number) => void
) => {
  handleRespCore(resp, success, fail, {
    auth: true,
    notifyError: true,
    notifySuccess: false
  });
};

/**
 * Response handler without any notifications
 * - Checks auth (401 redirects to login)
 * - No error notifications
 * - No success notifications
 */
export const handleRespWithoutNotify = <T>(
  resp: ApiResponse<T> | PaginatedResponse<T> | ApiResponse<unknown>,
  success?: (data: T, pagination?: PaginationInfo) => void,
  fail?: (message: string, code: number) => void
) => {
  handleRespCore(resp, success, fail, {
    auth: true,
    notifyError: false,
    notifySuccess: false
  });
};

/**
 * Response handler for public APIs (no auth check)
 * - No auth check
 * - No error notifications
 * - No success notifications
 */
export const handleRespWithoutAuthAndNotify = <T>(
  resp: ApiResponse<T> | PaginatedResponse<T> | ApiResponse<unknown>,
  success?: (data: T, pagination?: PaginationInfo) => void,
  fail?: (message: string, code: number) => void
) => {
  handleRespCore(resp, success, fail, {
    auth: false,
    notifyError: false,
    notifySuccess: false
  });
};

/**
 * Response handler with success notification
 * Used for CUD operations (Create, Update, Delete)
 * - Checks auth (401 redirects to login)
 * - Shows error notifications
 * - Shows success notifications
 */
export const handleRespWithNotifySuccess = <T>(
  resp: ApiResponse<T> | PaginatedResponse<T> | ApiResponse<unknown>,
  success?: (data: T, pagination?: PaginationInfo) => void,
  fail?: (message: string, code: number) => void
) => {
  handleRespCore(resp, success, fail, {
    auth: true,
    notifyError: true,
    notifySuccess: true
  });
};

/**
 * Response handler for public APIs with success notification
 * - No auth check
 * - Shows error notifications
 * - Shows success notifications
 */
export const handleRespWithoutAuthButNotifySuccess = <T>(
  resp: ApiResponse<T> | PaginatedResponse<T> | ApiResponse<unknown>,
  success?: (data: T, pagination?: PaginationInfo) => void,
  fail?: (message: string, code: number) => void
) => {
  handleRespCore(resp, success, fail, {
    auth: false,
    notifyError: true,
    notifySuccess: true
  });
};

// Batch operation types
export interface BatchRequestItem {
  name: string;
  request: () => Promise<ApiResponse<unknown>>;
}

export interface BatchResult {
  name: string;
  success: boolean;
  error?: string;
}

export interface BatchOptions {
  batchSize?: number;
  onProgress?: (completed: number, total: number) => void;
  onSuccess?: (results: BatchResult[]) => void;
  onComplete?: (results: BatchResult[]) => void;
}

/**
 * Handle batch operations with progress tracking
 *
 * @param items - Array of batch request items
 * @param options - Batch processing options
 * @returns Array of batch results
 */
export const handleBatchResp = async (
  items: BatchRequestItem[],
  options: BatchOptions = {}
): Promise<BatchResult[]> => {
  const {
    batchSize = 5,
    onProgress,
    onSuccess,
    onComplete
  } = options;

  const results: BatchResult[] = [];
  let completed = 0;

  // Process in batches
  for (let i = 0; i < items.length; i += batchSize) {
    const batch = items.slice(i, i + batchSize);

    const batchResults = await Promise.all(
      batch.map(async (item) => {
        try {
          const resp = await item.request();

          if (resp.code === 200 || resp.code === 201) {
            return {
              name: item.name,
              success: true
            };
          } else {
            return {
              name: item.name,
              success: false,
              error: resp.message || 'Unknown error'
            };
          }
        } catch (error) {
          return {
            name: item.name,
            success: false,
            error: error instanceof Error ? error.message : 'Unknown error'
          };
        }
      })
    );

    results.push(...batchResults);
    completed += batch.length;

    if (onProgress) {
      onProgress(completed, items.length);
    }
  }

  // Notify based on results
  const successCount = results.filter(r => r.success).length;
  const failCount = results.filter(r => !r.success).length;

  if (globalShowSnackbar) {
    if (failCount === 0) {
      globalShowSnackbar(`All ${successCount} operations completed successfully`, 'success');
      if (onSuccess) onSuccess(results);
    } else if (successCount === 0) {
      globalShowSnackbar(`All ${failCount} operations failed`, 'error');
    } else {
      globalShowSnackbar(`${successCount} succeeded, ${failCount} failed`, 'warning');
    }
  }

  if (onComplete) {
    onComplete(results);
  }

  return results;
};
