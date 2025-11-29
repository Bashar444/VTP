import { LoginRequest, LoginResponse, User, RegisterRequest } from '@/types/api';
import { api } from './api.client';

export class AuthService {
  static async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>(
      '/auth/login',
      credentials
    );
    return response.data;
  }

  static async register(data: RegisterRequest): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>(
      '/auth/register',
      data
    );
    return response.data;
  }

  static async refreshToken(token: string): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>(
      '/auth/refresh',
      { refreshToken: token }
    );
    return response.data;
  }

  static async logout(): Promise<void> {
    await api.post('/auth/logout', {});
  }

  static async verifyEmail(email: string): Promise<{ available: boolean }> {
    const response = await api.post<{ available: boolean }>(
      '/auth/verify-email',
      { email }
    );
    return response.data;
  }

  static async forgotPassword(email: string): Promise<{ success: boolean }> {
    const response = await api.post<{ success: boolean }>(
      '/auth/forgot-password',
      { email }
    );
    return response.data;
  }

  static async resetPassword(
    token: string,
    password: string
  ): Promise<{ success: boolean }> {
    const response = await api.post<{ success: boolean }>(
      '/auth/reset-password',
      { token, password }
    );
    return response.data;
  }

  static async getCurrentUser(): Promise<User> {
    const response = await api.get<User>('/auth/me');
    return response.data;
  }

  static async updateProfile(data: Partial<User>): Promise<User> {
    const response = await api.put<User>('/auth/profile', data);
    return response.data;
  }
}
