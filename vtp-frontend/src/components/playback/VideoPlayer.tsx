import { useEffect, useState } from 'react';
import { Play, Pause, Volume2, VolumeX, Maximize2, Minimize2, Settings } from 'lucide-react';
import { useVideoPlayer } from '@/hooks/useVideoPlayer';
import { cn } from '@/utils/cn';

interface VideoPlayerProps {
  videoUrl: string;
  videoId: string;
  title?: string;
  thumbnail?: string;
  onProgress?: (time: number, duration: number) => void;
  onComplete?: () => void;
  showControls?: boolean;
  autoPlay?: boolean;
  className?: string;
}

export const VideoPlayer: React.FC<VideoPlayerProps> = ({
  videoUrl,
  videoId,
  title,
  thumbnail,
  onProgress,
  onComplete,
  showControls = true,
  autoPlay = false,
  className,
}) => {
  const { videoRef, state, controls } = useVideoPlayer(videoUrl, videoId, (currentTime) => {
    onProgress?.(currentTime, state.duration);
  });

  const [showVolumeSlider, setShowVolumeSlider] = useState(false);
  const [hideControlsTimeout, setHideControlsTimeout] = useState<NodeJS.Timeout | null>(null);
  const [showControls_, setShowControls_] = useState(true);

  // Auto-hide controls on mouse inactivity
  const handleMouseMove = () => {
    setShowControls_(true);
    
    if (hideControlsTimeout) clearTimeout(hideControlsTimeout);
    
    if (state.isPlaying) {
      const timeout = setTimeout(() => setShowControls_(false), 3000);
      setHideControlsTimeout(timeout);
    }
  };

  useEffect(() => {
    return () => {
      if (hideControlsTimeout) clearTimeout(hideControlsTimeout);
    };
  }, [hideControlsTimeout]);

  const formatTime = (seconds: number) => {
    if (!Number.isFinite(seconds)) return '0:00';
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);
    return h > 0
      ? `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
      : `${m}:${s.toString().padStart(2, '0')}`;
  };

  const handlePlayPause = async () => {
    if (state.isPlaying) {
      await controls.pause();
    } else {
      await controls.play();
    }
  };

  return (
    <div
      className={cn('w-full bg-black rounded-lg overflow-hidden group', className)}
      onMouseMove={handleMouseMove}
      onMouseLeave={() => {
        if (state.isPlaying) setShowControls_(false);
      }}
    >
      {/* Video Container */}
      <div className="relative w-full bg-gray-900">
        <video
          ref={videoRef}
          className="w-full h-auto"
          poster={thumbnail}
          autoPlay={autoPlay}
          onClick={handlePlayPause}
        />

        {/* Loading Indicator */}
        {state.isLoading && (
          <div className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-50">
            <div className="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
          </div>
        )}

        {/* Error Message */}
        {state.error && (
          <div className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-75">
            <div className="text-center text-white">
              <p className="text-lg font-semibold mb-2">Playback Error</p>
              <p className="text-gray-300 text-sm">{state.error}</p>
            </div>
          </div>
        )}

        {/* Play Button Overlay */}
        {!state.isPlaying && !state.isLoading && (
          <button
            onClick={handlePlayPause}
            className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-40 hover:bg-opacity-50 transition-colors group/play"
          >
            <div className="w-20 h-20 bg-blue-600 rounded-full flex items-center justify-center hover:bg-blue-700 transition-colors">
              <Play className="w-8 h-8 text-white ml-1" fill="white" />
            </div>
          </button>
        )}

        {/* Controls Container */}
        {showControls && showControls_ && (
          <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black via-black/50 to-transparent pt-12 pb-3 px-3 transition-opacity duration-300">
            {/* Progress Bar */}
            <div className="flex items-center gap-2 mb-3">
              <div className="flex-1 group/progress">
                <input
                  type="range"
                  min="0"
                  max={state.duration || 0}
                  value={state.currentTime}
                  onChange={(e) => controls.seek(parseFloat(e.target.value))}
                  className="w-full h-1 bg-gray-600 rounded-lg appearance-none cursor-pointer group-hover/progress:h-2 transition-all"
                  style={{
                    background: `linear-gradient(to right, #3b82f6 0%, #3b82f6 ${
                      (state.currentTime / (state.duration || 1)) * 100
                    }%, #4b5563 ${(state.currentTime / (state.duration || 1)) * 100}%, #4b5563 100%)`,
                  }}
                />
              </div>
            </div>

            {/* Control Buttons */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {/* Play/Pause */}
                <button
                  onClick={handlePlayPause}
                  className="p-2 text-white hover:bg-white hover:bg-opacity-20 rounded-lg transition-colors"
                  title={state.isPlaying ? 'Pause' : 'Play'}
                >
                  {state.isPlaying ? (
                    <Pause className="w-5 h-5" fill="white" />
                  ) : (
                    <Play className="w-5 h-5" fill="white" />
                  )}
                </button>

                {/* Volume Control */}
                <div className="relative group/volume">
                  <button
                    onClick={controls.toggleMute}
                    className="p-2 text-white hover:bg-white hover:bg-opacity-20 rounded-lg transition-colors"
                    title={state.isMuted ? 'Unmute' : 'Mute'}
                  >
                    {state.isMuted ? (
                      <VolumeX className="w-5 h-5" />
                    ) : (
                      <Volume2 className="w-5 h-5" />
                    )}
                  </button>

                  {/* Volume Slider */}
                  {showVolumeSlider && (
                    <div className="absolute bottom-full left-0 mb-2 bg-gray-800 rounded-lg p-2 flex flex-col gap-2">
                      <input
                        type="range"
                        min="0"
                        max="1"
                        step="0.1"
                        value={state.volume}
                        onChange={(e) => controls.setVolume(parseFloat(e.target.value))}
                        className="w-1 h-20 appearance-none cursor-pointer"
                        style={{
                          writingMode: 'bt-lr',
                        }}
                      />
                    </div>
                  )}
                </div>

                {/* Time Display */}
                <div className="text-white text-sm font-mono ml-2">
                  {formatTime(state.currentTime)} / {formatTime(state.duration)}
                </div>
              </div>

              {/* Right Controls */}
              <div className="flex items-center gap-2">
                {/* Quality Selector */}
                {state.availableQualities.length > 0 && (
                  <div className="group/quality">
                    <button
                      className="p-2 text-white hover:bg-white hover:bg-opacity-20 rounded-lg transition-colors text-sm font-semibold"
                      title="Quality"
                    >
                      {state.currentQuality === 'auto' ? 'Auto' : state.currentQuality}
                    </button>
                    <div className="absolute bottom-full right-0 mb-2 bg-gray-800 rounded-lg overflow-hidden opacity-0 group-hover/quality:opacity-100 transition-opacity">
                      <button
                        onClick={() => controls.setQuality(0)}
                        className={cn(
                          'block w-full px-4 py-2 text-white text-left hover:bg-gray-700 transition-colors',
                          state.currentQuality === 'auto' && 'bg-blue-600'
                        )}
                      >
                        Auto
                      </button>
                      {state.availableQualities.map(quality => (
                        <button
                          key={quality.height}
                          onClick={() => controls.setQuality(quality.height)}
                          className={cn(
                            'block w-full px-4 py-2 text-white text-left hover:bg-gray-700 transition-colors',
                            state.currentQuality === quality.name && 'bg-blue-600'
                          )}
                        >
                          {quality.name}
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                {/* Fullscreen */}
                <button
                  onClick={controls.toggleFullscreen}
                  className="p-2 text-white hover:bg-white hover:bg-opacity-20 rounded-lg transition-colors"
                  title={state.isFullscreen ? 'Exit fullscreen' : 'Fullscreen'}
                >
                  {state.isFullscreen ? (
                    <Minimize2 className="w-5 h-5" />
                  ) : (
                    <Maximize2 className="w-5 h-5" />
                  )}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Title */}
      {title && <div className="p-4 text-white font-semibold">{title}</div>}
    </div>
  );
};
