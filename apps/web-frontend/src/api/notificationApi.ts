import { nestFetch } from '../lib/api';
import { getAccessToken } from '../lib/auth';

export type NotificationStatus = 'sent' | 'failed' | 'queued' | 'pending';

export type NotificationHistoryItem = {
  id: string;
  type: string;
  recipient: string;
  status: NotificationStatus;
  sent_at: string;
  product_name?: string | null;
  full_name?: string | null;
  subject?: string | null;
};

export type NotificationHistoryQuery = {
  page?: number;
  limit?: number;
  type?: string;
  status?: NotificationStatus;
};

export type NotificationHistoryResponse = {
  items: NotificationHistoryItem[];
  total: number;
  page: number;
  limit: number;
};

function authToken(): string {
  const token = getAccessToken();
  if (!token) {
    throw new Error('You must be logged in to view notification history.');
  }
  return token;
}

function pickString(obj: Record<string, unknown>, keys: string[]): string | undefined {
  for (const key of keys) {
    const value = obj[key];
    if (typeof value === 'string' && value) return value;
  }
  return undefined;
}

function normalizeStatus(value: unknown): NotificationStatus {
  const raw = String(value ?? 'sent').toLowerCase();
  if (raw === 'failed' || raw === 'error') return 'failed';
  if (raw === 'queued') return 'queued';
  if (raw === 'pending') return 'pending';
  return 'sent';
}

function normalizeItem(raw: unknown): NotificationHistoryItem | null {
  if (!raw || typeof raw !== 'object') return null;
  const obj = raw as Record<string, unknown>;

  const id = pickString(obj, ['id', 'notification_id', 'notificationId']);
  const type = pickString(obj, ['type', 'event_type', 'eventType', 'notification_type']);
  const recipient = pickString(obj, ['recipient', 'email', 'to']);
  const sentAt = pickString(obj, ['sent_at', 'sentAt', 'created_at', 'createdAt']);

  if (!id || !type || !recipient || !sentAt) return null;

  return {
    id,
    type,
    recipient,
    status: normalizeStatus(obj.status),
    sent_at: sentAt,
    product_name: pickString(obj, ['product_name', 'productName']) ?? null,
    full_name: pickString(obj, ['full_name', 'fullName']) ?? null,
    subject: pickString(obj, ['subject', 'title']) ?? null,
  };
}

function normalizeResponse(payload: unknown): NotificationHistoryResponse {
  const root = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {};
  const data = root.data && typeof root.data === 'object' ? (root.data as Record<string, unknown>) : root;

  const rawItems = Array.isArray(data.items)
    ? data.items
    : Array.isArray(data.notifications)
      ? data.notifications
      : Array.isArray(data)
        ? data
        : [];

  const items = rawItems
    .map(normalizeItem)
    .filter((item): item is NotificationHistoryItem => item !== null);

  const total = Number(data.total ?? data.totalRecords ?? items.length);
  const page = Number(data.page ?? data.currentPage ?? 1);
  const limit = Number(data.limit ?? data.pageSize ?? (items.length || 20));

  return { items, total, page, limit };
}

export async function getNotificationHistory(
  query: NotificationHistoryQuery = {},
): Promise<NotificationHistoryResponse> {
  const params = new URLSearchParams();
  if (query.page) params.set('page', String(query.page));
  if (query.limit) params.set('limit', String(query.limit));
  if (query.type) params.set('type', query.type);
  if (query.status) params.set('status', query.status);

  const qs = params.toString();
  const path = qs ? `/notifications/history?${qs}` : '/notifications/history';

  const payload = await nestFetch<unknown>(path, { method: 'GET', cache: 'no-store' }, authToken());
  return normalizeResponse(payload);
}
