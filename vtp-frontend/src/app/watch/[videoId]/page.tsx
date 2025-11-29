"use client";
export const dynamic = 'force-dynamic';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth.store';
import { VideoPlayer } from '@/components/playback/VideoPlayer';
import { PlaybackSettings, QualitySelector, SubtitleSelector } from '@/components/playback/PlaybackSettings';
import { WatchHistory, RecommendedVideos } from '@/components/playback/WatchHistory';
import VideoService, { VideoMetadata } from '@/services/video.service';
import { AlertCircle, Loader, Share2, Flag } from 'lucide-react';

const PLAYBACK_SPEEDS = [0.5, 0.75, 1, 1.25, 1.5, 1.75, 2];

export default function VideoPlaybackPage() {
  const params = useParams();
  const router = useRouter();
  const videoId = params?.videoId as string;
  const authStore = useAuthStore();

  const [video, setVideo] = useState<VideoMetadata | null>(null);
  const [subtitles, setSubtitles] = useState<Array<{ language: string; url: string }>>([]);
  const [currentSpeed, setCurrentSpeed] = useState(1);
  const [currentSubtitle, setCurrentSubtitle] = useState<string>('');
  const [showSettings, setShowSettings] = useState(false);
  const [showQuality, setShowQuality] = useState(false);
  const [showSubtitles, setShowSubtitles] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [watchProgress, setWatchProgress] = useState({ current: 0, total: 0 });

  // Fetch video metadata
  useEffect(() => {
    const fetchVideo = async () => {
      try {
        setIsLoading(true);
        const [videoData, subtitleData] = await Promise.all([
          VideoService.getVideoMetadata(videoId),
          VideoService.getSubtitles(videoId),
        ]);
        // Sanitize video metadata and subtitles to ensure plain objects
        const plainVideo: VideoMetadata = JSON.parse(JSON.stringify(videoData)) as VideoMetadata;
        const plainSubtitles: Array<{ language: string; url: string }> = subtitleData.map(s => JSON.parse(JSON.stringify(s)) as { language: string; url: string });
        setVideo(plainVideo);
        setSubtitles(plainSubtitles);
        setError(null);
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Failed to load video';
        setError(message);
      } finally {
        setIsLoading(false);
      }
    };

    if (videoId) {
      fetchVideo();
    }
  }, [videoId]);

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
  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="max-w-md w-full bg-gray-800 rounded-lg p-6">
          <div className="flex items-center gap-3 mb-4">
            <AlertCircle className="w-6 h-6 text-red-500" />
            <h2 className="text-xl font-semibold text-white">Error Loading Video</h2>
          </div>
          <p className="text-gray-400 mb-6">{error}</p>
          <button
            onClick={() => router.push('/courses')}
            className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
          >
            Back to Courses
          </button>
        </div>
      </div>
    );
  }

  // Show loading state
  if (isLoading || !video) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-900">
        <div className="text-center">
          <Loader className="w-8 h-8 animate-spin text-blue-400 mx-auto mb-4" />
          <p className="text-gray-400">Loading video...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 p-4">
      <div className="max-w-6xl mx-auto">
        {/* Video Player */}
        <VideoPlayer
          videoUrl={video.hlsUrl}
          videoId={video.id}
          title={video.title}
          thumbnail={video.thumbnail}
          onProgress={(current, total) => {
            setWatchProgress({ current, total });
          }}
          className="mb-6"
        />

        {/* Video Info */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Title & Metadata */}
            <div className="bg-gray-800 rounded-lg p-6">
              <h1 className="text-3xl font-bold text-white mb-3">{video.title}</h1>
              <p className="text-gray-400 mb-4">{video.description}</p>

              {/* Stats */}
              <div className="flex flex-wrap gap-6 text-sm text-gray-400 mb-4">
                <div>
                  <span className="text-gray-500">Duration:</span>
                  <span className="text-white ml-2 font-semibold">
                    {Math.floor(video.duration / 60)} min
                  </span>
                </div>
                <div>
                  <span className="text-gray-500">Uploaded:</span>
                  <span className="text-white ml-2 font-semibold">
                    {new Date(video.createdAt).toLocaleDateString()}
                  </span>
                </div>
              </div>

              {/* Actions */}
              <div className="flex flex-wrap gap-2">
                <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors">
                  <Share2 className="w-4 h-4" />
                  Share
                </button>
                <button className="flex items-center gap-2 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors">
                  <Flag className="w-4 h-4" />
                  Report Issue
                </button>
              </div>
            </div>

            {/* Description & Details */}
            <div className="bg-gray-800 rounded-lg p-6">
              <h2 className="text-xl font-bold text-white mb-4">About This Video</h2>
              <p className="text-gray-400 leading-relaxed">{video.description}</p>

              {/* Lecture & Course Info */}
              <div className="mt-6 grid grid-cols-2 gap-4">
                <div className="bg-gray-700 rounded-lg p-4">
                  <p className="text-sm text-gray-400 mb-1">Lecture ID</p>
                  <p className="text-white font-semibold">{video.lectureId}</p>
                </div>
                <div className="bg-gray-700 rounded-lg p-4">
                  <p className="text-sm text-gray-400 mb-1">Course ID</p>
                  <p className="text-white font-semibold">{video.courseId}</p>
                </div>
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Watch History */}
            <WatchHistory
              userId={authStore.user?.id || ''}
              limit={5}
              onVideoSelect={(id) => router.push(`/watch/${id}`)}
            />

            {/* Recommended Videos */}
            <RecommendedVideos
              limit={3}
              onVideoSelect={(id) => router.push(`/watch/${id}`)}
            />
          </div>
        </div>

        {/* Recommended Section */}
        <div className="mt-12">
          <RecommendedVideos
            limit={6}
            onVideoSelect={(id) => router.push(`/watch/${id}`)}
          />
        </div>
      </div>

      {/* Settings Modal */}
      {showSettings && (
        <div
          className="fixed bottom-20 left-4 z-50"
          onClick={() => setShowSettings(false)}
        >
          <PlaybackSettings
            currentSpeed={currentSpeed}
            availableSpeeds={PLAYBACK_SPEEDS}
            onSpeedChange={setCurrentSpeed}
            onClose={() => setShowSettings(false)}
          />
        </div>
      )}

      {/* Quality Selector Modal */}
      {showQuality && video.recordings && (
        <div
          className="fixed bottom-20 right-4 z-50"
          onClick={() => setShowQuality(false)}
        >
          <QualitySelector
            qualities={video.recordings.map(r => ({
              height: parseInt(r.quality),
              bitrate: r.bitrate,
              name: r.quality,
            }))}
            currentQuality={video.recordings[0]?.quality || '1080p'}
            onQualityChange={() => setShowQuality(false)}
          />
        </div>
      )}

      {/* Subtitle Selector Modal */}
      {showSubtitles && subtitles.length > 0 && (
        <div
          className="fixed bottom-20 right-4 z-50"
          onClick={() => setShowSubtitles(false)}
        >
          <SubtitleSelector
            subtitles={subtitles}
            currentSubtitle={currentSubtitle}
            onSubtitleChange={(lang) => {
              setCurrentSubtitle(lang);
              setShowSubtitles(false);
            }}
          />
        </div>
      )}
    </div>
  );
}
