import { es } from '../i18n/es';
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
  return humanizeError(error, fallback);
}

/** Map backend/Postgres text to a Spanish sentence the operator can act on. */
export function humanizeError(error: unknown, fallback = es.errors.generic): string {
  if (!error) return fallback;

  const code = isApiError(error) ? error.code : '';
  const raw = isApiError(error) || error instanceof Error ? error.message : '';
  const lower = raw.toLowerCase();

  if (code === 'dependency_cycle' || lower.includes('dependency cycle')) {
    const path = extractCyclePath(raw);
    return path ? `${es.errors.dependencyCycle}: ${path}` : es.errors.dependencyCycle;
  }
  if (lower.includes('duplicate component')) return es.errors.duplicateComponent;
  if (lower.includes('cannot be a component of its own')) return es.errors.selfComponent;
  if (
    lower.includes('duplicate key') ||
    lower.includes('unique constraint') ||
    lower.includes('products_pkey') ||
    lower.includes('already exists')
  ) {
    return es.errors.duplicateProduct;
  }
  if (lower.includes('update or delete') || lower.includes('is still referenced')) {
    return es.errors.productInUse;
  }
  if (lower.includes('invalid unit') || lower.includes('unit mismatch')) {
    return es.errors.unitMismatch;
  }
  if (
    lower.includes('component_product_id') ||
    lower.includes('at least one component') ||
    lower.includes('foreign key') ||
    lower.includes('is not present')
  ) {
    return es.errors.missingComponent;
  }
  if (code === 'validation_error' && raw) return mapValidationMessage(raw);
  if (code === 'conflict') return es.errors.conflict;
  if (code === 'not_found') return es.errors.notFound;
  if (/violates|constraint|sqlstate|pq:/.test(lower)) return fallback;
  if (raw) return raw;
  return fallback;
}

function mapValidationMessage(raw: string): string {
  const lower = raw.toLowerCase();
  if (lower.includes('duplicate')) return es.errors.duplicateComponent;
  if (lower.includes('own recipe')) return es.errors.selfComponent;
  if (lower.includes('component')) return es.errors.missingComponent;
  if (lower.includes('unit')) return es.errors.unitMismatch;
  return raw.replace(/^validation error:\s*/i, '');
}

function extractCyclePath(message: string): string {
  const match = message.match(/([A-Z0-9]+(?:\s*->\s*[A-Z0-9]+)+)/i);
  return match?.[1]?.replace(/\s+/g, ' ') ?? '';
}

/** Normalize any fetch error into the typed API shape used by production. */
export function resolveApiError(error: unknown): ApiError {
  const normalized = toDisplayError(error);
  if (isApiError(normalized)) return normalized;
  return { status: 0, code: 'unknown_error', message: normalized.message };
}
