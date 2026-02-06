const API_BASE = window.location.origin;

async function signData(data: string, secretKey: string): Promise<string> {
  const encoder = new TextEncoder();
  const keyData = encoder.encode(secretKey);
  const key = await crypto.subtle.importKey(
    'raw',
    keyData,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  );
  const signature = await crypto.subtle.sign('HMAC', key, encoder.encode(data));
  return btoa(String.fromCharCode(...new Uint8Array(signature)))
    .replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

interface AuthState {
  currentUser: string | null;
  token: string | null;
  secretKey: string | null;
}

let authState: AuthState = {
  currentUser: null,
  token: null,
  secretKey: null,
};

export function getAuthState() {
  return authState;
}

export function setAuthState(state: Partial<AuthState>) {
  authState = { ...authState, ...state };
}

async function apiCall(endpoint: string, method = 'GET', data: unknown = null, requireSignature = false) {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  if (authState.token) {
    headers['Authorization'] = `Bearer ${authState.token}`;
  }

  const bodyStr = data ? JSON.stringify(data) : '';

  if (requireSignature && authState.secretKey && authState.currentUser) {
    const timestamp = Date.now().toString();
    const signature = await signData(bodyStr + timestamp, authState.secretKey);
    headers['X-Combinator-Signature'] = signature;
    headers['X-Combinator-User-ID'] = authState.currentUser;
    headers['X-Combinator-Timestamp'] = timestamp;
  }

  const options: RequestInit = { method, headers };
  if (data) {
    options.body = bodyStr;
  }

  const response = await fetch(API_BASE + endpoint, options);
  const result = await response.json();

  if (!response.ok) {
    throw new Error(result.error || 'Request failed');
  }

  return result;
}

export const authAPI = {
  sendCode: (email: string) => apiCall('/auth/send-code', 'POST', { email }),
  register: (email: string, code: string, password: string) => apiCall('/auth/register', 'POST', { email, code, password }),
  login: (email: string, password: string) => apiCall('/auth/login', 'POST', { email, password }),
};

export const rdbAPI = {
  list: () => apiCall('/api/rdb', 'GET'),
  create: (name: string) => apiCall('/api/rdb', 'POST', { name }, true),
  delete: (id: string) => apiCall(`/api/rdb/${id}`, 'DELETE', {}, true),
};

export const kvAPI = {
  list: () => apiCall('/api/kv', 'GET'),
  create: (kv_type: string, url: string) => apiCall('/api/kv', 'POST', { kv_type, url }, true),
  delete: (id: string) => apiCall(`/api/kv/${id}`, 'DELETE', {}, true),
};

export const workerAPI = {
  list: () => apiCall('/api/worker', 'GET'),
  get: (id: string, offset?: number) => apiCall(`/api/worker/${id}${offset ? `?offset=${offset}` : ''}`, 'GET'),
  create: (worker_name: string) => apiCall('/api/worker', 'POST', { worker_name }),
  delete: (id: string) => apiCall(`/api/worker/${id}`, 'DELETE'),
};

export const domainAPI = {
  list: () => apiCall('/api/domain', 'GET'),
  get: (id: string) => apiCall(`/api/domain/${id}`, 'GET'),
  create: (domain: string, target: string) => apiCall('/api/domain', 'POST', { domain, target }, true),
  delete: (id: string) => apiCall(`/api/domain/${id}`, 'DELETE', {}, true),
};
