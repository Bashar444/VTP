import { describe, it, expect, beforeEach, vi } from 'vitest';
import VideoService, { VideoMetadata, PlaybackHistory } from '@/services/video.service';

// Mock API client
vi.mock('@/services/api.client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

import { api } from '@/services/api.client';

describe('VideoService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should get video metadata', async () => {
    const mockVideo: VideoMetadata = {
      id: 'video-1',
      title: 'Test Video',
      description: 'Test Description',
      lectureId: 'lecture-1',
      courseId: 'course-1',
      duration: 3600,
      thumbnail: 'https://example.com/thumb.jpg',
      hlsUrl: 'https://example.com/stream.m3u8',
      recordings: [],
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    (api.get as any).mockResolvedValue({ data: mockVideo });

    const result = await VideoService.getVideoMetadata('video-1');

    expect(api.get).toHaveBeenCalledWith('/videos/video-1');
    expect(result).toEqual(mockVideo);
  });

  it('should get lecture videos', async () => {
    const mockVideos = [
      {
        id: 'video-1',
        title: 'Lecture 1',
      },
    ];

    (api.get as any).mockResolvedValue({ data: mockVideos });

    const result = await VideoService.getLectureVideos('lecture-1');

    expect(api.get).toHaveBeenCalledWith('/lectures/lecture-1/videos');
    expect(result).toEqual(mockVideos);
  });

  it('should get course videos', async () => {
    const mockVideos = [
      { id: 'video-1', title: 'Video 1' },
      { id: 'video-2', title: 'Video 2' },
    ];

    (api.get as any).mockResolvedValue({ data: mockVideos });

    const result = await VideoService.getCourseVideos('course-1');

    expect(api.get).toHaveBeenCalledWith('/courses/course-1/videos');
    expect(result).toEqual(mockVideos);
  });

  it('should get watch history', async () => {
    const mockHistory: PlaybackHistory[] = [
      {
        id: 'history-1',
        videoId: 'video-1',
        userId: 'user-1',
        watchedAt: new Date(),
        watchedDuration: 1800,
        totalDuration: 3600,
        completedPercentage: 50,
      },
    ];

    (api.get as any).mockResolvedValue({ data: mockHistory });

    const result = await VideoService.getWatchHistory('user-1', 10);

    expect(api.get).toHaveBeenCalledWith('/users/user-1/watch-history', {
      params: { limit: 10 },
    });
    expect(result).toEqual(mockHistory);
  });

  it('should record playback start', async () => {
    (api.post as any).mockResolvedValue({ data: {} });

    await VideoService.recordPlaybackStart('video-1');

    expect(api.post).toHaveBeenCalledWith(
      '/videos/video-1/playback/start',
      expect.objectContaining({
        startedAt: expect.any(Date),
      })
    );
  });

  it('should update playback progress', async () => {
    const mockProgress = {
      videoId: 'video-1',
      currentTime: 1800,
      totalDuration: 3600,
      completionPercentage: 50,
      lastWatchedAt: new Date(),
    };

    (api.post as any).mockResolvedValue({ data: mockProgress });

    const result = await VideoService.updatePlaybackProgress('video-1', 1800, 3600);

    expect(api.post).toHaveBeenCalledWith(
      '/videos/video-1/playback/progress',
      expect.objectContaining({
        currentTime: 1800,
        totalDuration: 3600,
        completionPercentage: 50,
      })
    );
    expect(result).toEqual(mockProgress);
  });

  it('should record playback completion', async () => {
    (api.post as any).mockResolvedValue({ data: {} });

    await VideoService.recordPlaybackCompletion('video-1', 3600);

    expect(api.post).toHaveBeenCalledWith(
      '/videos/video-1/playback/complete',
      expect.objectContaining({
        completedAt: expect.any(Date),
        totalWatchedDuration: 3600,
        completionPercentage: 100,
      })
    );
  });

  it('should get video analytics', async () => {
    const mockAnalytics = {
      videoId: 'video-1',
      totalViews: 1000,
      averageWatchTime: 1800,
      completionRate: 75,
      engagementScore: 8.5,
      lastUpdated: new Date(),
    };

    (api.get as any).mockResolvedValue({ data: mockAnalytics });

    const result = await VideoService.getVideoAnalytics('video-1');

    expect(api.get).toHaveBeenCalledWith('/videos/video-1/analytics');
    expect(result).toEqual(mockAnalytics);
  });

  it('should get recommended videos', async () => {
    const mockRecommended = [
      { id: 'video-1', title: 'Recommended 1' },
      { id: 'video-2', title: 'Recommended 2' },
    ];

    (api.get as any).mockResolvedValue({ data: mockRecommended });

    const result = await VideoService.getRecommendedVideos(5);

    expect(api.get).toHaveBeenCalledWith('/videos/recommended', {
      params: { limit: 5 },
    });
    expect(result).toEqual(mockRecommended);
  });

  it('should search videos', async () => {
    const mockResults = [
      { id: 'video-1', title: 'Search Result 1' },
    ];

    (api.get as any).mockResolvedValue({ data: mockResults });

    const result = await VideoService.searchVideos('react', 10);

    expect(api.get).toHaveBeenCalledWith('/videos/search', {
      params: { q: 'react', limit: 10 },
    });
    expect(result).toEqual(mockResults);
  });

  it('should get subtitles', async () => {
    const mockSubtitles = [
      { language: 'en', url: 'https://example.com/en.vtt' },
      { language: 'ar', url: 'https://example.com/ar.vtt' },
    ];

    (api.get as any).mockResolvedValue({ data: mockSubtitles });

    const result = await VideoService.getSubtitles('video-1');

    expect(api.get).toHaveBeenCalledWith('/videos/video-1/subtitles');
    expect(result).toEqual(mockSubtitles);
  });

  it('should report video issue', async () => {
    const mockResponse = {
      id: 'report-1',
      status: 'submitted',
    };

    (api.post as any).mockResolvedValue({ data: mockResponse });

    const result = await VideoService.reportIssue(
      'video-1',
      'buffering',
      'Video keeps buffering',
      100
    );

    expect(api.post).toHaveBeenCalledWith(
      '/videos/video-1/report',
      expect.objectContaining({
        issueType: 'buffering',
        description: 'Video keeps buffering',
        timestamp: 100,
      })
    );
    expect(result).toEqual(mockResponse);
  });
});
