import { useEffect, useRef, useState, useCallback } from 'react';
import Hls from 'hls.js';
import VideoService from '@/services/video.service';

export interface VideoPlayerState {
  isPlaying: boolean;
  currentTime: number;
  duration: number;
  volume: number;
  isMuted: boolean;
  isFullscreen: boolean;
  currentQuality: string;
  availableQualities: Array<{ height: number; bitrate: number; name: string }>;
  isLoading: boolean;
  error: string | null;
  bufferedTime: number;
}

export interface VideoPlayerControls {
  play: () => Promise<void>;
  pause: () => Promise<void>;
  seek: (time: number) => void;
  setVolume: (volume: number) => void;
  toggleMute: () => void;
  toggleFullscreen: () => void;
  setQuality: (qualityHeight: number) => void;
  setPlaybackSpeed: (speed: number) => void;
}

export const useVideoPlayer = (
  videoUrl: string,
  videoId: string,
  onTimeUpdate?: (currentTime: number) => void
) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const hlsRef = useRef<Hls | null>(null);
  const progressIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const [state, setState] = useState<VideoPlayerState>({
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    volume: 1,
    isMuted: false,
    isFullscreen: false,
    currentQuality: 'auto',
    availableQualities: [],
    isLoading: false,
    error: null,
    bufferedTime: 0,
  });

  // Initialize HLS player
  useEffect(() => {
    if (!videoRef.current || !videoUrl) return;

    const video = videoRef.current;

    try {
      setState(prev => ({ ...prev, isLoading: true }));

      if (Hls.isSupported()) {
        const hls = new Hls({
          debug: false,
          enableWorker: true,
          lowLatencyMode: false,
          backBufferLength: 90,
        });

        hlsRef.current = hls;

        hls.loadSource(videoUrl);
        hls.attachMedia(video);

        // Parse available qualities
        hls.on(Hls.Events.MANIFEST_PARSED, () => {
          const levels = hls.levels.map((level) => ({
            height: level.height,
            bitrate: level.bitrate,
            name: `${level.height}p`,
          }));

          setState(prev => ({
            ...prev,
            availableQualities: levels,
            isLoading: false,
          }));
        });

        hls.on(Hls.Events.ERROR, (event, data) => {
          console.error('HLS error:', event, data);
          if (data.fatal) {
            setState(prev => ({
              ...prev,
              error: `Playback error: ${data.details}`,
            }));
          }
        });

        // Listen for quality changes
        hls.on(Hls.Events.LEVEL_SWITCHING, (event, data) => {
          const level = hls.levels[data.level];
          setState(prev => ({
            ...prev,
            currentQuality: level.height.toString(),
          }));
        });

        return () => {
          hls.destroy();
          hlsRef.current = null;
        };
      } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
        // Native HLS support (Safari)
        video.src = videoUrl;
        setState(prev => ({ ...prev, isLoading: false }));
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to initialize player';
      setState(prev => ({ ...prev, error: message, isLoading: false }));
    }
  }, [videoUrl]);

  // Record playback start
  useEffect(() => {
    VideoService.recordPlaybackStart(videoId).catch(err =>
      console.error('Failed to record playback start:', err)
    );
  }, [videoId]);

  // Update playback progress periodically
  useEffect(() => {
    if (state.isPlaying && videoRef.current) {
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
      }

      progressIntervalRef.current = setInterval(() => {
        const video = videoRef.current;
        if (video && state.duration > 0) {
          const currentTime = video.currentTime;
          setState(prev => ({
            ...prev,
            currentTime,
            bufferedTime: video.buffered.length > 0 ? video.buffered.end(0) : 0,
          }));

          onTimeUpdate?.(currentTime);

          // Update progress every 10 seconds
          if (Math.floor(currentTime) % 10 === 0) {
            VideoService.updatePlaybackProgress(videoId, currentTime, state.duration).catch(
              err => console.error('Failed to update progress:', err)
            );
          }
        }
      }, 500);
    }

    return () => {
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
      }
    };
  }, [state.isPlaying, state.duration, videoId, onTimeUpdate]);

  // Video event listeners
  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const handlePlay = () => {
      setState(prev => ({ ...prev, isPlaying: true, error: null }));
    };

    const handlePause = () => {
      setState(prev => ({ ...prev, isPlaying: false }));
    };

    const handleLoadedMetadata = () => {
      setState(prev => ({
        ...prev,
        duration: video.duration,
        isLoading: false,
      }));
    };

    const handleEnded = async () => {
      setState(prev => ({ ...prev, isPlaying: false }));
      try {
        await VideoService.recordPlaybackCompletion(videoId, video.currentTime);
      } catch (err) {
        console.error('Failed to record completion:', err);
      }
    };

    const handleError = () => {
      setState(prev => ({
        ...prev,
        error: 'Failed to load video. Please try again.',
      }));
    };

    video.addEventListener('play', handlePlay);
    video.addEventListener('pause', handlePause);
    video.addEventListener('loadedmetadata', handleLoadedMetadata);
    video.addEventListener('ended', handleEnded);
    video.addEventListener('error', handleError);

    return () => {
      video.removeEventListener('play', handlePlay);
      video.removeEventListener('pause', handlePause);
      video.removeEventListener('loadedmetadata', handleLoadedMetadata);
      video.removeEventListener('ended', handleEnded);
      video.removeEventListener('error', handleError);
    };
  }, [videoId]);

  // Player controls
  const controls: VideoPlayerControls = {
    play: async () => {
      const video = videoRef.current;
      if (video) {
        try {
          await video.play();
        } catch (err) {
          console.error('Play failed:', err);
        }
      }
    },

    pause: async () => {
      const video = videoRef.current;
      if (video) {
        video.pause();
      }
    },

    seek: (time: number) => {
      const video = videoRef.current;
      if (video) {
        video.currentTime = Math.max(0, Math.min(time, video.duration));
        setState(prev => ({ ...prev, currentTime: video.currentTime }));
      }
    },

    setVolume: (volume: number) => {
      const video = videoRef.current;
      if (video) {
        const clampedVolume = Math.max(0, Math.min(1, volume));
        video.volume = clampedVolume;
        setState(prev => ({
          ...prev,
          volume: clampedVolume,
          isMuted: clampedVolume === 0,
        }));
      }
    },

    toggleMute: () => {
      const video = videoRef.current;
      if (video) {
        video.muted = !video.muted;
        setState(prev => ({ ...prev, isMuted: !prev.isMuted }));
      }
    },

    toggleFullscreen: () => {
      const video = videoRef.current;
      if (video && video.parentElement) {
        if (!document.fullscreenElement) {
          video.parentElement.requestFullscreen().catch(err =>
            console.error('Fullscreen request failed:', err)
          );
          setState(prev => ({ ...prev, isFullscreen: true }));
        } else {
          document.exitFullscreen();
          setState(prev => ({ ...prev, isFullscreen: false }));
        }
      }
    },

    setQuality: (qualityHeight: number) => {
      if (hlsRef.current) {
        if (qualityHeight === 0) {
          // Auto quality
          hlsRef.current.currentLevel = -1;
        } else {
          const levelIndex = hlsRef.current.levels.findIndex(
            level => level.height === qualityHeight
          );
          if (levelIndex >= 0) {
            hlsRef.current.currentLevel = levelIndex;
          }
        }
      }
    },

    setPlaybackSpeed: (speed: number) => {
      const video = videoRef.current;
      if (video) {
        video.playbackRate = speed;
      }
    },
  };

  return {
    videoRef,
    state,
    controls,
  };
};
