'use client';

import Link from 'next/link';
import { ForgotPasswordForm } from '@/components/auth/PasswordForms';

export default function ForgotPasswordPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 sm:px-6 lg:px-8">
      <div className="w-full max-w-md space-y-8">
        <div>
          <h1 className="text-center text-3xl font-extrabold text-gray-900">
            VTP
          </h1>
          <h2 className="mt-2 text-center text-xl font-bold text-gray-900">
            Reset your password
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Enter your email to receive password reset instructions
          </p>
        </div>

        <div className="bg-white py-12 px-6 shadow rounded-lg sm:px-12">
          <ForgotPasswordForm />

          <div className="mt-6 text-center">
            <Link
              href="/login"
              className="text-sm font-medium text-blue-600 hover:text-blue-500"
            >
              Back to sign in
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
