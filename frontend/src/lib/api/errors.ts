import { isApiError, type ApiError } from '../../types/api';

interface ErrorEnvelope {
  error?: { code?: string; message?: string };
}

function isErrorEnvelope(value: unknown): value is ErrorEnvelope {
  if (typeof value !== 'object' || value === null || !('error' in value)) {
    return false;
  }

  const body = (value as ErrorEnvelope).error;
  return typeof body === 'object' && body !== null &&
    typeof body.code === 'string' && typeof body.message === 'string';
}

function parseEnvelope(message: string): { code: string; message: string } | null {
  try {
    const parsed: unknown = JSON.parse(message);
    if (isErrorEnvelope(parsed) && parsed.error) {
      return {
        code: parsed.error.code ?? 'unknown_error',
        message: parsed.error.message ?? '',
      };
    }
  } catch {
    return null;
  }

  return null;
}

/** Normalize fetch/ApiError into the typed envelope used by every feature. */
export function toDisplayError(error: unknown): ApiError | Error {
  if (isApiError(error)) {
    const nested = parseEnvelope(error.message);
    if (nested) {
      return { status: error.status, code: nested.code, message: nested.message };
    }
    return error;
  }

  if (error instanceof Error) {
    const nested = parseEnvelope(error.message);
    if (nested) {
      return { status: 0, code: nested.code, message: nested.message };
    }
    return error;
  }

  return new Error(String(error));
}

export function errorMessage(error: ApiError | Error | null, fallback: string): string {
  if (!error) return fallback;
  return error.message || fallback;
}

/** Normalize any fetch error into the typed API shape used by production. */
export function resolveApiError(error: unknown): ApiError {
  const normalized = toDisplayError(error);
  if (isApiError(normalized)) return normalized;
  return { status: 0, code: 'unknown_error', message: normalized.message };
}
