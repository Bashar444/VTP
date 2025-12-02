"use client";

import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import Link from 'next/link';

interface Course {
  id: string;
  title: string;
  description: string;
  student_count: number;
  is_published: boolean;
  created_at: string;
}

interface StreamSession {
  id: string;
  room_id: string;
  title: string;
  is_live: boolean;
  created_at: string;
}

export default function InstructorCoursesPage() {
  const { user, token } = useAuth();
  const router = useRouter();
  const [courses, setCourses] = useState<Course[]>([]);
  const [streams, setStreams] = useState<StreamSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateStream, setShowCreateStream] = useState(false);
  const [newStreamTitle, setNewStreamTitle] = useState('');
  const [createdRoomId, setCreatedRoomId] = useState<string | null>(null);

  useEffect(() => {
    if (!user) {
      router.push('/login');
      return;
    }
    if (user.role !== 'teacher' && user.role !== 'instructor' && user.role !== 'admin') {
      router.push('/my-courses');
      return;
    }

    fetchInstructorData();
  }, [user, router]);

  const fetchInstructorData = async () => {
    try {
      setLoading(true);
      // Fetch instructor's courses
      const coursesRes = await fetch('http://localhost:8080/api/v1/courses?instructor=me', {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (coursesRes.ok) {
        const data = await coursesRes.json();
        setCourses(data.courses || []);
      }

      // Fetch active streams
      const streamsRes = await fetch('http://localhost:8080/api/v1/streaming/sessions', {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (streamsRes.ok) {
        const data = await streamsRes.json();
        setStreams(data.sessions || []);
      }
    } catch (err) {
      console.error('Failed to fetch data:', err);
    } finally {
      setLoading(false);
    }
  };

  const createStreamSession = async () => {
    if (!newStreamTitle.trim()) return;
    
    try {
      const roomId = `stream-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
      const res = await fetch('http://localhost:8080/api/v1/streaming/sessions/start', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({
          room_id: roomId,
          title: newStreamTitle
        })
      });

      if (res.ok) {
        setCreatedRoomId(roomId);
        setNewStreamTitle('');
        fetchInstructorData();
      }
    } catch (err) {
      console.error('Failed to create stream:', err);
    }
  };

  const copyStreamLink = (roomId: string) => {
    const link = `${window.location.origin}/stream/${roomId}`;
    navigator.clipboard.writeText(link);
    alert('تم نسخ رابط البث!');
  };

  if (!user || (user.role !== 'teacher' && user.role !== 'admin')) return null;

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
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">لوحة تحكم المعلم</h1>
            <p className="text-gray-600 mt-1">مرحباً، {user.full_name}</p>
          </div>
          <div className="flex gap-3">
            <button
              onClick={() => setShowCreateStream(true)}
              className="bg-green-600 text-white px-4 py-2 rounded-md hover:bg-green-700 flex items-center gap-2"
            >
              <span>🎥</span> إنشاء بث مباشر
            </button>
            <Link
              href="/instructor/courses/new"
              className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700"
            >
              + إضافة دورة جديدة
            </Link>
          </div>
        </div>

        {/* Create Stream Modal */}
        {showCreateStream && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md">
              <h2 className="text-xl font-bold mb-4">إنشاء بث مباشر جديد</h2>
              
              {createdRoomId ? (
                <div className="space-y-4">
                  <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                    <p className="text-green-800 font-medium mb-2">تم إنشاء البث بنجاح!</p>
                    <p className="text-sm text-gray-600 mb-2">رابط البث:</p>
                    <div className="bg-white border rounded p-2 text-sm break-all">
                      {window.location.origin}/stream/{createdRoomId}
                    </div>
                  </div>
                  <div className="flex gap-3">
                    <button
                      onClick={() => copyStreamLink(createdRoomId)}
                      className="flex-1 bg-indigo-600 text-white py-2 rounded-md hover:bg-indigo-700"
                    >
                      نسخ الرابط
                    </button>
                    <Link
                      href={`/stream/${createdRoomId}`}
                      className="flex-1 bg-green-600 text-white py-2 rounded-md hover:bg-green-700 text-center"
                    >
                      بدء البث
                    </Link>
                  </div>
                  <button
                    onClick={() => {
                      setShowCreateStream(false);
                      setCreatedRoomId(null);
                    }}
                    className="w-full text-gray-600 py-2 hover:text-gray-800"
                  >
                    إغلاق
                  </button>
                </div>
              ) : (
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      عنوان البث
                    </label>
                    <input
                      type="text"
                      value={newStreamTitle}
                      onChange={(e) => setNewStreamTitle(e.target.value)}
                      placeholder="مثال: محاضرة الرياضيات - الفصل الأول"
                      className="w-full border rounded-md px-3 py-2"
                    />
                  </div>
                  <div className="flex gap-3">
                    <button
                      onClick={createStreamSession}
                      disabled={!newStreamTitle.trim()}
                      className="flex-1 bg-green-600 text-white py-2 rounded-md hover:bg-green-700 disabled:opacity-50"
                    >
                      إنشاء البث
                    </button>
                    <button
                      onClick={() => setShowCreateStream(false)}
                      className="flex-1 border border-gray-300 py-2 rounded-md hover:bg-gray-50"
                    >
                      إلغاء
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Quick Stats */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="text-3xl font-bold text-indigo-600">{courses.length}</div>
            <div className="text-gray-600">الدورات</div>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <div className="text-3xl font-bold text-green-600">
              {courses.reduce((sum, c) => sum + (c.student_count || 0), 0)}
            </div>
            <div className="text-gray-600">الطلاب المسجلين</div>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <div className="text-3xl font-bold text-orange-600">
              {streams.filter(s => s.is_live).length}
            </div>
            <div className="text-gray-600">بث مباشر الآن</div>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <div className="text-3xl font-bold text-purple-600">{streams.length}</div>
            <div className="text-gray-600">إجمالي البثوث</div>
          </div>
        </div>

        {/* Active Streams */}
        {streams.filter(s => s.is_live).length > 0 && (
          <div className="mb-8">
            <h2 className="text-xl font-bold text-gray-900 mb-4">البثوث النشطة</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {streams.filter(s => s.is_live).map((stream) => (
                <div key={stream.id} className="bg-white rounded-lg shadow p-4 border-r-4 border-green-500">
                  <div className="flex justify-between items-start">
                    <div>
                      <h3 className="font-semibold">{stream.title}</h3>
                      <p className="text-sm text-gray-500">معرف الغرفة: {stream.room_id}</p>
                    </div>
                    <span className="bg-green-100 text-green-800 text-xs px-2 py-1 rounded-full flex items-center gap-1">
                      <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
                      مباشر
                    </span>
                  </div>
                  <div className="mt-3 flex gap-2">
                    <button
                      onClick={() => copyStreamLink(stream.room_id)}
                      className="text-sm text-indigo-600 hover:text-indigo-700"
                    >
                      نسخ الرابط
                    </button>
                    <Link
                      href={`/stream/${stream.room_id}`}
                      className="text-sm text-green-600 hover:text-green-700"
                    >
                      الانضمام
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Courses List */}
        <div>
          <h2 className="text-xl font-bold text-gray-900 mb-4">دوراتي</h2>
          {courses.length === 0 ? (
            <div className="bg-white rounded-lg shadow p-8 text-center">
              <p className="text-gray-600 mb-4">لم تقم بإنشاء أي دورات بعد.</p>
              <Link
                href="/instructor/courses/new"
                className="inline-block bg-indigo-600 text-white px-6 py-3 rounded-md hover:bg-indigo-700"
              >
                إنشاء دورة جديدة
              </Link>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {courses.map((course) => (
                <div key={course.id} className="bg-white rounded-lg shadow p-6">
                  <div className="flex justify-between items-start mb-3">
                    <h3 className="text-lg font-semibold text-gray-900">{course.title}</h3>
                    <span className={`text-xs px-2 py-1 rounded-full ${
                      course.is_published 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-yellow-100 text-yellow-800'
                    }`}>
                      {course.is_published ? 'منشورة' : 'مسودة'}
                    </span>
                  </div>
                  <p className="text-sm text-gray-600 mb-4 line-clamp-2">{course.description}</p>
                  <div className="flex justify-between items-center text-sm">
                    <span className="text-gray-500">{course.student_count || 0} طالب</span>
                    <Link
                      href={`/instructor/courses/${course.id}`}
                      className="text-indigo-600 hover:text-indigo-700"
                    >
                      إدارة ←
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
