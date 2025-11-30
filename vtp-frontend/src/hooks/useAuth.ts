import { useAuthStore } from '@/store/auth.store';
import { useCallback } from 'react';
import { AuthService } from '@/services/auth.service';

export const useAuth = () => {
  const auth = useAuthStore();

  const login = useCallback(
    async (email: string, password: string) => {
      auth.setLoading(true);
      try {
        const res = await AuthService.login({ email, password });
        auth.setAuth(res.user, res.token, res.refreshToken);
        return res;
      } catch (error) {
        throw error;
      } finally {
        auth.setLoading(false);
      }
    },
    [auth]
  );

  const register = useCallback(
    async (
      firstName: string,
      lastName: string,
      email: string,
      password: string,
      role: 'student' | 'instructor' | 'admin'
    ) => {
      auth.setLoading(true);
      try {
        const res = await AuthService.register({ firstName, lastName, email, password, role });
        auth.setAuth(res.user, res.token, res.refreshToken);
        return res;
      } catch (error) {
        throw error;
      } finally {
        auth.setLoading(false);
      }
    },
    [auth]
  );

  const logout = useCallback(() => {
    auth.clearAuth();
  }, [auth]);

  const isAuthenticated = !!auth.token;

  return {
    user: auth.user,
    token: auth.token,
    isAuthenticated,
    isLoading: auth.isLoading,
    login,
    register,
    logout,
  };
};
