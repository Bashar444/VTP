'use client';
export const dynamic = 'force-dynamic';

import { useEffect, useState, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth.store';
import { Mic, MicOff, Video, VideoOff, PhoneOff, MessageSquare, Users, Copy, Check } from 'lucide-react';

declare global {
  interface Window {
    JitsiMeetExternalAPI: any;
  }
}

export default function StreamingPage() {
  const params = useParams();
  const router = useRouter();
  const roomId = params?.roomId as string;
  const authStore = useAuthStore();
  const jitsiContainerRef = useRef<HTMLDivElement>(null);
  const jitsiApiRef = useRef<any>(null);
  
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isAudioMuted, setIsAudioMuted] = useState(false);
  const [isVideoMuted, setIsVideoMuted] = useState(false);
  const [participantCount, setParticipantCount] = useState(1);
  const [showChat, setShowChat] = useState(false);
  const [copied, setCopied] = useState(false);

  // Load Jitsi script
  useEffect(() => {
    if (!roomId) return;

    const script = document.createElement('script');
    script.src = 'https://meet.jit.si/external_api.js';
    script.async = true;
    script.onload = () => {
      initJitsi();
    };
    script.onerror = () => {
      setError('Failed to load video conferencing');
      setIsLoading(false);
    };
    document.body.appendChild(script);

    return () => {
      if (jitsiApiRef.current) {
        jitsiApiRef.current.dispose();
      }
      document.body.removeChild(script);
    };
  }, [roomId]);

  const initJitsi = () => {
    if (!jitsiContainerRef.current || !window.JitsiMeetExternalAPI) return;

    try {
      const domain = 'meet.jit.si';
      const options = {
        roomName: `VTP-${roomId}`,
        width: '100%',
        height: '100%',
        parentNode: jitsiContainerRef.current,
        userInfo: {
          displayName: authStore.user?.full_name || 'مستخدم',
          email: authStore.user?.email || '',
        },
        configOverwrite: {
          startWithAudioMuted: false,
          startWithVideoMuted: false,
          prejoinPageEnabled: false,
          disableDeepLinking: true,
          enableWelcomePage: false,
          enableClosePage: false,
          defaultLanguage: 'ar',
          toolbarButtons: [
            'microphone',
            'camera',
            'desktop',
            'fullscreen',
            'hangup',
            'chat',
            'raisehand',
            'tileview',
            'settings',
          ],
        },
        interfaceConfigOverwrite: {
          SHOW_JITSI_WATERMARK: false,
          SHOW_WATERMARK_FOR_GUESTS: false,
          SHOW_BRAND_WATERMARK: false,
          BRAND_WATERMARK_LINK: '',
          DEFAULT_BACKGROUND: '#1a1a2e',
          DISABLE_JOIN_LEAVE_NOTIFICATIONS: false,
          MOBILE_APP_PROMO: false,
          HIDE_INVITE_MORE_HEADER: true,
        },
      };

      jitsiApiRef.current = new window.JitsiMeetExternalAPI(domain, options);

      // Event listeners
      jitsiApiRef.current.addListener('videoConferenceJoined', () => {
        setIsLoading(false);
      });

      jitsiApiRef.current.addListener('videoConferenceLeft', () => {
        router.push('/instructor/streaming');
      });

      jitsiApiRef.current.addListener('participantJoined', () => {
        setParticipantCount(prev => prev + 1);
      });

      jitsiApiRef.current.addListener('participantLeft', () => {
        setParticipantCount(prev => Math.max(1, prev - 1));
      });

      jitsiApiRef.current.addListener('audioMuteStatusChanged', (data: any) => {
        setIsAudioMuted(data.muted);
      });

      jitsiApiRef.current.addListener('videoMuteStatusChanged', (data: any) => {
        setIsVideoMuted(data.muted);
      });

    } catch (err) {
      setError('فشل في تهيئة مكالمة الفيديو');
      setIsLoading(false);
    }
  };

  const toggleAudio = () => {
    if (jitsiApiRef.current) {
      jitsiApiRef.current.executeCommand('toggleAudio');
    }
  };

  const toggleVideo = () => {
    if (jitsiApiRef.current) {
      jitsiApiRef.current.executeCommand('toggleVideo');
    }
  };

  const hangUp = () => {
    if (jitsiApiRef.current) {
      jitsiApiRef.current.executeCommand('hangup');
    }
    router.push('/instructor/streaming');
  };

  const toggleChat = () => {
    if (jitsiApiRef.current) {
      jitsiApiRef.current.executeCommand('toggleChat');
    }
    setShowChat(!showChat);
  };

  const copyLink = () => {
    const link = `${window.location.origin}/stream/${roomId}`;
    navigator.clipboard.writeText(link);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (error) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center" dir="rtl">
        <div className="text-center text-white">
          <div className="text-6xl mb-4">❌</div>
          <h1 className="text-2xl font-bold mb-2">خطأ في البث</h1>
          <p className="text-gray-400 mb-4">{error}</p>
          <button
            onClick={() => router.push('/instructor/streaming')}
            className="bg-indigo-600 px-6 py-2 rounded-lg hover:bg-indigo-700"
          >
            العودة
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 flex flex-col" dir="rtl">
      {/* Header */}
      <div className="bg-gray-800 px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h1 className="text-white font-bold">بث مباشر</h1>
          <span className="bg-red-600 text-white text-xs px-2 py-1 rounded-full flex items-center gap-1">
            <span className="w-2 h-2 bg-white rounded-full animate-pulse"></span>
            مباشر
          </span>
          <span className="text-gray-400 text-sm flex items-center gap-1">
            <Users size={16} />
            {participantCount} مشارك
          </span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={copyLink}
            className="bg-gray-700 text-white px-3 py-2 rounded-lg hover:bg-gray-600 flex items-center gap-2 text-sm"
          >
            {copied ? <Check size={16} /> : <Copy size={16} />}
            {copied ? 'تم النسخ' : 'نسخ الرابط'}
          </button>
        </div>
      </div>

      {/* Jitsi Container */}
      <div className="flex-1 relative">
        {isLoading && (
          <div className="absolute inset-0 bg-gray-900 flex items-center justify-center z-10">
            <div className="text-center text-white">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white mx-auto mb-4"></div>
              <p>جاري تحميل البث...</p>
            </div>
          </div>
        )}
        <div ref={jitsiContainerRef} className="w-full h-full" style={{ minHeight: 'calc(100vh - 140px)' }} />
      </div>

      {/* Bottom Controls */}
      <div className="bg-gray-800 px-4 py-3 flex items-center justify-center gap-4">
        <button
          onClick={toggleAudio}
          className={`p-4 rounded-full ${isAudioMuted ? 'bg-red-600' : 'bg-gray-700'} hover:opacity-80`}
        >
          {isAudioMuted ? <MicOff className="text-white" /> : <Mic className="text-white" />}
        </button>
        <button
          onClick={toggleVideo}
          className={`p-4 rounded-full ${isVideoMuted ? 'bg-red-600' : 'bg-gray-700'} hover:opacity-80`}
        >
          {isVideoMuted ? <VideoOff className="text-white" /> : <Video className="text-white" />}
        </button>
        <button
          onClick={toggleChat}
          className={`p-4 rounded-full ${showChat ? 'bg-indigo-600' : 'bg-gray-700'} hover:opacity-80`}
        >
          <MessageSquare className="text-white" />
        </button>
        <button
          onClick={hangUp}
          className="p-4 rounded-full bg-red-600 hover:bg-red-700"
        >
          <PhoneOff className="text-white" />
        </button>
      </div>
    </div>
  );
}
