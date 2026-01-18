"use client";

import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect, useState, useCallback } from 'react';
import Link from 'next/link';

interface StreamSession {
  id: string;
  room_id: string;
  title: string;
  description: string;
  status: 'scheduled' | 'live' | 'ended' | 'cancelled';
  course_id?: string;
  course_title?: string;
  started_at?: string;
  ended_at?: string;
  is_recording: boolean;
  max_participants: number;
  participant_count?: number;
  stream_url?: string;
  created_at: string;
}

export default function InstructorStreamingPage() {
  const { user, token } = useAuth();
  const router = useRouter();
  const [sessions, setSessions] = useState<StreamSession[]>([]);
  const [liveSessions, setLiveSessions] = useState<StreamSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newStreamTitle, setNewStreamTitle] = useState('');
  const [newStreamDescription, setNewStreamDescription] = useState('');
  const [createdSession, setCreatedSession] = useState<StreamSession | null>(null);

  const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  const fetchSessions = useCallback(async () => {
    try {
      setLoading(true);
      
      // Fetch instructor's sessions
      const sessionsRes = await fetch(`${API_URL}/api/v1/streaming/sessions?instructor_id=me`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (sessionsRes.ok) {
        const data = await sessionsRes.json();
        setSessions(data.sessions || []);
        setLiveSessions((data.sessions || []).filter((s: StreamSession) => s.status === 'live'));
      }
    } catch (err) {
      console.error('Failed to fetch sessions:', err);
    } finally {
      setLoading(false);
    }
  }, [token, API_URL]);

  useEffect(() => {
    if (!user) {
      router.push('/login');
      return;
    }
    if (user.role !== 'teacher' && user.role !== 'instructor' && user.role !== 'admin') {
      router.push('/my-courses');
      return;
    }

    fetchSessions();
  }, [user, router, fetchSessions]);

  const createAndStartStream = async () => {
    if (!newStreamTitle.trim()) return;

    try {
      const res = await fetch(`${API_URL}/api/v1/streaming/sessions/start`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({
          title: newStreamTitle,
          description: newStreamDescription
        })
      });

      if (res.ok) {
        const data = await res.json();
        setCreatedSession(data);
        setNewStreamTitle('');
        setNewStreamDescription('');
        fetchSessions();
      } else {
        const error = await res.json();
        alert(`خطأ: ${error.error || 'فشل في إنشاء البث'}`);
      }
    } catch (err) {
      console.error('Failed to create stream:', err);
      alert('فشل في إنشاء البث');
    }
  };

  const stopStream = async (sessionId: string) => {
    try {
      const res = await fetch(`${API_URL}/api/v1/streaming/sessions/${sessionId}/stop`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      });

      if (res.ok) {
        fetchSessions();
      } else {
        const error = await res.json();
        alert(`خطأ: ${error.error || 'فشل في إيقاف البث'}`);
      }
    } catch (err) {
      console.error('Failed to stop stream:', err);
    }
  };

  const copyStreamLink = (roomId: string) => {
    const link = `${window.location.origin}/stream/${roomId}`;
    navigator.clipboard.writeText(link);
    alert('تم نسخ رابط البث!');
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('ar-SA');
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'live':
        return (
          <span className="bg-green-100 text-green-800 text-xs px-2 py-1 rounded-full flex items-center gap-1">
            <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
            مباشر
          </span>
        );
      case 'scheduled':
        return <span className="bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded-full">مجدول</span>;
      case 'ended':
        return <span className="bg-gray-100 text-gray-800 text-xs px-2 py-1 rounded-full">منتهي</span>;
      case 'cancelled':
        return <span className="bg-red-100 text-red-800 text-xs px-2 py-1 rounded-full">ملغي</span>;
      default:
        return null;
    }
  };

  if (!user || (user.role !== 'teacher' && user.role !== 'instructor' && user.role !== 'admin')) {
    return null;
  }

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
            <h1 className="text-3xl font-bold text-gray-900">البث المباشر</h1>
            <p className="text-gray-600 mt-1">إدارة جلسات البث المباشر الخاصة بك</p>
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="bg-green-600 text-white px-6 py-3 rounded-lg hover:bg-green-700 flex items-center gap-2 shadow-lg"
          >
            <span className="text-xl">🎥</span>
            <span>بث مباشر جديد</span>
          </button>
        </div>

        {/* Create Stream Modal */}
        {showCreateModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-lg">
              <h2 className="text-xl font-bold mb-4">إنشاء بث مباشر جديد</h2>
              
              {createdSession ? (
                <div className="space-y-4">
                  <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                    <div className="flex items-center gap-2 mb-2">
                      <span className="text-green-500 text-2xl">✓</span>
                      <p className="text-green-800 font-medium">تم إنشاء البث بنجاح!</p>
                    </div>
                    <p className="text-sm text-gray-600 mb-2">رابط البث للطلاب:</p>
                    <div className="bg-white border rounded p-3 text-sm break-all font-mono">
                      {window.location.origin}/stream/{createdSession.room_id}
                    </div>
                  </div>
                  
                  <div className="flex gap-3">
                    <button
                      onClick={() => copyStreamLink(createdSession.room_id)}
                      className="flex-1 bg-indigo-600 text-white py-2 rounded-md hover:bg-indigo-700"
                    >
                      📋 نسخ الرابط
                    </button>
                    <Link
                      href={`/stream/${createdSession.room_id}`}
                      className="flex-1 bg-green-600 text-white py-2 rounded-md hover:bg-green-700 text-center"
                    >
                      🎬 بدء البث
                    </Link>
                  </div>
                  
                  <button
                    onClick={() => {
                      setShowCreateModal(false);
                      setCreatedSession(null);
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
                      عنوان البث <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      value={newStreamTitle}
                      onChange={(e) => setNewStreamTitle(e.target.value)}
                      placeholder="مثال: محاضرة الرياضيات - الفصل الأول"
                      className="w-full border rounded-md px-3 py-2 focus:ring-2 focus:ring-green-500 focus:border-transparent"
                    />
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      وصف البث (اختياري)
                    </label>
                    <textarea
                      value={newStreamDescription}
                      onChange={(e) => setNewStreamDescription(e.target.value)}
                      placeholder="وصف قصير عن محتوى البث..."
                      rows={3}
                      className="w-full border rounded-md px-3 py-2 focus:ring-2 focus:ring-green-500 focus:border-transparent"
                    />
                  </div>
                  
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 text-sm text-blue-800">
                    <p>💡 سيتم بدء البث فور الإنشاء. يمكنك مشاركة الرابط مع طلابك.</p>
                  </div>
                  
                  <div className="flex gap-3">
                    <button
                      onClick={createAndStartStream}
                      disabled={!newStreamTitle.trim()}
                      className="flex-1 bg-green-600 text-white py-2 rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      🎬 إنشاء وبدء البث
                    </button>
                    <button
                      onClick={() => setShowCreateModal(false)}
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

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center text-2xl">
                🔴
              </div>
              <div>
                <div className="text-3xl font-bold text-green-600">{liveSessions.length}</div>
                <div className="text-gray-600">بث مباشر الآن</div>
              </div>
            </div>
          </div>
          
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center text-2xl">
                📺
              </div>
              <div>
                <div className="text-3xl font-bold text-blue-600">{sessions.length}</div>
                <div className="text-gray-600">إجمالي الجلسات</div>
              </div>
            </div>
          </div>
          
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center text-2xl">
                ⏱️
              </div>
              <div>
                <div className="text-3xl font-bold text-purple-600">
                  {sessions.filter(s => s.status === 'ended').length}
                </div>
                <div className="text-gray-600">بثوث منتهية</div>
              </div>
            </div>
          </div>
        </div>

        {/* Live Sessions Section */}
        {liveSessions.length > 0 && (
          <div className="mb-8">
            <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
              <span className="w-3 h-3 bg-red-500 rounded-full animate-pulse"></span>
              البثوث المباشرة الآن
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {liveSessions.map((session) => (
                <div key={session.id} className="bg-white rounded-lg shadow-lg p-6 border-r-4 border-green-500">
                  <div className="flex justify-between items-start mb-4">
                    <div>
                      <h3 className="text-lg font-semibold text-gray-900">{session.title}</h3>
                      {session.description && (
                        <p className="text-sm text-gray-500 mt-1">{session.description}</p>
                      )}
                    </div>
                    {getStatusBadge(session.status)}
                  </div>
                  
                  <div className="text-sm text-gray-500 mb-4">
                    <p>معرف الغرفة: <span className="font-mono">{session.room_id}</span></p>
                    {session.started_at && <p>بدأ في: {formatDate(session.started_at)}</p>}
                  </div>
                  
                  <div className="flex gap-2 flex-wrap">
                    <Link
                      href={`/stream/${session.room_id}`}
                      className="bg-green-600 text-white px-4 py-2 rounded-md hover:bg-green-700 text-sm"
                    >
                      🎬 الانضمام للبث
                    </Link>
                    <button
                      onClick={() => copyStreamLink(session.room_id)}
                      className="border border-indigo-600 text-indigo-600 px-4 py-2 rounded-md hover:bg-indigo-50 text-sm"
                    >
                      📋 نسخ الرابط
                    </button>
                    <button
                      onClick={() => {
                        if (confirm('هل أنت متأكد من إيقاف البث؟')) {
                          stopStream(session.id);
                        }
                      }}
                      className="border border-red-600 text-red-600 px-4 py-2 rounded-md hover:bg-red-50 text-sm"
                    >
                      ⏹️ إيقاف البث
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* All Sessions */}
        <div>
          <h2 className="text-xl font-bold text-gray-900 mb-4">سجل الجلسات</h2>
          
          {sessions.length === 0 ? (
            <div className="bg-white rounded-lg shadow p-8 text-center">
              <div className="text-6xl mb-4">🎥</div>
              <p className="text-gray-600 mb-4">لم تقم بإنشاء أي بثوث بعد.</p>
              <button
                onClick={() => setShowCreateModal(true)}
                className="inline-block bg-green-600 text-white px-6 py-3 rounded-md hover:bg-green-700"
              >
                إنشاء أول بث مباشر
              </button>
            </div>
          ) : (
            <div className="bg-white rounded-lg shadow overflow-hidden">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      العنوان
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      الحالة
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      تاريخ الإنشاء
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      الإجراءات
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {sessions.map((session) => (
                    <tr key={session.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">{session.title}</div>
                        <div className="text-sm text-gray-500 font-mono">{session.room_id}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getStatusBadge(session.status)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {formatDate(session.created_at)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        <div className="flex gap-2">
                          {session.status === 'live' && (
                            <Link
                              href={`/stream/${session.room_id}`}
                              className="text-green-600 hover:text-green-700"
                            >
                              الانضمام
                            </Link>
                          )}
                          <button
                            onClick={() => copyStreamLink(session.room_id)}
                            className="text-indigo-600 hover:text-indigo-700"
                          >
                            نسخ الرابط
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
