'use client';
export const dynamic = 'force-dynamic';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth.store';
import { Copy, Check, ArrowRight } from 'lucide-react';

export default function StreamingPage() {
  const params = useParams();
  const router = useRouter();
  const roomId = params?.roomId as string;
  const authStore = useAuthStore();
  
  const [copied, setCopied] = useState(false);
  
  // Generate Jitsi room URL with config
  const jitsiRoomName = `VTP-${roomId}`;
  const displayName = encodeURIComponent(authStore.user?.full_name || 'User');
  
  // Build Jitsi URL with parameters - using simpler URL format
  const jitsiUrl = `https://meet.jit.si/${jitsiRoomName}#config.prejoinPageEnabled=false&userInfo.displayName="${displayName}"`;

  const copyLink = () => {
    const link = `${window.location.origin}/stream/${roomId}`;
    navigator.clipboard.writeText(link);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const goBack = () => {
    router.push('/instructor/streaming');
  };

  return (
    <div className="min-h-screen bg-gray-900 flex flex-col" dir="rtl">
      {/* Header */}
      <div className="bg-gray-800 px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={goBack}
            className="text-gray-400 hover:text-white p-2"
          >
            <ArrowRight size={20} />
          </button>
          <h1 className="text-white font-bold">بث مباشر</h1>
          <span className="bg-red-600 text-white text-xs px-2 py-1 rounded-full flex items-center gap-1">
            <span className="w-2 h-2 bg-white rounded-full animate-pulse"></span>
            مباشر
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

      {/* Jitsi iframe - Direct embed */}
      <div className="flex-1">
        <iframe
          src={jitsiUrl}
          style={{ width: '100%', height: 'calc(100vh - 60px)', border: 'none' }}
          allow="camera; microphone; fullscreen; display-capture; autoplay; clipboard-write"
          allowFullScreen
        />
      </div>
    </div>
  );
}
