const STORAGE_KEY = 'vda_auth';

interface AuthState {
  token: string;
  baseUrl: string;
}

function loadFromStorage(): AuthState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw);
  } catch {
    // ignore
  }
  return { token: '', baseUrl: '' };
}

function createAuthStore() {
  const stored = loadFromStorage();
  let token = $state(stored.token);
  let baseUrl = $state(stored.baseUrl || inferBaseUrl());

  function inferBaseUrl(): string {
    // In production, API is same origin
    if (location.pathname.startsWith('/ui')) {
      return location.origin;
    }
    // Dev mode: Vite dev server proxies or direct
    return 'http://localhost:4020';
  }

  function persist() {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ token, baseUrl }));
  }

  return {
    get token() { return token; },
    set token(v: string) { token = v; persist(); },
    get baseUrl() { return baseUrl; },
    set baseUrl(v: string) { baseUrl = v; persist(); },
    get isAuthenticated() { return token.length > 0; },
    clear() {
      token = '';
      baseUrl = inferBaseUrl();
      localStorage.removeItem(STORAGE_KEY);
    },
  };
}

export const authStore = createAuthStore();
