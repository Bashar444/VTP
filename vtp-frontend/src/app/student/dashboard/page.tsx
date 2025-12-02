"use client";

import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import Link from 'next/link';

interface EnrolledCourse {
  id: string;
  course_id: string;
  course_title: string;
  progress: number;
  last_accessed: string;
}

interface LiveStream {
  id: string;
  room_id: string;
  title: string;
  instructor_name: string;
  is_live: boolean;
}

interface Assignment {
  id: string;
  title: string;
  course_title: string;
  due_date: string;
  status: 'pending' | 'submitted' | 'graded';
}

export default function StudentDashboardPage() {
  const { user, token } = useAuth();
  const router = useRouter();
  const [enrolledCourses, setEnrolledCourses] = useState<EnrolledCourse[]>([]);
  const [liveStreams, setLiveStreams] = useState<LiveStream[]>([]);
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!user) {
      router.push('/login');
      return;
    }

    fetchStudentData();
  }, [user, router]);

  const fetchStudentData = async () => {
    try {
      setLoading(true);

      // Fetch enrolled courses
      const coursesRes = await fetch('http://localhost:8080/api/v1/enrollments/my-courses', {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (coursesRes.ok) {
        const data = await coursesRes.json();
        setEnrolledCourses(data || []);
      }

      // Fetch live streams
      const streamsRes = await fetch('http://localhost:8080/api/v1/streaming/sessions?live=true', {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (streamsRes.ok) {
        const data = await streamsRes.json();
        setLiveStreams(data.sessions || []);
      }

      // Fetch assignments
      const assignmentsRes = await fetch('http://localhost:8080/api/v1/assignments/my-assignments', {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (assignmentsRes.ok) {
        const data = await assignmentsRes.json();
        setAssignments(data || []);
      }
    } catch (err) {
      console.error('Failed to fetch data:', err);
    } finally {
      setLoading(false);
    }
  };

  if (!user) return null;

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
        {/* Welcome Header */}
        <div className="bg-gradient-to-l from-blue-600 to-indigo-700 rounded-xl p-8 mb-8 text-white">
          <h1 className="text-3xl font-bold mb-2">مرحباً {user.full_name} 👋</h1>
          <p className="text-blue-100">لوحة تحكم الطالب - منصة التعليم</p>
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-xl shadow p-6">
            <div className="text-3xl font-bold text-indigo-600 mb-2">{enrolledCourses.length}</div>
            <div className="text-gray-600">المواد المسجلة</div>
          </div>
          <div className="bg-white rounded-xl shadow p-6">
            <div className="text-3xl font-bold text-green-600 mb-2">{liveStreams.length}</div>
            <div className="text-gray-600">بث مباشر الآن</div>
          </div>
          <div className="bg-white rounded-xl shadow p-6">
            <div className="text-3xl font-bold text-orange-600 mb-2">
              {assignments.filter(a => a.status === 'pending').length}
            </div>
            <div className="text-gray-600">واجبات معلقة</div>
          </div>
          <div className="bg-white rounded-xl shadow p-6">
            <div className="text-3xl font-bold text-purple-600 mb-2">
              {assignments.filter(a => a.status === 'graded').length}
            </div>
            <div className="text-gray-600">واجبات مصححة</div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Live Streams */}
          <div className="bg-white rounded-xl shadow p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-gray-900">🔴 البث المباشر</h2>
            </div>
            {liveStreams.length === 0 ? (
              <p className="text-gray-500 text-center py-8">لا يوجد بث مباشر حالياً</p>
            ) : (
              <div className="space-y-4">
                {liveStreams.map(stream => (
                  <div key={stream.id} className="border border-red-200 bg-red-50 rounded-lg p-4">
                    <div className="flex items-center gap-2 mb-2">
                      <span className="w-3 h-3 bg-red-500 rounded-full animate-pulse"></span>
                      <span className="text-sm font-medium text-red-600">مباشر الآن</span>
                    </div>
                    <h3 className="font-semibold text-gray-900">{stream.title}</h3>
                    <p className="text-sm text-gray-600 mb-3">{stream.instructor_name}</p>
                    <Link
                      href={`/stream/${stream.room_id}`}
                      className="block text-center bg-red-600 text-white py-2 rounded-lg hover:bg-red-700"
                    >
                      انضم للبث
                    </Link>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* My Courses */}
          <div className="bg-white rounded-xl shadow p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-gray-900">📚 موادي</h2>
              <Link href="/courses" className="text-indigo-600 hover:underline text-sm">
                تصفح المزيد
              </Link>
            </div>
            {enrolledCourses.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-gray-500 mb-4">لم تسجل في أي مادة بعد</p>
                <Link
                  href="/courses"
                  className="inline-block bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700"
                >
                  تصفح المواد
                </Link>
              </div>
            ) : (
              <div className="space-y-4">
                {enrolledCourses.slice(0, 5).map(course => (
                  <Link
                    key={course.id}
                    href={`/courses/${course.course_id}`}
                    className="block border rounded-lg p-4 hover:bg-gray-50"
                  >
                    <h3 className="font-semibold text-gray-900">{course.course_title}</h3>
                    <div className="mt-2">
                      <div className="flex justify-between text-sm text-gray-600 mb-1">
                        <span>التقدم</span>
                        <span>{course.progress}%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div
                          className="bg-indigo-600 h-2 rounded-full"
                          style={{ width: `${course.progress}%` }}
                        ></div>
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>

          {/* Assignments */}
          <div className="bg-white rounded-xl shadow p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-gray-900">📝 الواجبات</h2>
              <Link href="/assignments" className="text-indigo-600 hover:underline text-sm">
                عرض الكل
              </Link>
            </div>
            {assignments.length === 0 ? (
              <p className="text-gray-500 text-center py-8">لا توجد واجبات حالياً</p>
            ) : (
              <div className="space-y-4">
                {assignments.slice(0, 5).map(assignment => (
                  <div key={assignment.id} className="border rounded-lg p-4">
                    <h3 className="font-semibold text-gray-900">{assignment.title}</h3>
                    <p className="text-sm text-gray-600">{assignment.course_title}</p>
                    <div className="flex items-center justify-between mt-2">
                      <span className="text-xs text-gray-500">
                        موعد التسليم: {new Date(assignment.due_date).toLocaleDateString('ar-SA')}
                      </span>
                      <span className={`text-xs px-2 py-1 rounded-full ${
                        assignment.status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                        assignment.status === 'submitted' ? 'bg-blue-100 text-blue-800' :
                        'bg-green-100 text-green-800'
                      }`}>
                        {assignment.status === 'pending' ? 'معلق' :
                         assignment.status === 'submitted' ? 'تم التسليم' : 'مصحح'}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Quick Actions */}
        <div className="mt-8 bg-white rounded-xl shadow p-6">
          <h2 className="text-xl font-bold text-gray-900 mb-4">🚀 إجراءات سريعة</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Link
              href="/courses"
              className="flex flex-col items-center p-4 border rounded-lg hover:bg-gray-50"
            >
              <span className="text-3xl mb-2">📖</span>
              <span className="text-gray-700">تصفح المواد</span>
            </Link>
            <Link
              href="/my-courses"
              className="flex flex-col items-center p-4 border rounded-lg hover:bg-gray-50"
            >
              <span className="text-3xl mb-2">📚</span>
              <span className="text-gray-700">موادي</span>
            </Link>
            <Link
              href="/assignments"
              className="flex flex-col items-center p-4 border rounded-lg hover:bg-gray-50"
            >
              <span className="text-3xl mb-2">📝</span>
              <span className="text-gray-700">الواجبات</span>
            </Link>
            <Link
              href="/profile"
              className="flex flex-col items-center p-4 border rounded-lg hover:bg-gray-50"
            >
              <span className="text-3xl mb-2">👤</span>
              <span className="text-gray-700">الملف الشخصي</span>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
