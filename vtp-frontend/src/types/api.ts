// API Types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  user: User;
}

export interface RegisterRequest {
  full_name: string;
  email: string;
  password: string;
  role: 'student' | 'instructor' | 'admin';
  phone?: string;
}

export interface User {
  user_id: string;
  email: string;
  full_name: string;
  phone?: string;
  role: 'student' | 'instructor' | 'admin';
  profile_picture_url?: string;
  created_at?: string;
  updated_at?: string;
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
