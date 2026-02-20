import type { ApiResult } from './types';
import { authStore } from '../stores/auth.svelte';

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '');
}

function ensureLeadingSlash(value: string): string {
  return value.startsWith('/') ? value : `/${value}`;
}

export function joinApiUrl(baseUrl: string, path: string): string {
  return `${trimTrailingSlash(baseUrl)}${ensureLeadingSlash(path)}`;
}

function looksLikeJson(text: string): boolean {
  const trimmed = text.trim();
  return trimmed.startsWith('{') || trimmed.startsWith('[');
}

async function parseResponseBody(res: Response): Promise<unknown> {
  const raw = await res.text();
  if (!raw) return {};

  const contentType = res.headers.get('content-type') ?? '';
  const shouldParseJson =
    contentType.includes('application/json') || looksLikeJson(raw);

  if (shouldParseJson) {
    try {
      return JSON.parse(raw);
    } catch {
      return {
        error: 'invalid_response_payload',
        details: 'Expected JSON response but received malformed payload',
        raw,
      };
    }
  }

  if (res.ok) {
    return { value: raw };
  }

  return {
    error: 'http_error',
    details: raw,
  };
}

export async function api<T>(
  path: string,
  opts: {
    method?: string;
    body?: unknown;
    idempotencyKey?: string;
  } = {},
): Promise<ApiResult<T>> {
  const { method = 'GET', body, idempotencyKey } = opts;
  const headers: Record<string, string> = {};

  if (body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }

  if (authStore.token) {
    headers['Authorization'] = `Bearer ${authStore.token}`;
  }

  if (idempotencyKey) {
    headers['Idempotency-Key'] = idempotencyKey;
  }

  const url = joinApiUrl(authStore.baseUrl, path);

  try {
    const res = await fetch(url, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });

    const data = (await parseResponseBody(res)) as T;
    return { ok: res.ok, status: res.status, data };
  } catch (err) {
    return {
      ok: false,
      status: 0,
      data: {
        error: 'network_error',
        details: err instanceof Error ? err.message : String(err),
        url,
      } as T,
    };
  }
}

export function idemKey(): string {
  return crypto.randomUUID();
}
