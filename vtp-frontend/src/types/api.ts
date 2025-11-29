// API Types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  refreshToken: string;
  user: User;
}

export interface RegisterRequest {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  role: 'student' | 'instructor' | 'admin';
}

export interface User {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  role: 'student' | 'instructor' | 'admin';
  profilePictureUrl?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Course {
  id: string;
  title: string;
  description: string;
  instructorId: string;
  instructor?: User;
  thumbnailUrl?: string;
  category: string;
  enrollmentCount: number;
  isPublished: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Lecture {
  id: string;
  courseId: string;
  title: string;
  description?: string;
  lectureNumber: number;
  mediaURL: string;
  durationSeconds: number;
  mediaType: 'HLS' | 'DASH' | 'MP4';
  isPublished: boolean;
  publishedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface StreamingSession {
  id: string;
  lectureId: string;
  instructorId: string;
  status: 'scheduled' | 'live' | 'ended';
  startTime: string;
  endTime?: string;
  participantCount: number;
  peerId?: string;
  roomId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Enrollment {
  id: string;
  userId: string;
  courseId: string;
  enrolledAt: string;
  completionPercentage: number;
  completedLectures: number;
  totalLectures: number;
}

export interface AnalyticsEvent {
  id: string;
  userId: string;
  lectureId: string;
  eventType:
    | 'PlaybackStarted'
    | 'PlaybackStopped'
    | 'QualityChanged'
    | 'BufferEvent'
    | 'Engagement';
  metadata: Record<string, any>;
  timestamp: string;
}

export interface StreamingMetrics {
  lectureId: string;
  totalViews: number;
  averageWatchTime: number;
  completionRate: number;
  bufferingEvents: number;
  averageBitrate: number;
  qualitySwitches: number;
}

export interface Alert {
  id: string;
  userId?: string;
  type: 'warning' | 'error' | 'info' | 'success';
  title: string;
  message: string;
  read: boolean;
  createdAt: string;
}
