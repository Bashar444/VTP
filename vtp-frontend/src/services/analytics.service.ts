import { api } from '@/lib/api';

// Type Definitions
export interface DashboardMetrics {
  totalStudents: number;
  totalCourses: number;
  activeUsers: number;
  totalRevenue: number;
  avgEngagementRate: number;
  avgCourseCompletion: number;
  newSignupsToday: number;
  newEnrollmentsToday: number;
}

export interface EngagementMetrics {
  date: string;
  activeUsers: number;
  videoWatchTime: number;
  lecturesCompleted: number;
  reviewsSubmitted: number;
  avgSessionDuration: number;
}

export interface CoursePerformance {
  courseId: string;
  courseName: string;
  enrollmentCount: number;
  completionRate: number;
  avgRating: number;
  revenue: number;
  totalWatchTime: number;
  activeStudents: number;
  dropoutRate: number;
  lastUpdated: string;
}

export interface StudentMetrics {
  studentId: string;
  studentName: string;
  email: string;
  enrolledCourses: number;
  completedCourses: number;
  totalWatchTime: number;
  avgCourseCompletion: number;
  lastActiveAt: string;
  engagementScore: number;
}

export interface VideoAnalytics {
  videoId: string;
  videoTitle: string;
  courseId: string;
  courseName: string;
  views: number;
  uniqueViewers: number;
  avgWatchTime: number;
  completionRate: number;
  avgQualitySelected: string;
  rewindCount: number;
  lastUpdated: string;
}

export interface SystemAlert {
  id: string;
  type: 'warning' | 'error' | 'info' | 'success';
  title: string;
  message: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  createdAt: string;
  resolved: boolean;
}

export interface RevenueMetrics {
  date: string;
  courseSales: number;
  totalRevenue: number;
  transactions: number;
  avgTransactionValue: number;
  topCourse: {
    courseId: string;
    courseName: string;
    revenue: number;
  };
}

export interface RetentionMetrics {
  month: string;
  startingUsers: number;
  churnedUsers: number;
  retentionRate: number;
  avgCoursesPerStudent: number;
}

export interface TimeSeriesData {
  timestamp: string;
  value: number;
}

export interface DimensionalData {
  label: string;
  value: number;
  percentage: number;
}

// Analytics Service
export class AnalyticsService {
  // Dashboard Metrics
  static async getDashboardMetrics(): Promise<DashboardMetrics> {
    const response = await api.get('/api/v1/analytics/dashboard');
    return response.data;
  }

  // Engagement Tracking
  static async getEngagementMetrics(
    startDate?: string,
    endDate?: string,
    interval?: 'daily' | 'weekly' | 'monthly'
  ): Promise<EngagementMetrics[]> {
    const response = await api.get('/api/v1/analytics/engagement', {
      params: {
        startDate,
        endDate,
        interval: interval || 'daily',
      },
    });
    return response.data;
  }

  static async getEngagementTrend(days: number = 30): Promise<TimeSeriesData[]> {
    const response = await api.get('/api/v1/analytics/engagement-trend', {
      params: { days },
    });
    return response.data;
  }

  // Course Performance Analytics
  static async getCoursePerformance(courseId?: string): Promise<CoursePerformance[]> {
    const response = await api.get('/api/v1/analytics/courses', {
      params: courseId ? { courseId } : {},
    });
    return response.data;
  }

  static async getTopPerformingCourses(limit: number = 10): Promise<CoursePerformance[]> {
    const response = await api.get('/api/v1/analytics/top-courses', {
      params: { limit },
    });
    return response.data;
  }

  static async getCourseCompletionRate(courseId: string): Promise<number> {
    const response = await api.get(`/api/v1/analytics/courses/${courseId}/completion-rate`);
    return response.data.rate;
  }

  static async getCourseDropoutAnalysis(courseId: string): Promise<DimensionalData[]> {
    const response = await api.get(`/api/v1/analytics/courses/${courseId}/dropout-analysis`);
    return response.data;
  }

  static async getCourseEngagementByLecture(
    courseId: string
  ): Promise<Array<{ lectureId: string; lectureName: string; engagement: number }>> {
    const response = await api.get(`/api/v1/analytics/courses/${courseId}/lecture-engagement`);
    return response.data;
  }

  // Student Analytics
  static async getStudentMetrics(studentId?: string): Promise<StudentMetrics[]> {
    const response = await api.get('/api/v1/analytics/students', {
      params: studentId ? { studentId } : {},
    });
    return response.data;
  }

  static async getTopStudents(limit: number = 10): Promise<StudentMetrics[]> {
    const response = await api.get('/api/v1/analytics/top-students', {
      params: { limit },
    });
    return response.data;
  }

  static async getStudentEngagementScore(studentId: string): Promise<number> {
    const response = await api.get(`/api/v1/analytics/students/${studentId}/engagement-score`);
    return response.data.score;
  }

  static async getStudentLearningPath(studentId: string): Promise<{
    completedCourses: Array<{ courseId: string; courseName: string; completedAt: string }>;
    inProgressCourses: Array<{ courseId: string; courseName: string; progress: number }>;
    recommendedCourses: Array<{ courseId: string; courseName: string; relevanceScore: number }>;
  }> {
    const response = await api.get(`/api/v1/analytics/students/${studentId}/learning-path`);
    return response.data;
  }

  static async getAtRiskStudents(): Promise<StudentMetrics[]> {
    const response = await api.get('/api/v1/analytics/at-risk-students');
    return response.data;
  }

  // Video Analytics
  static async getVideoAnalytics(videoId?: string): Promise<VideoAnalytics[]> {
    const response = await api.get('/api/v1/analytics/videos', {
      params: videoId ? { videoId } : {},
    });
    return response.data;
  }

