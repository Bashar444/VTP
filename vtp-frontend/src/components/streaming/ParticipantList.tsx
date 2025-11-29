import { Users, Loader } from 'lucide-react';
import { cn } from '@/utils/cn';
import { useTranslations } from 'next-intl';

interface Participant {
  id: string;
  name: string;
  role: 'instructor' | 'student';
  isAudioEnabled: boolean;
  isVideoEnabled: boolean;
  joinedAt: Date;
}

interface ParticipantListProps {
  participants: Participant[];
  currentUserId: string;
  isLoading?: boolean;
  className?: string;
}

export const ParticipantList: React.FC<ParticipantListProps> = ({
  participants,
  currentUserId,
  isLoading = false,
  className,
}) => {
  return (
    <div className={cn('w-full max-w-sm bg-gray-900 rounded-lg p-4', className)}>
      <div className="flex items-center gap-2 mb-4">
        <Users size={20} className="text-blue-400" />
        <h3 className="font-semibold text-white">
          Participants ({participants.length})
        </h3>
      </div>

      <div className="space-y-2 max-h-96 overflow-y-auto">
        {isLoading && (
          <div className="flex items-center justify-center py-8">
            <Loader className="animate-spin text-blue-400" />
          </div>
        )}

        {!isLoading && participants.length === 0 && (
          <div className="text-center text-gray-400 py-8">
            Waiting for participants...
          </div>
        )}

        {participants.map((participant) => (
          <div
            key={participant.id}
            className="flex items-center gap-3 p-3 bg-gray-800 rounded-lg hover:bg-gray-750 transition-colors"
          >
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <p className="text-white font-medium truncate">{participant.name}</p>
                {participant.id === currentUserId && (
                  <span className="px-2 py-0.5 bg-blue-600 text-white text-xs rounded">
                    You
                  </span>
                )}
              </div>
              <p className="text-gray-400 text-sm capitalize">{participant.role}</p>
            </div>

            <div className="flex items-center gap-1">
              {participant.isAudioEnabled ? (
                <div className="w-2 h-2 bg-green-500 rounded-full" title="Audio on" />
              ) : (
                <div className="w-2 h-2 bg-red-500 rounded-full" title="Audio off" />
              )}
              {participant.isVideoEnabled ? (
                <div className="w-2 h-2 bg-green-500 rounded-full" title="Video on" />
              ) : (
                <div className="w-2 h-2 bg-red-500 rounded-full" title="Video off" />
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// Streaming Status Component
interface StreamingStatusProps {
  duration: number;
  participantCount: number;
  recordingStatus: 'idle' | 'recording' | 'paused';
  bitrate?: number;
  fps?: number;
  resolution?: string;
  className?: string;
}

export const StreamingStatus: React.FC<StreamingStatusProps> = ({
  duration,
  participantCount,
  recordingStatus,
  bitrate,
  fps,
  resolution,
  className,
}) => {
  const t = useTranslations();

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const recordingText = t(`stream.participant.recordingStatus.${recordingStatus}`);

  return (
    <div className={cn('bg-gray-900 rounded-lg p-4 space-y-3', className)}>
      <div className="flex items-center justify-between">
        <span className="text-gray-400">{t('stream.participant.duration')}</span>
        <span className="text-white font-mono font-semibold">{formatDuration(duration)}</span>
      </div>

      <div className="flex items-center justify-between">
        <span className="text-gray-400">{t('stream.participant.count')}</span>
        <span className="text-white font-semibold">{participantCount}</span>
      </div>

      <div className="flex items-center justify-between">
        <span className="text-gray-400">{t('stream.participant.recording')}</span>
        <div className="flex items-center gap-2">
          <div
            className={cn('w-2 h-2 rounded-full', {
              'bg-red-500 animate-pulse': recordingStatus === 'recording',
              'bg-yellow-500': recordingStatus === 'paused',
              'bg-gray-500': recordingStatus === 'idle',
            })}
          />
          <span className="text-white capitalize">{recordingText}</span>
        </div>
      </div>

      {bitrate && (
        <div className="flex items-center justify-between">
          <span className="text-gray-400">{t('stream.participant.bitrate')}</span>
          <span className="text-white text-sm">{(bitrate / 1000000).toFixed(1)} Mbps</span>
        </div>
      )}

      {fps && (
        <div className="flex items-center justify-between">
          <span className="text-gray-400">{t('stream.participant.fps')}</span>
          <span className="text-white text-sm">{fps} fps</span>
        </div>
      )}

      {resolution && (
        <div className="flex items-center justify-between">
          <span className="text-gray-400">{t('stream.participant.resolution')}</span>
          <span className="text-white text-sm">{resolution}</span>
        </div>
      )}
    </div>
  );
};
