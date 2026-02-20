const STORAGE_KEY = 'vda_auth';

interface AuthState {
  token: string;
  baseUrl: string;
}

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '');
}

function inferBaseUrl(): string {
  // In production, API is same origin
  if (location.pathname.startsWith('/ui')) {
    return location.origin;
  }
  // Dev mode: Vite dev server proxies or direct
  return 'http://localhost:4020';
}

function normalizeBaseUrl(raw: string): string {
  const trimmed = raw.trim();
  if (!trimmed) return inferBaseUrl();

  if (
    !trimmed.startsWith('/') &&
    !/^[a-zA-Z][a-zA-Z\d+\-.]*:\/\//.test(trimmed)
  ) {
    return trimTrailingSlash(`http://${trimmed}`);
  }

  return trimTrailingSlash(trimmed);
}

function loadFromStorage(): AuthState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw) as Partial<AuthState>;
      return {
        token: typeof parsed.token === 'string' ? parsed.token : '',
        baseUrl: typeof parsed.baseUrl === 'string' ? parsed.baseUrl : '',
      };
    }
  } catch {
    // ignore
  }
  return { token: '', baseUrl: '' };
}

function createAuthStore() {
  const stored = loadFromStorage();
  let token = $state(stored.token.trim());
  let baseUrl = $state(normalizeBaseUrl(stored.baseUrl));

  function persist() {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ token, baseUrl }));
  }

  return {
    get token() { return token; },
    set token(v: string) { token = v.trim(); persist(); },
    get baseUrl() { return baseUrl; },
    set baseUrl(v: string) { baseUrl = normalizeBaseUrl(v); persist(); },
    get isAuthenticated() { return token.length > 0; },
    clear() {
      token = '';
      baseUrl = normalizeBaseUrl('');
      localStorage.removeItem(STORAGE_KEY);
    },
  };
}

export const authStore = createAuthStore();
