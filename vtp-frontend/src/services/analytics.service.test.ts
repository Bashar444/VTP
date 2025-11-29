import { describe, it, expect, beforeEach, vi } from 'vitest';
import { AnalyticsService } from '@/services/analytics.service';
import { api } from '@/lib/api';

vi.mock('@/lib/api');

describe('AnalyticsService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getDashboardMetrics', () => {
    it('should fetch dashboard metrics', async () => {
      const mockMetrics = {
        totalStudents: 1000,
        totalCourses: 25,
        activeUsers: 450,
        totalRevenue: 15000,
        avgEngagementRate: 75,
        avgCourseCompletion: 68,
        newSignupsToday: 12,
        newEnrollmentsToday: 35,
      };
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockMetrics });

      const result = await AnalyticsService.getDashboardMetrics();

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/dashboard');
      expect(result).toEqual(mockMetrics);
    });
  });

  describe('getEngagementMetrics', () => {
    it('should fetch engagement metrics with date range', async () => {
      const mockData = [
        {
          date: '2024-01-01',
          activeUsers: 450,
          videoWatchTime: 2000,
          lecturesCompleted: 150,
          reviewsSubmitted: 45,
          avgSessionDuration: 25,
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockData });

      const result = await AnalyticsService.getEngagementMetrics(
        '2024-01-01',
        '2024-01-31',
        'daily'
      );

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/engagement', {
        params: {
          startDate: '2024-01-01',
          endDate: '2024-01-31',
          interval: 'daily',
        },
      });
      expect(result).toEqual(mockData);
    });
  });

  describe('getCoursePerformance', () => {
    it('should fetch course performance data', async () => {
      const mockData = [
        {
          courseId: 'course-1',
          courseName: 'JavaScript Basics',
          enrollmentCount: 500,
          completionRate: 75,
          avgRating: 4.5,
          revenue: 5000,
          totalWatchTime: 1000,
          activeStudents: 250,
          dropoutRate: 15,
          lastUpdated: '2024-01-15',
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockData });

      const result = await AnalyticsService.getCoursePerformance();

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/courses', {
        params: {},
      });
      expect(result).toEqual(mockData);
    });
  });

  describe('getTopPerformingCourses', () => {
    it('should fetch top performing courses with limit', async () => {
      const mockData = [
        {
          courseId: 'course-1',
          courseName: 'JavaScript Basics',
          enrollmentCount: 500,
          completionRate: 75,
          avgRating: 4.5,
          revenue: 5000,
          totalWatchTime: 1000,
          activeStudents: 250,
          dropoutRate: 15,
          lastUpdated: '2024-01-15',
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockData });

      const result = await AnalyticsService.getTopPerformingCourses(10);

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/top-courses', {
        params: { limit: 10 },
      });
      expect(result).toEqual(mockData);
    });
  });

  describe('getCourseCompletionRate', () => {
    it('should fetch completion rate for a specific course', async () => {
      vi.mocked(api.get).mockResolvedValueOnce({ data: { rate: 75 } });

      const result = await AnalyticsService.getCourseCompletionRate('course-1');

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/courses/course-1/completion-rate');
      expect(result).toBe(75);
    });
  });

  describe('getStudentMetrics', () => {
    it('should fetch student metrics', async () => {
      const mockData = [
        {
          studentId: 'student-1',
          studentName: 'John Doe',
          email: 'john@example.com',
          enrolledCourses: 5,
          completedCourses: 2,
          totalWatchTime: 500,
          avgCourseCompletion: 60,
          lastActiveAt: '2024-01-15',
          engagementScore: 78,
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockData });

      const result = await AnalyticsService.getStudentMetrics();

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/students', {
        params: {},
      });
      expect(result).toEqual(mockData);
    });
  });

  describe('getTopStudents', () => {
    it('should fetch top performing students', async () => {
      const mockData = [
        {
          studentId: 'student-1',
          studentName: 'John Doe',
          email: 'john@example.com',
          enrolledCourses: 10,
          completedCourses: 8,
          totalWatchTime: 2000,
          avgCourseCompletion: 95,
          lastActiveAt: '2024-01-15',
          engagementScore: 98,
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockData });

      const result = await AnalyticsService.getTopStudents(10);

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/top-students', {
        params: { limit: 10 },
      });
      expect(result).toEqual(mockData);
    });
  });

  describe('getVideoAnalytics', () => {
    it('should fetch video analytics', async () => {
      const mockData = [
        {
          videoId: 'video-1',
          videoTitle: 'Introduction to JavaScript',
          courseId: 'course-1',
          courseName: 'JavaScript Basics',
          views: 500,
          uniqueViewers: 400,
          avgWatchTime: 25,
          completionRate: 75,
          avgQualitySelected: '720p',
          rewindCount: 150,
          lastUpdated: '2024-01-15',
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockData });

      const result = await AnalyticsService.getVideoAnalytics();

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/videos', {
        params: {},
      });
      expect(result).toEqual(mockData);
    });
  });

  describe('getRevenueMetrics', () => {
    it('should fetch revenue metrics with date range', async () => {
      const mockData = [
        {
          date: '2024-01-15',
          courseSales: 10,
          totalRevenue: 500,
          transactions: 10,
          avgTransactionValue: 50,
          topCourse: {
            courseId: 'course-1',
            courseName: 'JavaScript Basics',
            revenue: 300,
          },
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockData });

      const result = await AnalyticsService.getRevenueMetrics('2024-01-01', '2024-01-31');

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/revenue', {
        params: {
          startDate: '2024-01-01',
          endDate: '2024-01-31',
        },
      });
      expect(result).toEqual(mockData);
    });
  });

  describe('getTotalRevenue', () => {
    it('should fetch total revenue for a period', async () => {
      vi.mocked(api.get).mockResolvedValueOnce({ data: { total: 15000 } });

      const result = await AnalyticsService.getTotalRevenue(30);

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/total-revenue', {
        params: { days: 30 },
      });
      expect(result).toBe(15000);
    });
  });

  describe('getRetentionMetrics', () => {
    it('should fetch retention metrics', async () => {
      const mockData = [
        {
          month: 'January',
          startingUsers: 1000,
          churnedUsers: 50,
          retentionRate: 95,
          avgCoursesPerStudent: 2.5,
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockData });

      const result = await AnalyticsService.getRetentionMetrics();

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/retention');
      expect(result).toEqual(mockData);
    });
  });

  describe('getChurnRate', () => {
    it('should fetch churn rate', async () => {
      vi.mocked(api.get).mockResolvedValueOnce({ data: { rate: 5 } });

      const result = await AnalyticsService.getChurnRate();

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/churn-rate');
      expect(result).toBe(5);
    });
  });

  describe('getSystemAlerts', () => {
    it('should fetch system alerts', async () => {
      const mockAlerts = [
        {
          id: 'alert-1',
          type: 'warning' as const,
          title: 'High Load',
          message: 'Server load exceeding 80%',
          severity: 'high' as const,
          createdAt: '2024-01-15T10:00:00Z',
          resolved: false,
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockAlerts });

      const result = await AnalyticsService.getSystemAlerts(true);

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/alerts', {
        params: { unreadOnly: true },
      });
      expect(result).toEqual(mockAlerts);
    });
  });

  describe('markAlertAsResolved', () => {
    it('should mark an alert as resolved', async () => {
      const mockAlert = {
        id: 'alert-1',
        type: 'warning' as const,
        title: 'High Load',
        message: 'Server load exceeding 80%',
        severity: 'high' as const,
        createdAt: '2024-01-15T10:00:00Z',
        resolved: true,
      };
      vi.mocked(api.post).mockResolvedValueOnce({ data: mockAlert });

      const result = await AnalyticsService.markAlertAsResolved('alert-1');

      expect(api.post).toHaveBeenCalledWith('/api/v1/analytics/alerts/alert-1/resolve', {});
      expect(result.resolved).toBe(true);
    });
  });

  describe('getSystemHealth', () => {
    it('should fetch system health status', async () => {
      const mockHealth = {
        status: 'healthy' as const,
        uptime: 99.9,
        databaseHealth: 'healthy',
        apiResponseTime: 150,
        errorRate: 0.1,
        lastCheck: '2024-01-15T10:00:00Z',
      };
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockHealth });

      const result = await AnalyticsService.getSystemHealth();

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/system-health');
      expect(result.status).toBe('healthy');
      expect(result.uptime).toBe(99.9);
    });
  });

  describe('getPredictedChurn', () => {
    it('should fetch predicted churn data', async () => {
      const mockData = [
        {
          studentId: 'student-1',
          studentName: 'John Doe',
          churnRisk: 85,
          riskFactors: ['Low engagement', 'No recent activity'],
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockData });

      const result = await AnalyticsService.getPredictedChurn();

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/predictions/churn');
      expect(result).toEqual(mockData);
    });
  });

  describe('getRecommendations', () => {
    it('should fetch AI recommendations', async () => {
      const mockRecommendations = [
        {
          type: 'content' as const,
          priority: 'high' as const,
          title: 'Add More Interactive Content',
          description: 'Students respond better to interactive content',
          estimatedImpact: '+15% engagement',
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockRecommendations });

      const result = await AnalyticsService.getRecommendations();

      expect(api.get).toHaveBeenCalledWith('/api/v1/analytics/recommendations');
      expect(result).toEqual(mockRecommendations);
    });
  });
});
