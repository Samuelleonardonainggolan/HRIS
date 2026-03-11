// lib/api/auth.ts
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

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
}

export interface ErrorResponse {
  success: false;
  message: string;
  error: string;
}

class AuthService {
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    try {
      const response = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(credentials),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || 'Login failed');
      }

      // ✅ Validate response structure
      if (!data.access_token || !data.user) {
        console.error('Invalid response structure:', data);
        throw new Error('Invalid response from server');
      }

      // Save tokens to localStorage
      if (typeof window !== 'undefined') {
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token', data.refresh_token);
        localStorage.setItem('user', JSON.stringify(data.user));
      }

      return data;
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  }

  async logout(): Promise<void> {
    const token = this.getAccessToken();

    if (token) {
      try {
        await fetch(`${API_URL}/logout`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
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
    }
  }

  async refreshToken(): Promise<LoginResponse> {
    const refreshToken = this.getRefreshToken();

    if (!refreshToken) {
      throw new Error('No refresh token available');
    }

    try {
      const response = await fetch(`${API_URL}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || 'Token refresh failed');
      }

      // ✅ Validate response
      if (!data.access_token) {
        throw new Error('Invalid refresh response');
      }

      // Update tokens
      if (typeof window !== 'undefined') {
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token', data.refresh_token);
        if (data.user) {
          localStorage.setItem('user', JSON.stringify(data.user));
        }
      }

      return data;
    } catch (error) {
      console.error('Refresh token error:', error);
      // Clear tokens on refresh failure
      this.clearTokens();
      throw error;
    }
  }

  getAccessToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('access_token');
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