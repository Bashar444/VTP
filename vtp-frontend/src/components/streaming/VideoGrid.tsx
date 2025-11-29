import { useEffect, useRef } from 'react';
import { cn } from '@/utils/cn';

interface VideoGridProps {
  localStream: MediaStream | null;
  remoteStreams: Map<string, MediaStream>;
  localUserId: string;
  participantCount: number;
  className?: string;
}

export const VideoGrid: React.FC<VideoGridProps> = ({
  localStream,
  remoteStreams,
  localUserId,
  participantCount,
  className,
}) => {
  const localVideoRef = useRef<HTMLVideoElement>(null);
  const remoteVideoRefs = useRef<Map<string, HTMLVideoElement>>(new Map());

  useEffect(() => {
    if (localVideoRef.current && localStream) {
      localVideoRef.current.srcObject = localStream;
    }
  }, [localStream]);

  useEffect(() => {
    remoteStreams.forEach((stream, peerId) => {
      const videoElement = remoteVideoRefs.current.get(peerId);
      if (videoElement) {
        videoElement.srcObject = stream;
      }
    });
  }, [remoteStreams]);

  const totalParticipants = remoteStreams.size + (localStream ? 1 : 0);
  const gridCols = totalParticipants <= 2 ? 1 : totalParticipants <= 4 ? 2 : 3;

  return (
    <div
      className={cn(
        'w-full bg-gray-900 rounded-lg overflow-hidden',
        `grid grid-cols-${gridCols} gap-2 p-2`,
        className
      )}
      style={{
        display: 'grid',
        gridTemplateColumns: `repeat(${gridCols}, 1fr)`,
        gap: '0.5rem',
        padding: '0.5rem',
      }}
    >
      {/* Local Video */}
      {localStream && (
        <div className="relative bg-black rounded-lg overflow-hidden aspect-video">
          <video
            ref={localVideoRef}
            autoPlay
            muted
            playsInline
            className="w-full h-full object-cover"
          />
          <div className="absolute bottom-2 left-2 bg-black bg-opacity-60 text-white px-2 py-1 rounded text-sm">
            You (Student)
          </div>
        </div>
      )}

      {/* Remote Videos */}
      {Array.from(remoteStreams.entries()).map(([peerId, stream]) => (
        <div key={peerId} className="relative bg-black rounded-lg overflow-hidden aspect-video">
          <video
            ref={(el) => {
              if (el && !remoteVideoRefs.current.has(peerId)) {
                remoteVideoRefs.current.set(peerId, el);
              }
            }}
            autoPlay
            playsInline
            className="w-full h-full object-cover"
          />
          <div className="absolute bottom-2 left-2 bg-black bg-opacity-60 text-white px-2 py-1 rounded text-sm">
            Participant {Array.from(remoteStreams.keys()).indexOf(peerId) + 1}
          </div>
        </div>
      ))}

      {/* Empty Placeholder when no streams */}
      {totalParticipants === 0 && (
        <div className="col-span-full flex items-center justify-center h-96 bg-gray-800 rounded-lg">
          <div className="text-center text-gray-400">
            <div className="text-lg mb-2">Waiting for participants...</div>
            <div className="text-sm">Total participants: {participantCount}</div>
          </div>
        </div>
      )}
    </div>
  );
};
