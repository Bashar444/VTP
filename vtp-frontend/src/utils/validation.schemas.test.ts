import { describe, it, expect } from 'vitest';
import {
  loginSchema,
  registerSchema,
  passwordResetSchema,
  newPasswordSchema,
} from '@/utils/validation.schemas';

describe('Validation Schemas', () => {
  describe('loginSchema', () => {
    it('validates valid login data', () => {
      const data = {
        email: 'test@example.com',
        password: 'password123',
        rememberMe: true,
      };

      const result = loginSchema.parse(data);
      expect(result).toEqual(data);
    });

    it('rejects invalid email', () => {
      const data = {
        email: 'invalid-email',
        password: 'password123',
      };

      expect(() => loginSchema.parse(data)).toThrow();
    });

    it('requires password', () => {
      const data = {
        email: 'test@example.com',
        password: '',
      };

      expect(() => loginSchema.parse(data)).toThrow();
    });
  });

  describe('registerSchema', () => {
    it('validates valid registration data', () => {
      const data = {
        firstName: 'John',
        lastName: 'Doe',
        email: 'john@example.com',
        password: 'Password123!',
        confirmPassword: 'Password123!',
        role: 'student' as const,
      };

      const result = registerSchema.parse(data);
      expect(result).toEqual(data);
    });

    it('rejects weak password', () => {
      const data = {
        firstName: 'John',
        lastName: 'Doe',
        email: 'john@example.com',
        password: 'weak',
        confirmPassword: 'weak',
        role: 'student' as const,
      };

      expect(() => registerSchema.parse(data)).toThrow();
    });

    it('rejects mismatched passwords', () => {
      const data = {
        firstName: 'John',
        lastName: 'Doe',
        email: 'john@example.com',
        password: 'Password123!',
        confirmPassword: 'Password456!',
        role: 'student' as const,
      };

      expect(() => registerSchema.parse(data)).toThrow();
    });

    it('requires uppercase letter in password', () => {
      const data = {
        firstName: 'John',
        lastName: 'Doe',
        email: 'john@example.com',
        password: 'password123!',
        confirmPassword: 'password123!',
        role: 'student' as const,
      };

      expect(() => registerSchema.parse(data)).toThrow();
    });

    it('requires number in password', () => {
      const data = {
        firstName: 'John',
        lastName: 'Doe',
        email: 'john@example.com',
        password: 'Password!',
        confirmPassword: 'Password!',
        role: 'student' as const,
      };

      expect(() => registerSchema.parse(data)).toThrow();
    });

    it('requires special character in password', () => {
      const data = {
        firstName: 'John',
        lastName: 'Doe',
        email: 'john@example.com',
        password: 'Password123',
        confirmPassword: 'Password123',
        role: 'student' as const,
      };

      expect(() => registerSchema.parse(data)).toThrow();
    });
  });

  describe('passwordResetSchema', () => {
    it('validates valid password reset email', () => {
      const data = {
        email: 'test@example.com',
      };

      const result = passwordResetSchema.parse(data);
      expect(result).toEqual(data);
    });

    it('rejects invalid email', () => {
      const data = {
        email: 'invalid-email',
      };

      expect(() => passwordResetSchema.parse(data)).toThrow();
    });
  });

  describe('newPasswordSchema', () => {
    it('validates valid new password data', () => {
      const data = {
        password: 'NewPassword123!',
        confirmPassword: 'NewPassword123!',
      };

      const result = newPasswordSchema.parse(data);
      expect(result).toEqual(data);
    });

    it('rejects mismatched passwords', () => {
      const data = {
        password: 'NewPassword123!',
        confirmPassword: 'DifferentPassword456!',
      };

      expect(() => newPasswordSchema.parse(data)).toThrow();
    });
  });
});
