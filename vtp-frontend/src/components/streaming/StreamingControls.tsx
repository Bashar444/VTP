import { useState, useEffect } from 'react';
import { Mic, MicOff, Video, VideoOff, Share2, Settings, LogOut } from 'lucide-react';
import { cn } from '@/utils/cn';

interface StreamingControlsProps {
  isAudioEnabled: boolean;
  isVideoEnabled: boolean;
  recordingStatus?: 'idle' | 'recording';
  onToggleAudio: (enabled: boolean) => Promise<void>;
  onToggleVideo: (enabled: boolean) => Promise<void>;
  onToggleScreenShare?: (enabled: boolean) => Promise<void>;
  onToggleRecording?: () => Promise<void>;
  onSettings?: () => void;
  onLeave: () => void;
  isLoading?: boolean;
  className?: string;
}

export const StreamingControls: React.FC<StreamingControlsProps> = ({
  isAudioEnabled,
  isVideoEnabled,
  recordingStatus = 'idle',
  onToggleAudio,
  onToggleVideo,
  onToggleScreenShare,
  onToggleRecording,
  onSettings,
  onLeave,
  isLoading = false,
  className,
}) => {
  const [audioLoading, setAudioLoading] = useState(false);
  const [videoLoading, setVideoLoading] = useState(false);
  const [screenShareLoading, setScreenShareLoading] = useState(false);
  const [isScreenSharing, setIsScreenSharing] = useState(false);

  const handleToggleAudio = async () => {
    try {
      setAudioLoading(true);
      await onToggleAudio(!isAudioEnabled);
    } finally {
      setAudioLoading(false);
    }
  };

  const handleToggleVideo = async () => {
    try {
      setVideoLoading(true);
      await onToggleVideo(!isVideoEnabled);
    } finally {
      setVideoLoading(false);
    }
  };

  const handleToggleScreenShare = async () => {
    try {
      setScreenShareLoading(true);
      if (onToggleScreenShare) {
        await onToggleScreenShare(!isScreenSharing);
        setIsScreenSharing(!isScreenSharing);
      }
    } finally {
      setScreenShareLoading(false);
    }
  };

  return (
    <div
      className={cn(
        'flex items-center justify-center gap-3 p-4 bg-gray-900 rounded-lg',
        className
      )}
    >
      {/* Audio Toggle */}
      <button
        onClick={handleToggleAudio}
        disabled={audioLoading || isLoading}
        className={cn(
          'flex items-center justify-center w-12 h-12 rounded-full transition-colors',
          isAudioEnabled
            ? 'bg-blue-600 hover:bg-blue-700 text-white'
            : 'bg-red-600 hover:bg-red-700 text-white',
          (audioLoading || isLoading) && 'opacity-50 cursor-not-allowed'
        )}
        title={isAudioEnabled ? 'Mute' : 'Unmute'}
      >
        {audioLoading ? (
          <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
        ) : isAudioEnabled ? (
          <Mic size={20} />
        ) : (
          <MicOff size={20} />
        )}
      </button>

      {/* Video Toggle */}
      <button
        onClick={handleToggleVideo}
        disabled={videoLoading || isLoading}
        className={cn(
          'flex items-center justify-center w-12 h-12 rounded-full transition-colors',
          isVideoEnabled
            ? 'bg-blue-600 hover:bg-blue-700 text-white'
            : 'bg-red-600 hover:bg-red-700 text-white',
          (videoLoading || isLoading) && 'opacity-50 cursor-not-allowed'
        )}
        title={isVideoEnabled ? 'Stop camera' : 'Start camera'}
      >
        {videoLoading ? (
          <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
        ) : isVideoEnabled ? (
          <Video size={20} />
        ) : (
          <VideoOff size={20} />
        )}
      </button>

      {/* Screen Share */}
      {onToggleScreenShare && (
        <button
          onClick={handleToggleScreenShare}
          disabled={screenShareLoading || isLoading}
          className={cn(
            'flex items-center justify-center w-12 h-12 rounded-full transition-colors',
            isScreenSharing
              ? 'bg-green-600 hover:bg-green-700 text-white'
              : 'bg-gray-700 hover:bg-gray-600 text-white',
            (screenShareLoading || isLoading) && 'opacity-50 cursor-not-allowed'
          )}
          title={isScreenSharing ? 'Stop sharing' : 'Share screen'}
        >
          {screenShareLoading ? (
            <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
          ) : (
            <Share2 size={20} />
          )}
        </button>
      )}

      {/* Recording */}
      {onToggleRecording && (
        <button
          onClick={onToggleRecording}
          disabled={isLoading}
          className={cn(
            'flex items-center justify-center w-12 h-12 rounded-full transition-colors',
            recordingStatus === 'recording'
              ? 'bg-yellow-600 hover:bg-yellow-700 text-white'
              : 'bg-gray-700 hover:bg-gray-600 text-white'
          )}
          title={recordingStatus === 'recording' ? 'Stop recording' : 'Start recording'}
        >
          {recordingStatus === 'recording' ? 'REC' : 'REC'}
        </button>
      )}

      {/* Settings */}
      {onSettings && (
        <button
          onClick={onSettings}
          disabled={isLoading}
          className="flex items-center justify-center w-12 h-12 rounded-full bg-gray-700 hover:bg-gray-600 text-white transition-colors"
          title="Settings"
        >
          <Settings size={20} />
        </button>
      )}

      <div className="w-px h-8 bg-gray-700" />

      {/* Leave Call */}
      <button
        onClick={onLeave}
        disabled={isLoading}
        className="flex items-center justify-center w-12 h-12 rounded-full bg-red-600 hover:bg-red-700 text-white transition-colors"
        title="Leave call"
      >
        <LogOut size={20} />
      </button>
    </div>
  );
};
