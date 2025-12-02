"use client";

import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import Link from 'next/link';

interface Course {
  id: string;
  title: string;
  description: string;
  instructor_id: string;
  instructor_name?: string;
  category: string;
  level: string;
  price: number;
  currency: string;
  student_count: number;
  is_published: boolean;
  created_at: string;
}

export default function AdminCoursesPage() {
  const { user, token } = useAuth();
  const router = useRouter();
  const [courses, setCourses] = useState<Course[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const categories = [
    { value: 'all', label: 'جميع التصنيفات' },
    { value: 'mathematics', label: 'الرياضيات' },
    { value: 'sciences', label: 'العلوم' },
    { value: 'languages', label: 'اللغات' },
    { value: 'social_studies', label: 'الدراسات الاجتماعية' },
    { value: 'humanities', label: 'العلوم الإنسانية' },
  ];

  useEffect(() => {
    if (!user) {
      router.push('/login');
      return;
    }
    if (user.role !== 'admin') {
      router.push('/dashboard');
      return;
    }

    fetchCourses();
  }, [user, router, page, categoryFilter]);

  const fetchCourses = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams({
        page: page.toString(),
        limit: '20',
        ...(categoryFilter !== 'all' && { category: categoryFilter })
      });

      const res = await fetch(`http://localhost:8080/api/v1/courses?${params}`, {
        headers: { Authorization: `Bearer ${token}` }
      });

      if (res.ok) {
        const data = await res.json();
        setCourses(data.courses || []);
        setTotalPages(data.total_pages || 1);
      }
    } catch (err) {
      console.error('Failed to fetch courses:', err);
    } finally {
      setLoading(false);
    }
  };

  const togglePublish = async (courseId: string, currentStatus: boolean) => {
    try {
      const res = await fetch(`http://localhost:8080/api/v1/courses/${courseId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ is_published: !currentStatus })
      });

      if (res.ok) {
        fetchCourses();
      }
    } catch (err) {
      console.error('Failed to update course:', err);
    }
  };

  const deleteCourse = async (courseId: string) => {
    if (!confirm('هل أنت متأكد من حذف هذه الدورة؟')) return;

    try {
      const res = await fetch(`http://localhost:8080/api/v1/courses/${courseId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` }
      });

      if (res.ok) {
        fetchCourses();
      }
    } catch (err) {
      console.error('Failed to delete course:', err);
    }
  };

  const filteredCourses = courses.filter(c => 
    c.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
    c.description?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (!user || user.role !== 'admin') return null;

  return (
    <div dir="rtl" className="min-h-screen bg-gray-50 pt-24 pb-12">
      <div className="max-w-7xl mx-auto px-4">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">إدارة الدورات</h1>
            <p className="text-gray-600 mt-1">عرض وإدارة جميع دورات المنصة</p>
          </div>
          <div className="flex gap-3">
            <Link
              href="/admin/dashboard"
              className="text-indigo-600 hover:text-indigo-700"
            >
              ← العودة للوحة التحكم
            </Link>
          </div>
        </div>

        {/* Subject Quick Filters */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <h3 className="text-sm font-medium text-gray-700 mb-3">المواد الدراسية</h3>
          <div className="flex flex-wrap gap-2">
            {[
              { name: 'الرياضيات', category: 'mathematics' },
              { name: 'الفيزياء', category: 'sciences' },
              { name: 'الكيمياء', category: 'sciences' },
              { name: 'الأحياء', category: 'sciences' },
              { name: 'العربية', category: 'languages' },
              { name: 'الإنجليزية', category: 'languages' },
              { name: 'التاريخ', category: 'social_studies' },
              { name: 'الجغرافيا', category: 'social_studies' },
              { name: 'الفلسفة', category: 'humanities' },
            ].map((subject) => (
              <button
                key={subject.name}
                onClick={() => setCategoryFilter(subject.category)}
                className={`px-3 py-1 rounded-full text-sm ${
                  categoryFilter === subject.category
                    ? 'bg-indigo-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {subject.name}
              </button>
            ))}
            <button
              onClick={() => setCategoryFilter('all')}
              className={`px-3 py-1 rounded-full text-sm ${
                categoryFilter === 'all'
                  ? 'bg-indigo-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              الكل
            </button>
          </div>
        </div>

        {/* Filters */}
        <div className="bg-white rounded-lg shadow p-4 mb-6">
          <div className="flex flex-wrap gap-4">
            <div className="flex-1 min-w-[200px]">
              <input
                type="text"
                placeholder="بحث في الدورات..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full border rounded-md px-3 py-2"
              />
            </div>
            <select
              value={categoryFilter}
              onChange={(e) => setCategoryFilter(e.target.value)}
              className="border rounded-md px-3 py-2"
            >
              {categories.map((cat) => (
                <option key={cat.value} value={cat.value}>
                  {cat.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Courses Grid */}
        {loading ? (
          <div className="text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">جاري التحميل...</p>
          </div>
        ) : filteredCourses.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-12 text-center">
            <p className="text-gray-600">لا توجد دورات</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredCourses.map((course) => (
              <div key={course.id} className="bg-white rounded-lg shadow overflow-hidden">
                <div className="p-6">
                  <div className="flex justify-between items-start mb-3">
                    <h3 className="text-lg font-semibold text-gray-900 line-clamp-2">
                      {course.title}
                    </h3>
                    <span className={`text-xs px-2 py-1 rounded-full flex-shrink-0 ${
                      course.is_published 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-yellow-100 text-yellow-800'
                    }`}>
                      {course.is_published ? 'منشورة' : 'مسودة'}
                    </span>
                  </div>
                  
                  <p className="text-sm text-gray-600 mb-4 line-clamp-2">
                    {course.description}
                  </p>

                  <div className="flex flex-wrap gap-2 mb-4">
                    <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                      {course.category}
                    </span>
                    <span className="text-xs bg-purple-100 text-purple-800 px-2 py-1 rounded">
                      {course.level}
                    </span>
                    <span className="text-xs bg-gray-100 text-gray-800 px-2 py-1 rounded">
                      {course.student_count || 0} طالب
                    </span>
                  </div>

                  <div className="text-sm text-gray-500 mb-4">
                    السعر: {course.price === 0 ? 'مجاني' : `${course.price} ${course.currency}`}
                  </div>

                  <div className="flex gap-2">
                    <button
                      onClick={() => togglePublish(course.id, course.is_published)}
                      className={`flex-1 text-sm py-2 rounded ${
                        course.is_published
                          ? 'bg-yellow-100 text-yellow-800 hover:bg-yellow-200'
                          : 'bg-green-100 text-green-800 hover:bg-green-200'
                      }`}
                    >
                      {course.is_published ? 'إلغاء النشر' : 'نشر'}
                    </button>
                    <Link
                      href={`/courses/${course.id}`}
                      className="flex-1 text-sm py-2 bg-indigo-100 text-indigo-800 rounded text-center hover:bg-indigo-200"
                    >
                      عرض
                    </Link>
                    <button
                      onClick={() => deleteCourse(course.id)}
                      className="text-sm py-2 px-3 bg-red-100 text-red-800 rounded hover:bg-red-200"
                    >
                      حذف
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="mt-6 flex justify-center gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-4 py-2 border rounded disabled:opacity-50"
            >
              السابق
            </button>
            <span className="px-4 py-2">
              صفحة {page} من {totalPages}
            </span>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-4 py-2 border rounded disabled:opacity-50"
            >
              التالي
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
