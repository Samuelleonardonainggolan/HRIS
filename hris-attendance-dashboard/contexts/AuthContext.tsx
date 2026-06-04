'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { authService, User, LoginRequest } from '@/lib/api/auth';
import { useRouter, usePathname } from 'next/navigation';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    // Check if user is logged in on mount
    const initAuth = () => {
      const currentUser = authService.getUser();
      const token = authService.getAccessToken();

      if (currentUser && token) {
        setUser(currentUser);
      } else {
        setUser(null);
        // Clear invalid/expired tokens from localStorage and cookies
        authService.clearTokens();
        
        // Only redirect to login if not already on login page
        if (pathname && !pathname.startsWith('/login')) {
          router.push('/login');
        }
      }
      setLoading(false);
    };

    initAuth();
  }, [pathname, router]);

  const login = async (credentials: LoginRequest) => {
    try {
      console.log('AuthContext: Attempting login...');
      
      const response = await authService.login(credentials);
      
      console.log('AuthContext: Login successful', {
        user: response.user.full_name,
        role: response.user.role,
      });

      setUser(response.user);

      // Redirect based on role
      const role = response.user.role;
      let redirectPath = '/dashboard';

      switch (role) {
        case 'manager_hr':
          redirectPath = '/dashboard/manager-hr';
          break;
        case 'manager_departemen':
          redirectPath = '/dashboard/manager-dept';
          break;
        case 'admin_departemen':
          redirectPath = '/dashboard/admin-dept';
          break;
        case 'staf':
          redirectPath = '/dashboard/staff';
          break;
        case 'accountant':
          redirectPath = '/dashboard/accountant';
          break;
        default:
          redirectPath = '/dashboard';
      }

      console.log('AuthContext: Redirecting to', redirectPath);
      router.push(redirectPath);
    } catch (error) {
      console.error('AuthContext: Login failed', error);
      setUser(null);
      throw error;
    }
  };

  const logout = async () => {
    try {
      await authService.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      setUser(null);
      router.push('/login');
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        login,
        logout,
        isAuthenticated: !!user,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};