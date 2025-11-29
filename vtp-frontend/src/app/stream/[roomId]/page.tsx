'use client';

import { useEffect, useState, useCallback } from 'react';
import { useStreamRoom } from '@/hooks/useStreamRoom';
import { useParams, useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth.store';
import { useMediasoup } from '@/hooks/useMediasoup';
import { VideoGrid } from '@/components/streaming/VideoGrid';
import { StreamingControls } from '@/components/streaming/StreamingControls';
import { ParticipantList, StreamingStatus } from '@/components/streaming/ParticipantList';
import { ChatPanel } from '@/components/streaming/ChatPanel';
import { AlertCircle, Loader } from 'lucide-react';

interface Participant {
  id: string;
  name: string;
  role: 'instructor' | 'student';
  isAudioEnabled: boolean;
  isVideoEnabled: boolean;
  joinedAt: Date;
}

export default function StreamingPage() {
  const params = useParams();
  const router = useRouter();
  const roomId = params?.roomId as string;
  const authStore = useAuthStore();

  const {
    localStream,
    remoteStreams,
    error: mediasoupError,
    isConnected,
    getLocalStream,
    toggleAudio,
    toggleVideo,
    consumeRemoteStream,
    disconnect,
  } = useMediasoup(roomId);

  const [isAudioEnabled, setIsAudioEnabled] = useState(true);
  const [isVideoEnabled, setIsVideoEnabled] = useState(true);
  const [isInitializing, setIsInitializing] = useState(true);
  const [participants, setParticipants] = useState<Participant[]>([]);
  const { peers, connected: signalingConnected, getParticipants } = useStreamRoom(roomId);
  const [streamDuration, setStreamDuration] = useState(0);
  const [recordingStatus, setRecordingStatus] = useState<'idle' | 'recording' | 'paused'>(
    'idle'
  );
  const [error, setError] = useState<string | null>(null);

  // Initialize local stream
  const initializeLocalStream = useCallback(async () => {
    try {
      if (isConnected) {
        const stream = await getLocalStream(isAudioEnabled, isVideoEnabled);
        setIsInitializing(false);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to initialize stream';
      setError(message);
      setIsInitializing(false);
    }
  }, [isConnected, getLocalStream, isAudioEnabled, isVideoEnabled]);

  useEffect(() => {
    if (isConnected && isInitializing) {
      initializeLocalStream();
    }
  }, [isConnected, isInitializing, initializeLocalStream]);

  // Handle audio toggle
  const handleToggleAudio = useCallback(async (enabled: boolean) => {
    try {
      await toggleAudio(enabled);
      setIsAudioEnabled(enabled);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to toggle audio';
      setError(message);
    }
  }, [toggleAudio]);

  // Handle video toggle
  const handleToggleVideo = useCallback(async (enabled: boolean) => {
    try {
      await toggleVideo(enabled);
      setIsVideoEnabled(enabled);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to toggle video';
      setError(message);
    }
  }, [toggleVideo]);

  // Handle leave call
  const handleLeaveCall = useCallback(async () => {
    try {
      await disconnect();
      router.push('/courses');
    } catch (err) {
      console.error('Error leaving call:', err);
      router.push('/courses');
    }
  }, [disconnect, router]);

  // Update stream duration
  useEffect(() => {
    const interval = setInterval(() => {
      setStreamDuration(prev => prev + 1);
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  // Sync participants from signaling
  useEffect(() => {
    if (signalingConnected) {
      // Initial fetch
      getParticipants();
    }
  }, [signalingConnected, getParticipants]);

  useEffect(() => {
    // Map peers into Participant type for UI
    const mapped: Participant[] = peers.map(p => ({
      id: p.id,
      name: p.name || 'Guest',
      role: (p as any).role === 'instructor' ? 'instructor' : 'student',
      isAudioEnabled: true,
      isVideoEnabled: true,
      joinedAt: p.joinedAt,
    }));
    setParticipants(mapped);
  }, [peers]);

  // Check authentication
  if (!authStore.isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-center">
          <Loader className="w-8 h-8 animate-spin text-blue-400 mx-auto mb-4" />
          <p className="text-gray-400">Redirecting to login...</p>
        </div>
      </div>
    );
  }

  // Show error state
  if (error || mediasoupError) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="max-w-md w-full bg-gray-800 rounded-lg p-6">
          <div className="flex items-center gap-3 mb-4">
            <AlertCircle className="w-6 h-6 text-red-500" />
            <h2 className="text-xl font-semibold text-white">Connection Error</h2>
          </div>
          <p className="text-gray-400 mb-6">{error || mediasoupError}</p>
          <button
            onClick={() => router.push('/courses')}
            className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
          >
            Go Back to Courses
          </button>
        </div>
      </div>
    );
  }

  // Show loading state
  if (isInitializing) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-center">
          <Loader className="w-8 h-8 animate-spin text-blue-400 mx-auto mb-4" />
          <p className="text-gray-400">Initializing streaming...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 p-4">
      {/* Header */}
      <div className="max-w-7xl mx-auto mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2">Live Lecture</h1>
            <p className="text-gray-400">Room ID: {roomId}</p>
          </div>
          <div className="text-right">
            <p className="text-gray-400">Instructor: {authStore.user?.firstName || 'Unknown'}</p>
            <p className="text-gray-400">
              Connection: <span className="text-green-400">{isConnected ? 'Media' : 'Media…'} / {signalingConnected ? 'Signal' : 'Signal…'}</span>
            </p>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Video Grid + Controls - 3 columns */}
        <div className="lg:col-span-3 space-y-4">
          {/* Video Grid */}
          <div className="bg-gray-800 rounded-lg overflow-hidden">
            <VideoGrid
              localStream={localStream}
              remoteStreams={remoteStreams}
              localUserId={authStore.user?.id || ''}
              participantCount={participants.length + 1}
            />
          </div>

          {/* Controls */}
          <StreamingControls
            isAudioEnabled={isAudioEnabled}
            isVideoEnabled={isVideoEnabled}
            onToggleAudio={handleToggleAudio}
            onToggleVideo={handleToggleVideo}
            onLeave={handleLeaveCall}
            isLoading={isInitializing}
          />
        </div>

        {/* Sidebar - participants, status, chat */}
        <div className="space-y-4">
          {/* Participants List */}
          <ParticipantList
            participants={participants}
            currentUserId={authStore.user?.id || ''}
          />

          {/* Streaming Status */}
          <StreamingStatus
            duration={streamDuration}
            participantCount={participants.length + 1}
            recordingStatus={recordingStatus}
            resolution="1280x720"
            fps={30}
          />
          <ChatPanel signaling={(signalingRef as any)?.current} roomId={roomId} />
        </div>
      </div>

      {/* Error notification */}
      {(error || mediasoupError) && (
        <div className="fixed bottom-4 right-4 max-w-sm bg-red-900 border border-red-600 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
            <div>
              <h3 className="font-semibold text-white mb-1">Error</h3>
              <p className="text-red-200 text-sm">{error || mediasoupError}</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
