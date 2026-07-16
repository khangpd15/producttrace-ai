import { nestFetch } from '../lib/api';

export type ForgotPasswordResponse = {
  message: string;
};

export type ValidateResetTokenResponse = {
  valid: boolean;
  message: string;
};

export type ResetPasswordResponse = {
  message: string;
};

export async function forgotPassword(email: string) {
  return nestFetch<ForgotPasswordResponse>('/auth/forgot-password', {
    method: 'POST',
    body: JSON.stringify({ email }),
  });
}

export async function validateResetToken(token: string, email: string) {
  const params = new URLSearchParams({ token, email });
  return nestFetch<ValidateResetTokenResponse>(
    `/auth/validate-reset-token?${params.toString()}`,
    { method: 'GET', cache: 'no-store' },
  );
}

export async function resetPassword(email: string, otpCode: string, newPassword: string) {
  return nestFetch<ResetPasswordResponse>('/auth/reset-password', {
    method: 'POST',
    body: JSON.stringify({
      email,
      otp_code: otpCode,
      new_password: newPassword,
    }),
  });
}
