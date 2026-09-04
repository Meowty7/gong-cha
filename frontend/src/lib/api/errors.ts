import { isApiError, type ApiError } from '../../types/api';

interface ErrorEnvelope {
  error?: { code?: string; message?: string };
}

function parseEnvelope(message: string): { code: string; message: string } | null {
  try {
    const data = JSON.parse(message) as ErrorEnvelope;
    const nested = data.error;
    if (nested && typeof nested.message === 'string' && nested.message) {
      return { code: nested.code ?? 'unknown_error', message: nested.message };
    }
  } catch {
    return null;
  }
  return null;
}

/** Normalize fetch/ApiError into the typed envelope, unwrapping backend `{ error: { code, message } }`. */
export function toDisplayError(error: unknown): ApiError | Error {
  if (isApiError(error)) {
    const nested = parseEnvelope(error.message);
    if (nested) {
      return { status: error.status, code: nested.code, message: nested.message };
    }
    return error;
  }
  if (error instanceof Error) return error;
  return new Error(String(error));
}

export function errorMessage(error: ApiError | Error | null, fallback: string): string {
  if (!error) return '';
  if (isApiError(error)) return error.message || fallback;
  return error.message || fallback;
}
