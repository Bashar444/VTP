import { api } from './api.client';

export interface VideoMetadata {
  id: string;
  title: string;
  description: string;
  lectureId: string;
  courseId: string;
  duration: number;
  thumbnail: string;
  hlsUrl: string;
  recordings: VideoRecording[];
  createdAt: Date;
  updatedAt: Date;
}

export interface VideoRecording {
  id: string;
  quality: '360p' | '480p' | '720p' | '1080p';
  bitrate: number;
  url: string;
  size: number;
}

export interface PlaybackHistory {
  id: string;
  videoId: string;
  userId: string;
  watchedAt: Date;
  watchedDuration: number;
  totalDuration: number;
  completedPercentage: number;
}

export interface VideoAnalytics {
  videoId: string;
  totalViews: number;
  averageWatchTime: number;
  completionRate: number;
  engagementScore: number;
  lastUpdated: Date;
}

export interface VideoProgress {
  videoId: string;
  currentTime: number;
  totalDuration: number;
  completionPercentage: number;
  lastWatchedAt: Date;
}

class VideoService {
  /**
   * Get video metadata by ID
   */
  async getVideoMetadata(videoId: string): Promise<VideoMetadata> {
    const response = await api.get(`/videos/${videoId}`);
    return response.data;
  }

  /**
   * Get videos for a lecture
   */
  async getLectureVideos(lectureId: string): Promise<VideoMetadata[]> {
    const response = await api.get(`/lectures/${lectureId}/videos`);
    return response.data;
  }

  /**
   * Get videos for a course
   */
  async getCourseVideos(courseId: string): Promise<VideoMetadata[]> {
    const response = await api.get(`/courses/${courseId}/videos`);
    return response.data;
  }

  /**
   * Get user's watch history
   */
  async getWatchHistory(userId: string, limit: number = 10): Promise<PlaybackHistory[]> {
    const response = await api.get(`/users/${userId}/watch-history`, {
      params: { limit },
    });
    return response.data;
  }

  /**
   * Get watch history for a specific video
   */
  async getVideoWatchers(videoId: string): Promise<PlaybackHistory[]> {
    const response = await api.get(`/videos/${videoId}/watchers`);
    return response.data;
  }

  /**
   * Record playback start
   */
  async recordPlaybackStart(videoId: string): Promise<void> {
    await api.post(`/videos/${videoId}/playback/start`, {
      startedAt: new Date(),
    });
  }

  /**
   * Update playback progress
   */
  async updatePlaybackProgress(
    videoId: string,
    currentTime: number,
    totalDuration: number
  ): Promise<VideoProgress> {
    const response = await api.post(`/videos/${videoId}/playback/progress`, {
      currentTime,
      totalDuration,
      completionPercentage: (currentTime / totalDuration) * 100,
      lastWatchedAt: new Date(),
    });
    return response.data;
  }

  /**
   * Record playback completion
   */
  async recordPlaybackCompletion(videoId: string, totalWatchedDuration: number): Promise<void> {
    await api.post(`/videos/${videoId}/playback/complete`, {
      completedAt: new Date(),
      totalWatchedDuration,
      completionPercentage: 100,
    });
  }

  /**
   * Get current playback progress
   */
  async getPlaybackProgress(videoId: string): Promise<VideoProgress> {
    const response = await api.get(`/videos/${videoId}/playback/progress`);
    return response.data;
  }

  /**
   * Get video analytics
   */
  async getVideoAnalytics(videoId: string): Promise<VideoAnalytics> {
    const response = await api.get(`/videos/${videoId}/analytics`);
    return response.data;
  }

  /**
   * Get course video analytics
   */
  async getCourseVideoAnalytics(courseId: string): Promise<VideoAnalytics[]> {
    const response = await api.get(`/courses/${courseId}/videos/analytics`);
    return response.data;
  }

  /**
   * Create video subtitle/caption
   */
  async createSubtitle(
    videoId: string,
    language: string,
    content: string
  ): Promise<{ id: string; language: string; url: string }> {
    const response = await api.post(`/videos/${videoId}/subtitles`, {
      language,
      content,
    });
    return response.data;
  }

  /**
   * Get video subtitles
   */
  async getSubtitles(videoId: string): Promise<Array<{ language: string; url: string }>> {
    const response = await api.get(`/videos/${videoId}/subtitles`);
    return response.data;
  }

  /**
   * Report video issue
   */
  async reportIssue(
    videoId: string,
    issueType: string,
    description: string,
    timestamp?: number
  ): Promise<{ id: string; status: string }> {
    const response = await api.post(`/videos/${videoId}/report`, {
      issueType,
      description,
      timestamp,
      reportedAt: new Date(),
    });
    return response.data;
  }

  /**
   * Get recommended videos based on view history
   */
  async getRecommendedVideos(limit: number = 5): Promise<VideoMetadata[]> {
    const response = await api.get(`/videos/recommended`, {
      params: { limit },
    });
    return response.data;
  }

  /**
   * Search videos
   */
  async searchVideos(query: string, limit: number = 10): Promise<VideoMetadata[]> {
    const response = await api.get(`/videos/search`, {
      params: { q: query, limit },
    });
    return response.data;
  }
}

export default new VideoService();
