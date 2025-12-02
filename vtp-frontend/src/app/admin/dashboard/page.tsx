"use client";

import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import Link from 'next/link';

interface DashboardStats {
  total_users: number;
  total_students: number;
  total_instructors: number;
  total_courses: number;
  active_streams: number;
  total_enrollments: number;
}

export default function AdminDashboardPage() {
  const { user, token } = useAuth();
  const router = useRouter();
  const [stats, setStats] = useState<DashboardStats>({
    total_users: 0,
    total_students: 0,
    total_instructors: 0,
    total_courses: 0,
    active_streams: 0,
    total_enrollments: 0
  });
  const [recentUsers, setRecentUsers] = useState<any[]>([]);
  const [recentCourses, setRecentCourses] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!user) {
      router.push('/login');
      return;
    }
    if (user.role !== 'admin') {
      router.push('/my-courses');
      return;
    }

    fetchDashboardData();
  }, [user, router]);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      
      // Fetch stats
      const statsRes = await fetch('http://localhost:8080/api/v1/admin/stats', {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (statsRes.ok) {
        const data = await statsRes.json();
        setStats(data);
      }

      // Fetch recent users
      const usersRes = await fetch('http://localhost:8080/api/v1/admin/users?limit=5', {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (usersRes.ok) {
        const data = await usersRes.json();
        setRecentUsers(data.users || []);
      }

      // Fetch recent courses
      const coursesRes = await fetch('http://localhost:8080/api/v1/courses?limit=5', {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (coursesRes.ok) {
        const data = await coursesRes.json();
        setRecentCourses(data.courses || []);
      }
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
    } finally {
      setLoading(false);
    }
  };

  if (!user || user.role !== 'admin') return null;

  if (loading) {
    return (
      <div dir="rtl" className="min-h-screen bg-gray-50 pt-24 pb-12">
        <div className="max-w-7xl mx-auto px-4">
          <div className="text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">جاري التحميل...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div dir="rtl" className="min-h-screen bg-gray-50 pt-24 pb-12">
      <div className="max-w-7xl mx-auto px-4">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">لوحة تحكم المدير</h1>
          <p className="text-gray-600 mt-1">مرحباً، {user.full_name} - نظرة عامة على المنصة</p>
        </div>

        {/* Quick Navigation */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <Link
            href="/admin/users"
            className="bg-white rounded-lg shadow p-6 hover:shadow-md transition flex items-center gap-4"
          >
            <div className="bg-blue-100 p-3 rounded-full">
              <span className="text-2xl">👥</span>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900">إدارة المستخدمين</h3>
              <p className="text-sm text-gray-500">عرض وإدارة جميع المستخدمين</p>
            </div>
          </Link>
          <Link
            href="/admin/courses"
            className="bg-white rounded-lg shadow p-6 hover:shadow-md transition flex items-center gap-4"
          >
            <div className="bg-green-100 p-3 rounded-full">
              <span className="text-2xl">📚</span>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900">إدارة الدورات</h3>
              <p className="text-sm text-gray-500">عرض وإدارة جميع الدورات</p>
            </div>
          </Link>
          <Link
            href="/stream"
            className="bg-white rounded-lg shadow p-6 hover:shadow-md transition flex items-center gap-4"
          >
            <div className="bg-purple-100 p-3 rounded-full">
              <span className="text-2xl">🎥</span>
            </div>
            <div>
              <h3 className="font-semibold text-gray-900">البث المباشر</h3>
              <p className="text-sm text-gray-500">مراقبة البثوث النشطة</p>
            </div>
          </Link>
        </div>

        {/* Statistics Cards */}
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-8">
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-2xl font-bold text-blue-600">{stats.total_users}</div>
            <div className="text-sm text-gray-600">إجمالي المستخدمين</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-2xl font-bold text-green-600">{stats.total_students}</div>
            <div className="text-sm text-gray-600">الطلاب</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-2xl font-bold text-purple-600">{stats.total_instructors}</div>
            <div className="text-sm text-gray-600">المعلمين</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-2xl font-bold text-orange-600">{stats.total_courses}</div>
            <div className="text-sm text-gray-600">الدورات</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-2xl font-bold text-red-600">{stats.active_streams}</div>
            <div className="text-sm text-gray-600">بث مباشر</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-2xl font-bold text-indigo-600">{stats.total_enrollments}</div>
            <div className="text-sm text-gray-600">التسجيلات</div>
          </div>
        </div>

        {/* Recent Activity */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Recent Users */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-4 border-b flex justify-between items-center">
              <h2 className="text-lg font-semibold">أحدث المستخدمين</h2>
              <Link href="/admin/users" className="text-sm text-indigo-600 hover:text-indigo-700">
                عرض الكل
              </Link>
            </div>
            <div className="p-4">
              {recentUsers.length === 0 ? (
                <p className="text-gray-500 text-center py-4">لا يوجد مستخدمين</p>
              ) : (
                <div className="space-y-3">
                  {recentUsers.map((u: any) => (
                    <div key={u.id} className="flex items-center justify-between py-2 border-b last:border-0">
                      <div>
                        <div className="font-medium">{u.full_name}</div>
                        <div className="text-sm text-gray-500">{u.email}</div>
                      </div>
                      <span className={`text-xs px-2 py-1 rounded-full ${
                        u.role === 'admin' ? 'bg-red-100 text-red-800' :
                        u.role === 'instructor' ? 'bg-purple-100 text-purple-800' :
                        'bg-blue-100 text-blue-800'
                      }`}>
                        {u.role === 'admin' ? 'مدير' : u.role === 'instructor' ? 'معلم' : 'طالب'}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Recent Courses */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-4 border-b flex justify-between items-center">
              <h2 className="text-lg font-semibold">أحدث الدورات</h2>
              <Link href="/admin/courses" className="text-sm text-indigo-600 hover:text-indigo-700">
                عرض الكل
              </Link>
            </div>
            <div className="p-4">
              {recentCourses.length === 0 ? (
                <p className="text-gray-500 text-center py-4">لا يوجد دورات</p>
              ) : (
                <div className="space-y-3">
                  {recentCourses.map((c: any) => (
                    <div key={c.id} className="flex items-center justify-between py-2 border-b last:border-0">
                      <div>
                        <div className="font-medium">{c.title}</div>
                        <div className="text-sm text-gray-500">{c.category}</div>
                      </div>
                      <span className={`text-xs px-2 py-1 rounded-full ${
                        c.is_published ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
                      }`}>
                        {c.is_published ? 'منشورة' : 'مسودة'}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Subject Categories (9 subjects) */}
        <div className="mt-8">
          <h2 className="text-xl font-bold text-gray-900 mb-4">المواد الدراسية</h2>
          <div className="grid grid-cols-3 md:grid-cols-5 lg:grid-cols-9 gap-4">
            {[
              { name: 'الرياضيات', icon: '🔢', color: 'bg-blue-100' },
              { name: 'الفيزياء', icon: '⚛️', color: 'bg-purple-100' },
              { name: 'الكيمياء', icon: '🧪', color: 'bg-green-100' },
              { name: 'الأحياء', icon: '🧬', color: 'bg-teal-100' },
              { name: 'العربية', icon: '📖', color: 'bg-orange-100' },
              { name: 'الإنجليزية', icon: '🔤', color: 'bg-red-100' },
              { name: 'التاريخ', icon: '📜', color: 'bg-yellow-100' },
              { name: 'الجغرافيا', icon: '🌍', color: 'bg-indigo-100' },
              { name: 'الفلسفة', icon: '💭', color: 'bg-pink-100' },
            ].map((subject) => (
              <div key={subject.name} className={`${subject.color} rounded-lg p-4 text-center`}>
                <div className="text-2xl mb-2">{subject.icon}</div>
                <div className="text-sm font-medium">{subject.name}</div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
