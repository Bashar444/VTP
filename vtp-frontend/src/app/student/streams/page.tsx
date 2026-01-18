"use client";

import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect, useState, useCallback } from 'react';
import Link from 'next/link';

interface LiveStream {
  id: string;
  room_id: string;
  title: string;
  description: string;
  instructor_id: string;
  instructor_name?: string;
  course_id?: string;
  course_title?: string;
  status: string;
  started_at?: string;
  participant_count?: number;
  stream_url?: string;
}

export default function StudentStreamsPage() {
  const { user, token } = useAuth();
  const router = useRouter();
  const [liveStreams, setLiveStreams] = useState<LiveStream[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  const fetchLiveStreams = useCallback(async () => {
    try {
      setLoading(true);
      
      // Fetch live sessions
      const res = await fetch(`${API_URL}/api/v1/streaming/sessions/live`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      
      if (res.ok) {
        const data = await res.json();
        setLiveStreams(data.sessions || []);
      }
    } catch (err) {
      console.error('Failed to fetch live streams:', err);
    } finally {
      setLoading(false);
    }
  }, [token, API_URL]);

  useEffect(() => {
    if (!user) {
      router.push('/login');
      return;
    }

    fetchLiveStreams();
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(fetchLiveStreams, 30000);
    return () => clearInterval(interval);
  }, [user, router, fetchLiveStreams]);

  const filteredStreams = liveStreams.filter(stream => 
    stream.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (stream.instructor_name && stream.instructor_name.toLowerCase().includes(searchQuery.toLowerCase())) ||
    (stream.course_title && stream.course_title.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    
    if (diffMins < 1) return 'الآن';
    if (diffMins < 60) return `منذ ${diffMins} دقيقة`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `منذ ${diffHours} ساعة`;
    return date.toLocaleDateString('ar-SA');
  };

  if (!user) return null;

  if (loading) {
    return (
      <div dir="rtl" className="min-h-screen bg-gray-50 pt-24 pb-12">
        <div className="max-w-7xl mx-auto px-4">
          <div className="text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">جاري البحث عن البثوث المباشرة...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div dir="rtl" className="min-h-screen bg-gray-50 pt-24 pb-12">
      <div className="max-w-7xl mx-auto px-4">
        {/* Header */}
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
              <span className="w-4 h-4 bg-red-500 rounded-full animate-pulse"></span>
              البثوث المباشرة
            </h1>
            <p className="text-gray-600 mt-1">شاهد دروسك المباشرة الآن</p>
          </div>
          
          <div className="flex items-center gap-4">
            {/* Search */}
            <div className="relative">
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="ابحث عن بث..."
                className="w-64 border rounded-lg px-4 py-2 pr-10 focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              />
              <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400">
                🔍
              </span>
            </div>
            
            {/* Refresh button */}
            <button
              onClick={fetchLiveStreams}
              className="p-2 border rounded-lg hover:bg-gray-100 transition-colors"
              title="تحديث"
            >
              🔄
            </button>
          </div>
        </div>

        {/* Live Now Banner */}
        {filteredStreams.length > 0 && (
          <div className="bg-gradient-to-r from-red-500 to-pink-500 rounded-lg p-4 mb-8 text-white">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-white bg-opacity-20 rounded-full flex items-center justify-center">
                <span className="text-2xl">📺</span>
              </div>
              <div>
                <p className="font-bold text-lg">{filteredStreams.length} بث مباشر متاح الآن!</p>
                <p className="text-sm opacity-90">انضم إلى معلميك وتابع دروسك في الوقت الفعلي</p>
              </div>
            </div>
          </div>
        )}

        {/* Streams Grid */}
        {filteredStreams.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-12 text-center">
            <div className="text-6xl mb-4">📺</div>
            <h3 className="text-xl font-semibold text-gray-900 mb-2">لا توجد بثوث مباشرة حالياً</h3>
            <p className="text-gray-600 mb-4">
              {searchQuery 
                ? 'لم يتم العثور على بثوث تطابق بحثك.' 
                : 'سيظهر هنا البثوث المباشرة عندما يبدأ معلموك بالبث.'
              }
            </p>
            <div className="flex justify-center gap-4">
              <Link
                href="/student/dashboard"
                className="text-indigo-600 hover:text-indigo-700"
              >
                ← العودة للوحة التحكم
              </Link>
              {searchQuery && (
                <button
                  onClick={() => setSearchQuery('')}
                  className="text-gray-600 hover:text-gray-700"
                >
                  مسح البحث
                </button>
              )}
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredStreams.map((stream) => (
              <div
                key={stream.id}
                className="bg-white rounded-lg shadow-lg overflow-hidden hover:shadow-xl transition-shadow"
              >
                {/* Stream Preview/Thumbnail */}
                <div className="relative bg-gradient-to-br from-indigo-500 to-purple-600 h-40">
                  <div className="absolute inset-0 flex items-center justify-center">
                    <span className="text-6xl opacity-50">🎬</span>
                  </div>
                  
                  {/* Live Badge */}
                  <div className="absolute top-3 right-3">
                    <span className="bg-red-500 text-white text-xs px-2 py-1 rounded-full flex items-center gap-1">
                      <span className="w-2 h-2 bg-white rounded-full animate-pulse"></span>
                      مباشر
                    </span>
                  </div>
                  
                  {/* Participant Count */}
                  {stream.participant_count !== undefined && (
                    <div className="absolute bottom-3 left-3">
                      <span className="bg-black bg-opacity-50 text-white text-xs px-2 py-1 rounded">
                        👥 {stream.participant_count} مشاهد
                      </span>
                    </div>
                  )}
                </div>
                
                {/* Stream Info */}
                <div className="p-4">
                  <h3 className="font-semibold text-gray-900 text-lg mb-1 line-clamp-1">
                    {stream.title}
                  </h3>
                  
                  {stream.description && (
                    <p className="text-sm text-gray-500 mb-3 line-clamp-2">
                      {stream.description}
                    </p>
                  )}
                  
                  <div className="space-y-1 text-sm text-gray-500 mb-4">
                    {stream.instructor_name && (
                      <p>👨‍🏫 {stream.instructor_name}</p>
                    )}
                    {stream.course_title && (
                      <p>📚 {stream.course_title}</p>
                    )}
                    {stream.started_at && (
                      <p>⏱️ بدأ {formatTime(stream.started_at)}</p>
                    )}
                  </div>
                  
                  <Link
                    href={`/stream/${stream.room_id}`}
                    className="block w-full bg-green-600 text-white text-center py-3 rounded-lg hover:bg-green-700 transition-colors font-medium"
                  >
                    🎬 انضم الآن
                  </Link>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Help Section */}
        <div className="mt-12 bg-blue-50 border border-blue-200 rounded-lg p-6">
          <h3 className="font-semibold text-blue-900 mb-2">💡 نصائح للمشاهدة</h3>
          <ul className="text-sm text-blue-800 space-y-1">
            <li>• تأكد من اتصالك بالإنترنت قبل الانضمام</li>
            <li>• استخدم سماعات للحصول على أفضل جودة صوت</li>
            <li>• يمكنك طرح الأسئلة خلال البث عبر المحادثة</li>
            <li>• إذا واجهت مشاكل تقنية، حاول تحديث الصفحة</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
