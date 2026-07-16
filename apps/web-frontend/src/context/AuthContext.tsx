import { createContext, useContext, useMemo, useState, type ReactNode } from 'react';
import {
  clearAuthSession,
  getAccessToken,
  getRefreshToken,
  getStoredEmail,
  setAuthSession,
} from '../lib/auth';

type AuthContextValue = {
  isAuthenticated: boolean;
  email: string | null;
  accessToken: string | null;
  login: (accessToken: string, refreshToken: string, email?: string) => void;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [accessToken, setAccessToken] = useState<string | null>(() => getAccessToken());
  const [, setRefreshToken] = useState<string | null>(() => getRefreshToken());
  const [email, setEmail] = useState<string | null>(() => getStoredEmail());

  const value = useMemo<AuthContextValue>(
    () => ({
      isAuthenticated: Boolean(accessToken),
      email,
      accessToken,
      login: (nextAccessToken, nextRefreshToken, nextEmail) => {
        setAuthSession({
          accessToken: nextAccessToken,
          refreshToken: nextRefreshToken,
          email: nextEmail,
        });
        setAccessToken(nextAccessToken);
        setRefreshToken(nextRefreshToken);
        setEmail(nextEmail || getStoredEmail());
      },
      logout: () => {
        clearAuthSession();
        setAccessToken(null);
        setRefreshToken(null);
        setEmail(null);
      },
    }),
    [accessToken, email],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return ctx;
}
