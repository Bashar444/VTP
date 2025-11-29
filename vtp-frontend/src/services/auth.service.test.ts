import { describe, it, expect, vi, beforeEach } from 'vitest';
import { AuthService } from '@/services/auth.service';
import { api } from '@/services/api.client';

vi.mock('@/services/api.client');

describe('AuthService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('login', () => {
    it('sends login request with credentials', async () => {
      const mockResponse = {
        data: {
          token: 'test-token',
          refreshToken: 'test-refresh-token',
          user: {
            id: '1',
            email: 'test@example.com',
            firstName: 'Test',
            lastName: 'User',
            role: 'student' as const,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        },
      };

      vi.mocked(api.post).mockResolvedValue(mockResponse);

      const result = await AuthService.login({
        email: 'test@example.com',
        password: 'password123',
      });

      expect(api.post).toHaveBeenCalledWith('/auth/login', {
        email: 'test@example.com',
        password: 'password123',
      });
      expect(result).toEqual(mockResponse.data);
    });

    it('handles login errors', async () => {
      const mockError = new Error('Invalid credentials');
      vi.mocked(api.post).mockRejectedValue(mockError);

      await expect(
        AuthService.login({
          email: 'test@example.com',
          password: 'wrong-password',
        })
      ).rejects.toThrow('Invalid credentials');
    });
  });

  describe('register', () => {
    it('sends register request with user data', async () => {
      const mockResponse = {
        data: {
          token: 'test-token',
          refreshToken: 'test-refresh-token',
          user: {
            id: '2',
            email: 'newuser@example.com',
            firstName: 'New',
            lastName: 'User',
            role: 'student' as const,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        },
      };

      vi.mocked(api.post).mockResolvedValue(mockResponse);

      const result = await AuthService.register({
        firstName: 'New',
        lastName: 'User',
        email: 'newuser@example.com',
        password: 'Password123!',
        role: 'student',
      });

      expect(api.post).toHaveBeenCalledWith('/auth/register', {
        firstName: 'New',
        lastName: 'User',
        email: 'newuser@example.com',
        password: 'Password123!',
        role: 'student',
      });
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('refreshToken', () => {
    it('refreshes authentication token', async () => {
      const mockResponse = {
        data: {
          token: 'new-token',
          refreshToken: 'new-refresh-token',
          user: {
            id: '1',
            email: 'test@example.com',
            firstName: 'Test',
            lastName: 'User',
            role: 'student' as const,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        },
      };

      vi.mocked(api.post).mockResolvedValue(mockResponse);

      const result = await AuthService.refreshToken('old-refresh-token');

      expect(api.post).toHaveBeenCalledWith('/auth/refresh', {
        refreshToken: 'old-refresh-token',
      });
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('logout', () => {
    it('sends logout request', async () => {
      vi.mocked(api.post).mockResolvedValue({ data: {} });

      await AuthService.logout();

      expect(api.post).toHaveBeenCalledWith('/auth/logout', {});
    });
  });

  describe('verifyEmail', () => {
    it('checks if email is available', async () => {
      const mockResponse = {
        data: { available: true },
      };

      vi.mocked(api.post).mockResolvedValue(mockResponse);

      const result = await AuthService.verifyEmail('test@example.com');

      expect(api.post).toHaveBeenCalledWith('/auth/verify-email', {
        email: 'test@example.com',
      });
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('getCurrentUser', () => {
    it('fetches current user profile', async () => {
      const mockUser = {
        id: '1',
        email: 'test@example.com',
        firstName: 'Test',
        lastName: 'User',
        role: 'student' as const,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      vi.mocked(api.get).mockResolvedValue({ data: mockUser });

      const result = await AuthService.getCurrentUser();

      expect(api.get).toHaveBeenCalledWith('/auth/me');
      expect(result).toEqual(mockUser);
    });
  });
});
