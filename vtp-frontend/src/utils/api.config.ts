// API Client Configuration
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const API_ENDPOINTS = {
  // Auth endpoints
  auth: {
    login: '/api/v1/auth/login',
    register: '/api/v1/auth/register',
    logout: '/api/v1/auth/logout',
    refresh: '/api/v1/auth/refresh',
    profile: '/api/v1/auth/profile',
    verifyEmail: '/api/v1/auth/verify-email',
    resendVerification: '/api/v1/auth/resend-verification',
    resetPassword: '/api/v1/auth/reset-password',
    changePassword: '/api/v1/auth/change-password',
  },
  // Course endpoints
  courses: {
    list: '/api/v1/courses',
    create: '/api/v1/courses',
    getById: (id: string) => `/api/v1/courses/${id}`,
    update: (id: string) => `/api/v1/courses/${id}`,
    delete: (id: string) => `/api/v1/courses/${id}`,
  },
  // Lecture endpoints
  lectures: {
    list: (courseId: string) => `/api/v1/courses/${courseId}/lectures`,
    create: (courseId: string) => `/api/v1/courses/${courseId}/lectures`,
    getById: (courseId: string, lectureId: string) =>
      `/api/v1/courses/${courseId}/lectures/${lectureId}`,
    update: (courseId: string, lectureId: string) =>
      `/api/v1/courses/${courseId}/lectures/${lectureId}`,
    delete: (courseId: string, lectureId: string) =>
      `/api/v1/courses/${courseId}/lectures/${lectureId}`,
  },
  // Streaming endpoints
  streaming: {
    sessions: '/api/v1/streaming/sessions',
    startSession: '/api/v1/streaming/sessions/start',
    endSession: (sessionId: string) =>
      `/api/v1/streaming/sessions/${sessionId}/end`,
    getSession: (sessionId: string) => `/api/v1/streaming/sessions/${sessionId}`,
  },
  // Analytics endpoints
  analytics: {
    events: '/api/v1/analytics/events',
    createEvent: '/api/v1/analytics/events',
    metrics: '/api/v1/analytics/metrics',
    getLectureMetrics: (lectureId: string) =>
      `/api/v1/analytics/metrics/${lectureId}`,
  },
};

// Timeout in milliseconds
export const API_TIMEOUT = 30000;

// Retry configuration
export const RETRY_CONFIG = {
  maxRetries: 3,
  retryDelay: 1000,
  retryableStatuses: [408, 429, 500, 502, 503, 504],
};
