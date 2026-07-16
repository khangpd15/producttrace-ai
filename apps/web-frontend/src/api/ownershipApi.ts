import { goFetch } from '../lib/api';
import { getAccessToken } from '../lib/auth';

export type RequestOtpData = {
  product_id: string;
};

function authToken(): string {
  const token = getAccessToken();
  if (!token) {
    throw new Error('You must be logged in to perform this action.');
  }
  return token;
}

export async function customerRequestOtp(qrCode: string) {
  return goFetch<RequestOtpData>(
    '/ownership/request-otp',
    {
      method: 'POST',
      body: JSON.stringify({ qr_code: qrCode }),
    },
    authToken(),
  );
}

export async function customerVerifyAndRegister(otp: string, productId: string) {
  return goFetch(
    '/ownership/register',
    {
      method: 'POST',
      body: JSON.stringify({ otp, product_id: productId }),
    },
    authToken(),
  );
}

export async function adminRequestOtp(payload: {
  qr_code: string;
  owner_name: string;
  owner_email: string;
  owner_phone?: string;
}) {
  return goFetch<RequestOtpData>(
    '/ownership/admin/request-otp',
    {
      method: 'POST',
      body: JSON.stringify(payload),
    },
    authToken(),
  );
}

export async function adminVerifyAndRegister(payload: {
  otp: string;
  product_id: string;
  owner_name: string;
  owner_email: string;
  owner_phone?: string;
}) {
  return goFetch(
    '/ownership/admin/register',
    {
      method: 'POST',
      body: JSON.stringify(payload),
    },
    authToken(),
  );
}

export async function transferOwnership(
  ownershipId: string,
  payload: {
    new_owner_name: string;
    new_owner_email: string;
    new_owner_phone?: string;
    new_owner_address?: string;
  },
) {
  return goFetch(
    `/ownership/${ownershipId}/transfer`,
    {
      method: 'PUT',
      body: JSON.stringify(payload),
    },
    authToken(),
  );
}
