import type { ApiResult } from './types';
import { authStore } from '../stores/auth.svelte';

export async function api<T>(
  path: string,
  opts: {
    method?: string;
    body?: unknown;
    idempotencyKey?: string;
  } = {},
): Promise<ApiResult<T>> {
  const { method = 'GET', body, idempotencyKey } = opts;
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  if (authStore.token) {
    headers['Authorization'] = `Bearer ${authStore.token}`;
  }

  if (idempotencyKey) {
    headers['Idempotency-Key'] = idempotencyKey;
  }

  const url = `${authStore.baseUrl}${path}`;

  try {
    const res = await fetch(url, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    const data = await res.json();
    return { ok: res.ok, status: res.status, data };
  } catch (err) {
    return {
      ok: false,
      status: 0,
      data: { error: 'network_error', details: String(err) } as T,
    };
  }
}

export function idemKey(): string {
  return crypto.randomUUID();
}
