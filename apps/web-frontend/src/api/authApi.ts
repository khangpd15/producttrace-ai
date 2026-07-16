import { goFetch } from '../lib/api';

export type RegisterPayload = {
  email: string;
  phone: string;
  full_name: string;
  password: string;
};

export type LoginPayload = {
  email: string;
  password: string;
};

export type TokenData = {
  access_token: string;
  refresh_token: string;
};

export type RegisterData = {
  full_name: string;
  phone: string;
  email: string;
  status: string;
};

export async function register(payload: RegisterPayload) {
  return goFetch<RegisterData>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function login(payload: LoginPayload) {
  return goFetch<TokenData>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function verifyOtp(email: string, otp: string) {
  return goFetch('/auth/verify-otp', {
    method: 'POST',
    body: JSON.stringify({ email, otp }),
  });
}

export async function resendOtp(email: string) {
  return goFetch('/auth/resend-otp', {
    method: 'POST',
    body: JSON.stringify({ email }),
  });
}

export async function logout(refreshToken: string) {
  return goFetch('/auth/logout', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
}
