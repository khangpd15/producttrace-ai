import React, { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  getNotificationHistory,
  type NotificationHistoryItem,
  type NotificationStatus,
} from '../api/notificationApi';
import { useAuth } from '../context/AuthContext';
import { ApiError } from '../lib/api';

const EVENT_TYPE_LABELS: Record<string, string> = {
  'notification.sent': 'Cập nhật bảo hành',
  'warranty.expired': 'Hết hạn bảo hành',
  'ownership.transferred': 'Chuyển quyền sở hữu',
  'otp.registered': 'OTP đăng ký',
  'otp.verified': 'Xác thực email',
  'otp.password_reset': 'Đặt lại mật khẩu',
  'ownership.otp': 'OTP sở hữu sản phẩm',
};

const STATUS_LABELS: Record<NotificationStatus, string> = {
  sent: 'Đã gửi',
  failed: 'Thất bại',
  queued: 'Đang chờ',
  pending: 'Chờ xử lý',
};

const STATUS_STYLES: Record<NotificationStatus, string> = {
  sent: 'bg-green-100 text-green-800',
  failed: 'bg-red-100 text-red-800',
  queued: 'bg-yellow-100 text-yellow-800',
  pending: 'bg-gray-100 text-gray-800',
};

function formatEventType(type: string): string {
  return EVENT_TYPE_LABELS[type] ?? type;
}

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString('vi-VN');
}

export default function NotificationHistoryPage() {
  const { email, logout } = useAuth();
  const navigate = useNavigate();

  const [items, setItems] = useState<NotificationHistoryItem[]>([]);
  const [page, setPage] = useState(1);
  const [limit] = useState(20);
  const [total, setTotal] = useState(0);
  const [typeFilter, setTypeFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState<NotificationStatus | ''>('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  const totalPages = Math.max(1, Math.ceil(total / limit));

  const loadHistory = useCallback(async () => {
    setIsLoading(true);
    setError('');
    try {
      const result = await getNotificationHistory({
        page,
        limit,
        type: typeFilter || undefined,
        status: statusFilter || undefined,
      });
      setItems(result.items);
      setTotal(result.total);
    } catch (err: unknown) {
      const message =
        err instanceof ApiError || err instanceof Error
          ? err.message
          : 'Không thể tải lịch sử thông báo';
      setError(message);
      setItems([]);
      setTotal(0);
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, typeFilter, statusFilter]);

  useEffect(() => {
    loadHistory();
  }, [loadHistory]);

  const handleFilterSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setPage(1);
    loadHistory();
  };

  return (
    <div className="min-h-screen bg-gray-50 py-10 px-4 sm:px-6 lg:px-8">
      <div className="max-w-5xl mx-auto">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Lịch sử thông báo</h1>
            <p className="text-sm text-gray-600">Đăng nhập: {email || 'user'}</p>
          </div>
          <div className="flex gap-3 text-sm">
            <Link to="/ownership" className="text-indigo-600 hover:text-indigo-500">
              Ownership
            </Link>
            <button
              type="button"
              onClick={() => {
                logout();
                navigate('/login');
              }}
              className="text-gray-600 hover:text-gray-800"
            >
              Logout
            </button>
          </div>
        </div>

        <div className="bg-white shadow rounded-lg p-6 space-y-6">
          <form className="grid grid-cols-1 sm:grid-cols-4 gap-3" onSubmit={handleFilterSubmit}>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Loại thông báo</label>
              <select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
                className="block w-full px-3 py-2 border border-gray-300 rounded-md sm:text-sm"
              >
                <option value="">Tất cả</option>
                <option value="notification.sent">Cập nhật bảo hành</option>
                <option value="warranty.expired">Hết hạn bảo hành</option>
                <option value="ownership.transferred">Chuyển quyền sở hữu</option>
                <option value="otp.registered">OTP đăng ký</option>
                <option value="otp.verified">Xác thực email</option>
                <option value="otp.password_reset">Đặt lại mật khẩu</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Trạng thái</label>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value as NotificationStatus | '')}
                className="block w-full px-3 py-2 border border-gray-300 rounded-md sm:text-sm"
              >
                <option value="">Tất cả</option>
                <option value="sent">Đã gửi</option>
                <option value="failed">Thất bại</option>
                <option value="queued">Đang chờ</option>
                <option value="pending">Chờ xử lý</option>
              </select>
            </div>
            <div className="sm:col-span-2 flex items-end gap-2">
              <button
                type="submit"
                disabled={isLoading}
                className="px-4 py-2 rounded-md text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60"
              >
                Lọc
              </button>
              <button
                type="button"
                disabled={isLoading}
                onClick={() => {
                  setTypeFilter('');
                  setStatusFilter('');
                  setPage(1);
                }}
                className="px-4 py-2 rounded-md text-sm font-medium border border-gray-300 text-gray-700"
              >
                Xóa lọc
              </button>
            </div>
          </form>

          {error && <div className="text-sm text-red-600 bg-red-50 p-3 rounded">{error}</div>}

          {isLoading ? (
            <div className="text-sm text-gray-600 py-8 text-center">Đang tải lịch sử thông báo...</div>
          ) : items.length === 0 ? (
            <div className="text-sm text-gray-600 py-8 text-center">Chưa có thông báo nào.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 text-sm">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-3 py-2 text-left font-medium text-gray-700">Thời gian</th>
                    <th className="px-3 py-2 text-left font-medium text-gray-700">Loại</th>
                    <th className="px-3 py-2 text-left font-medium text-gray-700">Người nhận</th>
                    <th className="px-3 py-2 text-left font-medium text-gray-700">Sản phẩm</th>
                    <th className="px-3 py-2 text-left font-medium text-gray-700">Trạng thái</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {items.map((item) => (
                    <tr key={item.id} className="hover:bg-gray-50">
                      <td className="px-3 py-3 whitespace-nowrap text-gray-700">
                        {formatDate(item.sent_at)}
                      </td>
                      <td className="px-3 py-3 text-gray-900">
                        <div className="font-medium">{formatEventType(item.type)}</div>
                        {item.subject && (
                          <div className="text-xs text-gray-500 mt-0.5">{item.subject}</div>
                        )}
                      </td>
                      <td className="px-3 py-3 text-gray-700">
                        <div>{item.recipient}</div>
                        {item.full_name && (
                          <div className="text-xs text-gray-500 mt-0.5">{item.full_name}</div>
                        )}
                      </td>
                      <td className="px-3 py-3 text-gray-700">{item.product_name || '—'}</td>
                      <td className="px-3 py-3">
                        <span
                          className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_STYLES[item.status]}`}
                        >
                          {STATUS_LABELS[item.status]}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {!isLoading && total > 0 && (
            <div className="flex items-center justify-between pt-2">
              <p className="text-sm text-gray-600">
                Tổng {total} thông báo — Trang {page}/{totalPages}
              </p>
              <div className="flex gap-2">
                <button
                  type="button"
                  disabled={page <= 1 || isLoading}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  className="px-3 py-1.5 rounded-md text-sm border border-gray-300 disabled:opacity-50"
                >
                  Trước
                </button>
                <button
                  type="button"
                  disabled={page >= totalPages || isLoading}
                  onClick={() => setPage((p) => p + 1)}
                  className="px-3 py-1.5 rounded-md text-sm border border-gray-300 disabled:opacity-50"
                >
                  Sau
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
