/**
 * Typed API client with timeout, abort, error handling, and 204 support.
 */

import type { ApiError } from '../../types/api';

// ============================================================================
// Configuration
// ============================================================================

/** Validate and retrieve API base URL from environment */
function getApiBaseUrl(): string {
  const url = import.meta.env.PUBLIC_API_URL;
  
  if (!url) {
    throw new Error(
      'PUBLIC_API_URL is not defined. Check .env or environment variables.'
    );
  }

  // Basic URL validation
  try {
    new URL(url);
  } catch {
    throw new Error(
      `PUBLIC_API_URL is invalid: "${url}". Must be a valid HTTP/HTTPS URL.`
    );
  }

  return url;
}

const API_BASE_URL = getApiBaseUrl();

/** Default request timeout in milliseconds */
const DEFAULT_TIMEOUT_MS = 15_000;

// ============================================================================
// Core request function
// ============================================================================

export interface RequestOptions {
  /** HTTP method */
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  /** Request body (will be JSON-stringified) */
  body?: unknown;
  /** Request timeout in milliseconds (default: 15000) */
  timeoutMs?: number;
  /** Optional AbortSignal for external cancellation */
  signal?: AbortSignal;
  /** Additional headers */
  headers?: Record<string, string>;
}

/**
 * Make a typed API request with timeout, abort, and error handling.
 * 
 * @param endpoint - API endpoint path (e.g., '/health' or '/api/products')
 * @param options - Request options
 * @returns Parsed response data, or null for 204 No Content
 * @throws {ApiError} Structured error from backend
 * @throws {Error} Network, timeout, or abort errors
 */
export async function apiRequest<T>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<T | null> {
  const {
    method = 'GET',
    body,
    timeoutMs = DEFAULT_TIMEOUT_MS,
    signal: externalSignal,
    headers = {},
  } = options;

  // Create timeout abort controller
  const timeoutController = new AbortController();
  const timeoutId = setTimeout(() => timeoutController.abort(), timeoutMs);

  // Combine external and timeout signals
  const combinedSignal = externalSignal
    ? combineAbortSignals([externalSignal, timeoutController.signal])
    : timeoutController.signal;

  try {
    const url = `${API_BASE_URL}${endpoint}`;
    
    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        ...headers,
      },
      body: body ? JSON.stringify(body) : undefined,
      signal: combinedSignal,
    });

    clearTimeout(timeoutId);

    // Handle 204 No Content
    if (response.status === 204) {
      return null;
    }

    // Parse response body
    let data: unknown;
    const contentType = response.headers.get('content-type');
    
    if (contentType?.includes('application/json')) {
      data = await response.json();
    } else {
      // Non-JSON response for non-2xx status
      const text = await response.text();
      if (!response.ok) {
        throw createApiError(response.status, 'invalid_response', text);
      }
      data = null;
    }

    if (!response.ok) {
      throw toApiError(response.status, data);
    }

    return data as T;

  } catch (error) {
    clearTimeout(timeoutId);

    // Re-throw ApiError as-is
    if (isApiErrorResponse(error)) {
      throw error;
    }

    // Handle abort
    if (error instanceof Error && error.name === 'AbortError') {
      if (externalSignal?.aborted) {
        throw new Error('Request was cancelled');
      }
      throw new Error(`Request timeout after ${timeoutMs}ms`);
    }

    // Handle network errors
    if (error instanceof TypeError) {
      throw new Error(`Network error: ${error.message}`);
    }

    // Re-throw other errors
    throw error;
  }
}

// ============================================================================
// Convenience methods
// ============================================================================

export async function get<T>(
  endpoint: string,
  options?: Omit<RequestOptions, 'method' | 'body'>
): Promise<T | null> {
  return apiRequest<T>(endpoint, { ...options, method: 'GET' });
}

export async function post<T>(
  endpoint: string,
  body?: unknown,
  options?: Omit<RequestOptions, 'method' | 'body'>
): Promise<T | null> {
  return apiRequest<T>(endpoint, { ...options, method: 'POST', body });
}

export async function put<T>(
  endpoint: string,
  body?: unknown,
  options?: Omit<RequestOptions, 'method' | 'body'>
): Promise<T | null> {
  return apiRequest<T>(endpoint, { ...options, method: 'PUT', body });
}

export async function patch<T>(
  endpoint: string,
  body?: unknown,
  options?: Omit<RequestOptions, 'method' | 'body'>
): Promise<T | null> {
  return apiRequest<T>(endpoint, { ...options, method: 'PATCH', body });
}

export async function del<T>(
  endpoint: string,
  options?: Omit<RequestOptions, 'method' | 'body'>
): Promise<T | null> {
  return apiRequest<T>(endpoint, { ...options, method: 'DELETE' });
}

// ============================================================================
// Utilities
// ============================================================================

/** Type guard for API error responses */
function isApiErrorResponse(error: unknown): error is ApiError {
  return (
    typeof error === 'object' &&
    error !== null &&
    'status' in error &&
    'code' in error &&
    'message' in error
  );
}

/** Backend writes `{ error: { code, message } }`; normalize to ApiError. */
function toApiError(status: number, data: unknown): ApiError {
  if (isApiErrorResponse(data)) {
    return { status, code: data.code, message: data.message };
  }
  if (typeof data === 'object' && data !== null && 'error' in data) {
    const body = (data as { error: unknown }).error;
    if (
      typeof body === 'object' &&
      body !== null &&
      'code' in body &&
      'message' in body &&
      typeof (body as { code: unknown }).code === 'string' &&
      typeof (body as { message: unknown }).message === 'string'
    ) {
      return {
        status,
        code: (body as { code: string }).code,
        message: (body as { message: string }).message,
      };
    }
  }
  return createApiError(
    status,
    'unknown_error',
    typeof data === 'object' ? JSON.stringify(data) : String(data)
  );
}

/** Create a structured API error */
function createApiError(
  status: number,
  code: string,
  message: string
): ApiError {
  return { status, code, message };
}

/**
 * Combine multiple AbortSignals into one.
 * The combined signal aborts when any of the input signals aborts.
 */
function combineAbortSignals(signals: AbortSignal[]): AbortSignal {
  const controller = new AbortController();

  for (const signal of signals) {
    if (signal.aborted) {
      controller.abort();
      break;
    }
    signal.addEventListener('abort', () => controller.abort(), { once: true });
  }

  return controller.signal;
}
