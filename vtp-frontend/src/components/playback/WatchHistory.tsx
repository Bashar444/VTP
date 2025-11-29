import { useState, useEffect } from 'react';
import { Clock, Trash2, Play } from 'lucide-react';
import VideoService, { PlaybackHistory } from '@/services/video.service';
import { cn } from '@/utils/cn';

interface WatchHistoryProps {
  userId: string;
  limit?: number;
  onVideoSelect?: (videoId: string) => void;
  className?: string;
}

export const WatchHistory: React.FC<WatchHistoryProps> = ({
  userId,
  limit = 10,
  onVideoSelect,
  className,
}) => {
  const [history, setHistory] = useState<PlaybackHistory[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchHistory = async () => {
      try {
        setIsLoading(true);
        const data = await VideoService.getWatchHistory(userId, limit);
        setHistory(data);
        setError(null);
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Failed to load watch history';
        setError(message);
      } finally {
        setIsLoading(false);
      }
    };

    fetchHistory();
  }, [userId, limit]);

  const formatDate = (date: Date) => {
    return new Date(date).toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const formatDuration = (seconds: number) => {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);
    return h > 0
      ? `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
      : `${m}:${s.toString().padStart(2, '0')}`;
  };

  if (isLoading) {
    return (
      <div className={cn('space-y-4', className)}>
        {[...Array(3)].map((_, i) => (
          <div key={i} className="bg-gray-800 rounded-lg p-4 animate-pulse">
            <div className="h-4 bg-gray-700 rounded w-2/3 mb-2" />
            <div className="h-3 bg-gray-700 rounded w-1/2" />
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className={cn('bg-red-900 border border-red-700 rounded-lg p-4 text-red-200', className)}>
        <p className="font-semibold mb-1">Error Loading History</p>
        <p className="text-sm">{error}</p>
      </div>
    );
  }

  if (history.length === 0) {
    return (
      <div className={cn('bg-gray-800 rounded-lg p-8 text-center text-gray-400', className)}>
        <Clock className="w-8 h-8 mx-auto mb-2 opacity-50" />
        <p>No watch history yet</p>
        <p className="text-sm mt-1">Videos you watch will appear here</p>
      </div>
    );
  }

  return (
    <div className={cn('space-y-2', className)}>
      <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
        <Clock className="w-5 h-5" />
        Watch History
      </h3>

      {history.map(item => (
        <div
          key={item.id}
          className="bg-gray-800 rounded-lg p-4 hover:bg-gray-750 transition-colors group"
        >
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <h4 className="font-semibold text-white truncate mb-2">Video {item.videoId}</h4>
              <div className="space-y-1 text-sm text-gray-400">
                <p>Watched: {formatDate(item.watchedAt)}</p>
                <div className="flex items-center gap-2">
                  <div className="flex-1 bg-gray-700 rounded-full h-1.5">
                    <div
                      className="bg-blue-500 h-full rounded-full transition-all"
                      style={{ width: `${item.completedPercentage}%` }}
                    />
                  </div>
                  <span className="text-xs font-medium">
                    {Math.round(item.completedPercentage)}%
                  </span>
                </div>
                <p>
                  {formatDuration(item.watchedDuration)} / {formatDuration(item.totalDuration)}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              {onVideoSelect && (
                <button
                  onClick={() => onVideoSelect(item.videoId)}
                  className="p-2 text-gray-400 hover:text-white bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors opacity-0 group-hover:opacity-100"
                  title="Resume watching"
                >
                  <Play className="w-4 h-4" fill="currentColor" />
                </button>
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};

interface RecommendedVideosProps {
  limit?: number;
  onVideoSelect?: (videoId: string) => void;
  className?: string;
}

export const RecommendedVideos: React.FC<RecommendedVideosProps> = ({
  limit = 5,
  onVideoSelect,
  className,
}) => {
  const [videos, setVideos] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchRecommended = async () => {
      try {
        setIsLoading(true);
        const data = await VideoService.getRecommendedVideos(limit);
        setVideos(data);
        setError(null);
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Failed to load recommendations';
        setError(message);
      } finally {
        setIsLoading(false);
      }
    };

    fetchRecommended();
  }, [limit]);

  if (isLoading) {
    return (
      <div className={cn('grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4', className)}>
        {[...Array(3)].map((_, i) => (
          <div key={i} className="bg-gray-800 rounded-lg overflow-hidden animate-pulse">
            <div className="w-full aspect-video bg-gray-700" />
            <div className="p-3 space-y-2">
              <div className="h-4 bg-gray-700 rounded w-2/3" />
              <div className="h-3 bg-gray-700 rounded w-1/2" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className={cn('bg-red-900 border border-red-700 rounded-lg p-4 text-red-200', className)}>
        <p className="font-semibold mb-1">Error Loading Recommendations</p>
        <p className="text-sm">{error}</p>
      </div>
    );
  }

  return (
    <div className={cn('', className)}>
      <h3 className="text-lg font-semibold text-white mb-4">Recommended For You</h3>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {videos.map(video => (
          <div
            key={video.id}
            onClick={() => onVideoSelect?.(video.id)}
            className="bg-gray-800 rounded-lg overflow-hidden hover:bg-gray-750 transition-colors cursor-pointer group"
          >
            <div className="relative w-full aspect-video bg-gray-700 overflow-hidden">
              <img
                src={video.thumbnail || '/placeholder-video.jpg'}
                alt={video.title}
                className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-300"
              />
              <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-30 transition-colors" />
            </div>
            <div className="p-3">
              <h4 className="font-semibold text-white line-clamp-2 mb-1">{video.title}</h4>
              <p className="text-sm text-gray-400 line-clamp-2">{video.description}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
