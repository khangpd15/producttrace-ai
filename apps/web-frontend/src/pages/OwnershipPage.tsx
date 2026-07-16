import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  adminRequestOtp,
  adminVerifyAndRegister,
  customerRequestOtp,
  customerVerifyAndRegister,
  transferOwnership,
} from '../api/ownershipApi';
import { useAuth } from '../context/AuthContext';
import { ApiError } from '../lib/api';

type Mode = 'customer' | 'admin';

function extractOwnershipId(data: unknown): string | null {
  if (!data || typeof data !== 'object') return null;
  const obj = data as Record<string, unknown>;
  const candidates = [obj.ownership_id, obj.id, obj.ownershipId];
  for (const value of candidates) {
    if (typeof value === 'string' && value) return value;
  }
  return null;
}

export default function OwnershipPage() {
  const { email, logout } = useAuth();
  const navigate = useNavigate();

  const [mode, setMode] = useState<Mode>('customer');
  const [step, setStep] = useState<'request' | 'verify'>('request');

  const [qrCode, setQrCode] = useState('');
  const [ownerName, setOwnerName] = useState('');
  const [ownerEmail, setOwnerEmail] = useState('');
  const [ownerPhone, setOwnerPhone] = useState('');
  const [productId, setProductId] = useState('');
  const [otp, setOtp] = useState('');

  const [ownershipId, setOwnershipId] = useState('');
  const [transferName, setTransferName] = useState('');
  const [transferEmail, setTransferEmail] = useState('');
  const [transferPhone, setTransferPhone] = useState('');

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const resetMessages = () => {
    setError('');
    setSuccess('');
  };

  const handleRequestOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    resetMessages();
    setIsLoading(true);

    try {
      const result =
        mode === 'customer'
          ? await customerRequestOtp(qrCode)
          : await adminRequestOtp({
              qr_code: qrCode,
              owner_name: ownerName,
              owner_email: ownerEmail,
              owner_phone: ownerPhone || undefined,
            });

      const nextProductId = result.data?.product_id;
      if (!nextProductId) {
        throw new Error('OTP requested but product_id was missing in response.');
      }

      setProductId(nextProductId);
      setStep('verify');
      setSuccess(result.message || 'OTP sent to email. Please check inbox.');
    } catch (err: unknown) {
      const message =
        err instanceof ApiError || err instanceof Error ? err.message : 'Failed to request OTP';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    resetMessages();
    setIsLoading(true);

    try {
      const result =
        mode === 'customer'
          ? await customerVerifyAndRegister(otp, productId)
          : await adminVerifyAndRegister({
              otp,
              product_id: productId,
              owner_name: ownerName,
              owner_email: ownerEmail,
              owner_phone: ownerPhone || undefined,
            });

      const id = extractOwnershipId(result.data);
      if (id) setOwnershipId(id);

      setSuccess(result.message || 'Ownership registered successfully.');
    } catch (err: unknown) {
      const message =
        err instanceof ApiError || err instanceof Error ? err.message : 'Failed to verify OTP';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleTransfer = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!ownershipId) {
      setError('Ownership ID is required to transfer.');
      return;
    }

    resetMessages();
    setIsLoading(true);
    try {
      const result = await transferOwnership(ownershipId, {
        new_owner_name: transferName,
        new_owner_email: transferEmail,
        new_owner_phone: transferPhone || undefined,
      });
      setSuccess(result.message || 'Ownership transfer requested. Email notification will be sent.');
    } catch (err: unknown) {
      const message =
        err instanceof ApiError || err instanceof Error ? err.message : 'Failed to transfer ownership';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-10 px-4 sm:px-6 lg:px-8">
      <div className="max-w-2xl mx-auto">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Ownership OTP</h1>
            <p className="text-sm text-gray-600">Signed in as {email || 'user'}</p>
          </div>
          <div className="flex gap-3 text-sm">
            <Link to="/notifications" className="text-indigo-600 hover:text-indigo-500">
              Thông báo
            </Link>
            <Link to="/login" className="text-indigo-600 hover:text-indigo-500">
              Home
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
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => {
                setMode('customer');
                setStep('request');
                resetMessages();
              }}
              className={`px-3 py-1.5 rounded text-sm ${
                mode === 'customer' ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-700'
              }`}
            >
              Customer
            </button>
            <button
              type="button"
              onClick={() => {
                setMode('admin');
                setStep('request');
                resetMessages();
              }}
              className={`px-3 py-1.5 rounded text-sm ${
                mode === 'admin' ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-700'
              }`}
            >
              Admin
            </button>
          </div>

          {error && <div className="text-sm text-red-600 bg-red-50 p-3 rounded">{error}</div>}
          {success && <div className="text-sm text-green-700 bg-green-50 p-3 rounded">{success}</div>}

          {step === 'request' ? (
            <form className="space-y-4" onSubmit={handleRequestOtp}>
              <div>
                <label className="block text-sm font-medium text-gray-700">QR code</label>
                <input
                  required
                  value={qrCode}
                  onChange={(e) => setQrCode(e.target.value)}
                  className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md sm:text-sm"
                  placeholder="Scan or paste QR payload"
                />
              </div>

              {mode === 'admin' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Owner name</label>
                    <input
                      required
                      value={ownerName}
                      onChange={(e) => setOwnerName(e.target.value)}
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md sm:text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Owner email</label>
                    <input
                      type="email"
                      required
                      value={ownerEmail}
                      onChange={(e) => setOwnerEmail(e.target.value)}
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md sm:text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Owner phone (optional)</label>
                    <input
                      value={ownerPhone}
                      onChange={(e) => setOwnerPhone(e.target.value)}
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md sm:text-sm"
                    />
                  </div>
                </>
              )}

              <button
                type="submit"
                disabled={isLoading}
                className="w-full py-2 px-4 rounded-md text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60"
              >
                {isLoading ? 'Sending OTP...' : 'Request OTP email'}
              </button>
            </form>
          ) : (
            <form className="space-y-4" onSubmit={handleVerify}>
              <div className="text-sm text-gray-600">
                Product ID: <span className="font-mono text-gray-900">{productId}</span>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">OTP from email</label>
                <input
                  required
                  value={otp}
                  onChange={(e) => setOtp(e.target.value.trim())}
                  className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md tracking-widest sm:text-sm"
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setStep('request');
                    setOtp('');
                    resetMessages();
                  }}
                  className="flex-1 py-2 px-4 rounded-md text-sm font-medium border border-gray-300 text-gray-700"
                >
                  Back
                </button>
                <button
                  type="submit"
                  disabled={isLoading}
                  className="flex-1 py-2 px-4 rounded-md text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60"
                >
                  {isLoading ? 'Verifying...' : 'Verify & register'}
                </button>
              </div>
            </form>
          )}
        </div>

        <div className="mt-6 bg-white shadow rounded-lg p-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">Transfer ownership (email notify)</h2>
          <p className="text-sm text-gray-600">
            After a successful register, paste ownership ID if not auto-filled, then transfer to send email.
          </p>
          <form className="space-y-4" onSubmit={handleTransfer}>
            <div>
              <label className="block text-sm font-medium text-gray-700">Ownership ID</label>
              <input
                required
                value={ownershipId}
                onChange={(e) => setOwnershipId(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md sm:text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">New owner name</label>
              <input
                required
                value={transferName}
                onChange={(e) => setTransferName(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md sm:text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">New owner email</label>
              <input
                type="email"
                required
                value={transferEmail}
                onChange={(e) => setTransferEmail(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md sm:text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">New owner phone (optional)</label>
              <input
                value={transferPhone}
                onChange={(e) => setTransferPhone(e.target.value)}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md sm:text-sm"
              />
            </div>
            <button
              type="submit"
              disabled={isLoading}
              className="w-full py-2 px-4 rounded-md text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60"
            >
              {isLoading ? 'Transferring...' : 'Transfer ownership'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
