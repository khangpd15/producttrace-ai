export class ApiError extends Error {
  status: number;
  details?: unknown;

  constructor(message: string, status: number, details?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.details = details;
  }
}

export type GoApiResponse<T = unknown> = {
  success: boolean;
  message: string;
  data?: T;
  error?: unknown;
};

function getGoBaseUrl(): string {
  return (import.meta.env.VITE_GO_CORE_API_URL || 'http://localhost:8080/api').replace(/\/$/, '');
}

function getNestBaseUrl(): string {
  return (import.meta.env.VITE_NEST_AI_API_URL || 'http://localhost:3000').replace(/\/$/, '');
}

async function parseJsonSafe(response: Response): Promise<any> {
  const text = await response.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    return { message: text };
  }
}

export async function goFetch<T = unknown>(
  path: string,
  options: RequestInit = {},
  token?: string | null,
): Promise<GoApiResponse<T>> {
  const headers = new Headers(options.headers || {});
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json');
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const response = await fetch(`${getGoBaseUrl()}${path.startsWith('/') ? path : `/${path}`}`, {
    ...options,
    headers,
  });

  const payload = (await parseJsonSafe(response)) as GoApiResponse<T> | null;

  if (!response.ok || payload?.success === false) {
    const message =
      payload?.message ||
      (typeof payload?.error === 'string' ? payload.error : null) ||
      `Request failed (${response.status})`;
    throw new ApiError(message, response.status, payload?.error ?? payload);
  }

  return payload ?? { success: true, message: 'OK' };
}

export async function nestFetch<T = unknown>(
  path: string,
  options: RequestInit = {},
  token?: string | null,
): Promise<T> {
  const headers = new Headers(options.headers || {});
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json');
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const response = await fetch(`${getNestBaseUrl()}${path.startsWith('/') ? path : `/${path}`}`, {
    ...options,
    headers,
  });

  const payload = await parseJsonSafe(response);

  if (!response.ok) {
    const message =
      payload?.message ||
      (Array.isArray(payload?.message) ? payload.message.join(', ') : null) ||
      `Request failed (${response.status})`;
    throw new ApiError(message, response.status, payload);
  }

  return payload as T;
}
