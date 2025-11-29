import { useAuthStore } from '@/store/auth.store';
import { useCallback } from 'react';

export const useAuth = () => {
  const auth = useAuthStore();

  const login = useCallback(
    async (email: string, password: string) => {
      auth.setLoading(true);
      try {
        // API call would go here
        // const response = await api.post('/auth/login', { email, password });
        // auth.setAuth(response.data.user, response.data.token, response.data.refreshToken);
        console.log('Login:', email);
      } catch (error) {
        console.error('Login error:', error);
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
        // API call would go here
        // const response = await api.post('/auth/register', { firstName, lastName, email, password, role });
        // auth.setAuth(response.data.user, response.data.token, response.data.refreshToken);
        console.log('Register:', { firstName, lastName, email, role });
      } catch (error) {
        console.error('Register error:', error);
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