  static async getTopVideos(limit: number = 10): Promise<VideoAnalytics[]> {
    const response = await api.get('/api/v1/analytics/top-videos', {
      params: { limit },
    });
    return response.data;
  }

  static async getVideoEngagement(videoId: string): Promise<{
    views: number;
    uniqueViewers: number;
    avgWatchTime: number;
    completionRate: number;
    dropoffPoints: Array<{ timestamp: number; dropoffRate: number }>;
  }> {
    const response = await api.get(`/api/v1/analytics/videos/${videoId}/engagement`);
    return response.data;
  }

  static async getQualityPreferences(): Promise<
    Array<{ quality: string; percentage: number; count: number }>
  > {
    const response = await api.get('/api/v1/analytics/quality-preferences');
    return response.data;
  }

  // Revenue Analytics
  static async getRevenueMetrics(
    startDate?: string,
    endDate?: string
  ): Promise<RevenueMetrics[]> {
    const response = await api.get('/api/v1/analytics/revenue', {
      params: { startDate, endDate },
    });
    return response.data;
  }

  static async getRevenueBySource(): Promise<
    Array<{ source: string; revenue: number; percentage: number }>
  > {
    const response = await api.get('/api/v1/analytics/revenue-sources');
    return response.data;
  }

  static async getTotalRevenue(days: number = 30): Promise<number> {
    const response = await api.get('/api/v1/analytics/total-revenue', {
      params: { days },
    });
    return response.data.total;
  }

  static async getRevenueGrowth(days: number = 30): Promise<TimeSeriesData[]> {
    const response = await api.get('/api/v1/analytics/revenue-growth', {
      params: { days },
    });
    return response.data;
  }

  // Retention Analytics
  static async getRetentionMetrics(): Promise<RetentionMetrics[]> {
    const response = await api.get('/api/v1/analytics/retention');
    return response.data;
  }

  static async getChurnRate(): Promise<number> {
    const response = await api.get('/api/v1/analytics/churn-rate');
    return response.data.rate;
  }

  static async getCohortAnalysis(): Promise<
    Array<{
      cohort: string;
      size: number;
      week1Retention: number;
      week2Retention: number;
      week4Retention: number;
    }>
  > {
    const response = await api.get('/api/v1/analytics/cohort-analysis');
    return response.data;
  }

  // System Alerts
  static async getSystemAlerts(unreadOnly: boolean = false): Promise<SystemAlert[]> {
    const response = await api.get('/api/v1/analytics/alerts', {
      params: { unreadOnly },
    });
    return response.data;
  }

  static async markAlertAsResolved(alertId: string): Promise<SystemAlert> {
    const response = await api.post(`/api/v1/analytics/alerts/${alertId}/resolve`, {});
    return response.data;
  }

  static async dismissAlert(alertId: string): Promise<void> {
    await api.post(`/api/v1/analytics/alerts/${alertId}/dismiss`, {});
  }

  // Export Data
  static async exportAnalytics(
    format: 'csv' | 'pdf' | 'json',
    reportType: 'engagement' | 'performance' | 'revenue' | 'retention'
  ): Promise<Blob> {
    const response = await api.get('/api/v1/analytics/export', {
      params: { format, reportType },
      responseType: 'blob',
    });
    return response.data;
  }

  static async generateCustomReport(filters: {
    metrics: string[];
    startDate: string;
    endDate: string;
    groupBy?: string;
  }): Promise<any> {
    const response = await api.post('/api/v1/analytics/custom-report', filters);
    return response.data;
  }

  // Real-time Data (WebSocket ready)
  static async getRealTimeMetrics(): Promise<{
    activeNow: number;
    videoWatchersNow: number;
    currentStreams: number;
    recentEnrollments: number;
  }> {
    const response = await api.get('/api/v1/analytics/realtime');
    return response.data;
  }

  // Predictive Analytics
  static async getPredictedChurn(): Promise<
    Array<{
      studentId: string;
      studentName: string;
      churnRisk: number;
      riskFactors: string[];
    }>
  > {
    const response = await api.get('/api/v1/analytics/predictions/churn');
    return response.data;
  }

  static async getPredictedGrowth(months: number = 6): Promise<TimeSeriesData[]> {
    const response = await api.get('/api/v1/analytics/predictions/growth', {
      params: { months },
    });
    return response.data;
  }

  static async getRecommendations(): Promise<
    Array<{
      type: 'content' | 'marketing' | 'retention' | 'engagement';
      priority: 'high' | 'medium' | 'low';
      title: string;
      description: string;
      estimatedImpact: string;
    }>
  > {
    const response = await api.get('/api/v1/analytics/recommendations');
    return response.data;
  }

  // Admin Analytics
  static async getSystemHealth(): Promise<{
    status: 'healthy' | 'degraded' | 'critical';
    uptime: number;
    databaseHealth: string;
    apiResponseTime: number;
    errorRate: number;
    lastCheck: string;
  }> {
    const response = await api.get('/api/v1/analytics/system-health');
    return response.data;
  }

  static async getUserDemographics(): Promise<{
    byCountry: Array<{ country: string; count: number; percentage: number }>;
    byAgeGroup: Array<{ ageGroup: string; count: number; percentage: number }>;
    byLanguage: Array<{ language: string; count: number; percentage: number }>;
  }> {
    const response = await api.get('/api/v1/analytics/demographics');
    return response.data;
  }

  static async getCourseMetadata(): Promise<{
    totalCourses: number;
    avgLecturesPerCourse: number;
    avgDurationMinutes: number;
    categoriesDistribution: Array<{ category: string; count: number }>;
    instructorCount: number;
  }> {
    const response = await api.get('/api/v1/analytics/course-metadata');
    return response.data;
  }
}
