"use client";
export const dynamic = 'force-dynamic';

import Link from 'next/link';
import { LoginForm } from '@/components/auth/LoginForm';

export default function LoginPage() {
  return (
    <div dir="rtl" className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 sm:px-6 lg:px-8">
      <div className="w-full max-w-md space-y-8">
        <div>
          <h1 className="text-center text-3xl font-extrabold text-gray-900">
            منصة التعليم
          </h1>
          <h2 className="mt-2 text-center text-xl font-bold text-gray-900">
            تسجيل الدخول إلى حسابك
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            منصة تعليمية للطلاب السوريين
          </p>
        </div>

        <div className="bg-white py-12 px-6 shadow rounded-lg sm:px-12">
          <LoginForm />

          <div className="mt-6">
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">
                  ليس لديك حساب؟
                </span>
              </div>
            </div>

            <div className="mt-6">
              <Link
                href="/register"
                className="w-full inline-flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-500 hover:bg-gray-50"
              >
                إنشاء حساب جديد
              </Link>
            </div>
          </div>

          <div className="mt-6 text-center">
            <Link
              href="/forgot-password"
              className="text-sm font-medium text-blue-600 hover:text-blue-500"
            >
              نسيت كلمة المرور؟
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
