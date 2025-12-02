"use client";

import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function ProfilePage() {
  const { user } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!user) {
      router.push('/login');
    }
  }, [user, router]);

  if (!user) return null;

  const getRoleLabel = (role: string) => {
    switch (role) {
      case 'student': return 'طالب';
      case 'teacher': return 'مدرّس';
      case 'admin': return 'مدير';
      default: return role;
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12" dir="rtl">
      <div className="max-w-3xl mx-auto px-4">
        <h1 className="text-3xl font-bold text-white mb-8">الملف الشخصي</h1>

        <div className="bg-gray-800 rounded-lg shadow p-6 space-y-6">
          {/* Profile Info */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              الاسم الكامل
            </label>
            <input
              type="text"
              value={user.full_name}
              disabled
              className="w-full px-4 py-2 border border-gray-600 rounded-md bg-gray-700 text-white"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              البريد الإلكتروني
            </label>
            <input
              type="email"
              value={user.email}
              disabled
              className="w-full px-4 py-2 border border-gray-600 rounded-md bg-gray-700 text-white"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              الدور
            </label>
            <input
              type="text"
              value={getRoleLabel(user.role)}
              disabled
              className="w-full px-4 py-2 border border-gray-600 rounded-md bg-gray-700 text-white"
            />
          </div>

          {user.phone && (
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                رقم الهاتف
              </label>
              <input
                type="text"
                value={user.phone}
                disabled
                className="w-full px-4 py-2 border border-gray-600 rounded-md bg-gray-700 text-white"
              />
            </div>
          )}

          {/* Change Password */}
          <div className="pt-4 border-t border-gray-600">
            <h2 className="text-lg font-semibold text-white mb-4">تغيير كلمة المرور</h2>
            <button className="text-blue-400 hover:text-blue-300 font-medium">
              تحديث كلمة المرور
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
