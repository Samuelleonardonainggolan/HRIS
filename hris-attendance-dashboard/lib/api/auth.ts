// lib/api/auth.ts
const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith('/')
    ? process.env.NEXT_PUBLIC_API_URL
    : '/api/v1';

function resolveApiUrl(path: string) {
  if (API_BASE.startsWith('http')) return `${API_BASE}${path}`;
  const origin =
    typeof window !== 'undefined' ? window.location.origin : 'http://localhost:3000';
  return `${origin}${API_BASE}${path}`;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface User {
  id: string;
  nik: string; // This maps to payroll_number in backend
  payroll_number?: string;
  email: string;
  full_name: string;
  role: string;
  department: string; // department_name
  department_name?: string;
  department_id?: string;
  position: string; // position_name
  position_name?: string;
  position_id?: string;
  phone?: string;
  address?: string;
  avatar?: string;
  join_date: string;
  birth_date?: string;
  religion?: string;
  last_education?: string;
  year_enrolled?: string;
  employment_status?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// ✅ Backend actual response structure (without wrapper)
export interface LoginResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
  requires_face_registration?: boolean;
}

export interface ErrorResponse {
  success: false;
  message: string;
  error: string;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function isTokenExpired(token: string): boolean {
  try {
    const payloadBase64 = token.split('.')[1];
    if (!payloadBase64) return true;
    
    const base64 = payloadBase64.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    
    const payload = JSON.parse(jsonPayload);
    if (!payload.exp) return false;
    
    const currentTime = Math.floor(Date.now() / 1000);
    // Tambahkan buffer 5 detik
    return payload.exp < currentTime + 5;
  } catch (e) {
    return true;
  }
}

class AuthService {
  private setAccessTokenCookie(token: string): void {
    if (typeof window === 'undefined') return;
    const maxAgeSeconds = 60 * 60 * 24 * 7;
    document.cookie = `access_token=${encodeURIComponent(token)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`;
  }

  private clearAccessTokenCookie(): void {
    if (typeof window === 'undefined') return;
    document.cookie = 'access_token=; Path=/; Max-Age=0; SameSite=Lax';
  }

  async login(credentials: LoginRequest): Promise<LoginResponse> {
    try {
      const response = await fetch(resolveApiUrl('/auth/login'), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(credentials),
      });

      const raw = await response.text();
      let data: unknown;
      try {
        data = raw ? (JSON.parse(raw) as unknown) : {};
      } catch {
        throw new Error(`Invalid JSON response (status ${response.status})`);
      }

      if (!response.ok) {
        const errorData = data as { error?: string; message?: string };
        throw new Error(errorData.error || errorData.message || 'Login failed');
      }

      const payloadRaw = isRecord(data) && isRecord(data.data) ? data.data : data;
      if (!isRecord(payloadRaw)) {
        console.error('Invalid response structure:', data);
        throw new Error('Invalid response from server');
      }

      const accessToken = payloadRaw.access_token;
      const refreshToken = payloadRaw.refresh_token;
      const expiresIn = payloadRaw.expires_in;
      const user = payloadRaw.user;

      if (
        typeof accessToken !== 'string' ||
        typeof refreshToken !== 'string' ||
        typeof expiresIn !== 'number' ||
        !isRecord(user)
      ) {
        console.error('Invalid response structure:', data);
        throw new Error('Invalid response from server');
      }

      const payload = payloadRaw as unknown as LoginResponse;

      // Save tokens to localStorage
      if (typeof window !== 'undefined') {
        localStorage.setItem('access_token', payload.access_token);
        localStorage.setItem('refresh_token', payload.refresh_token);
        localStorage.setItem('user', JSON.stringify(payload.user));
        this.setAccessTokenCookie(payload.access_token);
      }

      return payload;
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  }

  async logout(): Promise<void> {
    const token = this.getAccessToken();

    if (token) {
      try {
        await fetch(resolveApiUrl('/logout'), {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
      } catch (error) {
        console.error('Logout error:', error);
        // Continue with local logout even if API call fails
      }
    }

    // Clear localStorage
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
      this.clearAccessTokenCookie();
    }
  }

  async refreshToken(): Promise<LoginResponse> {
    const refreshToken = this.getRefreshToken();

    if (!refreshToken) {
      throw new Error('No refresh token available');
    }

    try {
      const response = await fetch(resolveApiUrl('/auth/refresh'), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      const raw = await response.text();
      let data: unknown;
      try {
        data = raw ? (JSON.parse(raw) as unknown) : {};
      } catch {
        throw new Error(`Invalid JSON response (status ${response.status})`);
      }

      if (!response.ok) {
        const errorData = data as { error?: string; message?: string };
        throw new Error(errorData.error || errorData.message || 'Token refresh failed');
      }

      const payloadRaw = isRecord(data) && isRecord(data.data) ? data.data : data;
      if (!isRecord(payloadRaw) || typeof payloadRaw.access_token !== 'string') {
        throw new Error('Invalid refresh response');
      }

      const payload = payloadRaw as unknown as LoginResponse;

      // Update tokens
      if (typeof window !== 'undefined') {
        localStorage.setItem('access_token', payload.access_token);
        localStorage.setItem('refresh_token', payload.refresh_token);
        if (payload.user) {
          localStorage.setItem('user', JSON.stringify(payload.user));
        }
        this.setAccessTokenCookie(payload.access_token);
      }

      return payload;
    } catch (error) {
      console.error('Refresh token error:', error);
      // Clear tokens on refresh failure
      this.clearTokens();
      throw error;
    }
  }

  getAccessToken(): string | null {
    if (typeof window === 'undefined') return null;
    const token = localStorage.getItem('access_token');
    if (!token) return null;
    
    if (isTokenExpired(token)) {
      console.warn("Access token expired, returning null");
      return null;
    }
    
    return token;
  }

  getRefreshToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('refresh_token');
  }

  getUser(): User | null {
    if (typeof window === 'undefined') return null;
    const userStr = localStorage.getItem('user');
    try {
      return userStr ? JSON.parse(userStr) : null;
    } catch {
      return null;
    }
  }

  isAuthenticated(): boolean {
    return !!this.getAccessToken();
  }

  clearTokens(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
      this.clearAccessTokenCookie();
    }
  }

  // Helper to get auth headers
  getAuthHeaders(): Record<string, string> {
    const token = this.getAccessToken();
    return {
      'Content-Type': 'application/json',
      ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
    };
  }
}

export const authService = new AuthService();
