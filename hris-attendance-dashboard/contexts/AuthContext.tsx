'use client';

import { createContext, useContext, useEffect, useMemo, useState, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { authService, User, LoginRequest } from '@/lib/api/auth';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  isAuthenticated: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

function getRedirectPathByRole(role?: string) {
  switch (role) {
    case 'manager_hr':
      return '/dashboard/manager-hr';
    case 'manager_departemen':
      return '/dashboard/manager-dept';
    case 'admin_departemen':
      return '/dashboard/admin-dept';
    case 'staf':
      return '/dashboard/staff';
    default:
      return '/dashboard';
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const router = useRouter();

  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = async () => {
    setLoading(true);
    try {
      // === Pola 1: token + user di localStorage
      const currentUser = authService.getUser?.() ?? null;
      const token = authService.getAccessToken?.() ?? null;

      if (currentUser && token) {
        setUser(currentUser);
        return;
      }

      // === Pola 2: cookie/httpOnly -> ambil profile dari server (kalau tersedia)
      // (Jika authService.getProfile tidak ada, bagian ini akan dilewati)
      if (typeof (authService as any).getProfile === 'function') {
        const me = await (authService as any).getProfile();
        if (me) {
          setUser(me);
          // optional: persist user agar refresh berikutnya cepat
          if (typeof (authService as any).setUser === 'function') {
            (authService as any).setUser(me);
          }
          return;
        }
      }

      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // ✅ hanya sekali, tidak tergantung pathname/router
    refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const login = async (credentials: LoginRequest) => {
    setLoading(true);
    try {
      const response = await authService.login(credentials);

      // response.user wajib ada sesuai kode Anda sebelumnya
      setUser(response.user);

      // ✅ kalau authService punya setter, persist biar tidak hilang setelah refresh
      if (typeof (authService as any).setUser === 'function') {
        (authService as any).setUser(response.user);
      }
      // jika response punya access_token dan authService punya setter:
      if ((response as any).access_token && typeof (authService as any).setAccessToken === 'function') {
        (authService as any).setAccessToken((response as any).access_token);
      }

      const redirectPath = getRedirectPathByRole(response.user?.role);
      router.replace(redirectPath);
    } catch (error) {
      setUser(null);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    setLoading(true);
    try {
      await authService.logout?.();

      // bersihkan storage jika ada methodnya
      if (typeof (authService as any).clear === 'function') {
        (authService as any).clear();
      }
    } finally {
      setUser(null);
      setLoading(false);
      router.replace('/login');
    }
  };

  const value = useMemo<AuthContextType>(
    () => ({
      user,
      loading,
      isAuthenticated: !!user,
      login,
      logout,
      refresh,
    }),
    [user, loading]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider');
  return ctx;
}