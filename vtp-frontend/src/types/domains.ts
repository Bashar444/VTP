// Instructor domain types
export interface Instructor {
  id: string;
  user_id: string;
  name_ar: string;
  bio_ar: string;
  specialization: string[];
  hourly_rate: number;
  rating: number;
  total_reviews: number;
  years_experience: number;
  certifications_ar: string[];
  availability: Record<string, string[]>;
  is_verified: boolean;
  is_active: boolean;
  profile_image_url?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateInstructorDTO {
  user_id: string;
  name_ar: string;
  bio_ar: string;
  specialization: string[];
  hourly_rate: number;
  years_experience: number;
  certifications_ar: string[];
  availability: Record<string, string[]>;
  profile_image_url?: string;
}

export interface UpdateInstructorDTO {
  name_ar: string;
  bio_ar: string;
  specialization: string[];
  hourly_rate: number;
  years_experience: number;
  certifications_ar: string[];
  availability: Record<string, string[]>;
  profile_image_url?: string;
}

export interface InstructorFilters {
  subject_id?: string;
  min_rating?: number;
  is_verified?: boolean;
  page?: number;
  page_size?: number;
}

// Subject domain types
export interface Subject {
  id: string;
  name_ar: string;
  level: 'elementary' | 'intermediate' | 'advanced';
  category: string;
  created_at: string;
}

export interface CreateSubjectDTO {
  name_ar: string;
  level: 'elementary' | 'intermediate' | 'advanced';
  category: string;
}

// Meeting domain types
export interface Meeting {
  id: string;
  instructor_id: string;
  student_id?: string;
  subject_id: string;
  title_ar: string;
  scheduled_at: string;
  duration: number;
  meeting_url?: string;
  room_id?: string;
  status: 'scheduled' | 'completed' | 'cancelled' | 'no-show';
  end_time?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateMeetingDTO {
  instructor_id: string;
  student_id?: string;
  subject_id: string;
  title_ar: string;
  scheduled_at: string;
  duration: number;
  meeting_url?: string;
  room_id?: string;
}

export interface UpdateMeetingDTO {
  title_ar: string;
  scheduled_at: string;
  duration: number;
  meeting_url?: string;
  room_id?: string;
  status?: string;
}

export interface MeetingFilters {
  instructor_id?: string;
  student_id?: string;
  subject_id?: string;
  status?: string;
  page?: number;
  page_size?: number;
}

// Study Material domain types
export interface StudyMaterial {
  id: string;
  course_id?: string;
  instructor_id: string;
  title_ar: string;
  type: 'pdf' | 'slides' | 'notes' | 'worksheet' | 'video' | 'audio';
  file_url: string;
  file_size: number;
  downloads: number;
  created_at: string;
  updated_at: string;
}

// Assignments domain types
export interface Assignment {
  id: string;
  course_id?: string;
  instructor_id: string;
  title_ar: string;
  description_ar?: string;
  subject_id?: string;
  due_at: string;
  max_points: number;
  created_at: string;
  updated_at: string;
}

export interface CreateAssignmentDTO {
  course_id?: string;
  instructor_id: string;
  title_ar: string;
  description_ar?: string;
  subject_id?: string;
  due_at: string;
  max_points?: number;
}

export interface AssignmentSubmission {
  id: string;
  assignment_id: string;
  student_id: string;
  submitted_at: string;
  file_url?: string;
  notes?: string;
  grade?: number;
  graded_at?: string;
  feedback_ar?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateSubmissionDTO {
  assignment_id: string;
  student_id: string;
  file_url?: string;
  notes?: string;
}


export interface CreateMaterialDTO {
  course_id?: string;
  instructor_id: string;
  title_ar: string;
  type: 'pdf' | 'slides' | 'notes' | 'worksheet' | 'video' | 'audio';
  file_url: string;
  file_size: number;
}

export interface UpdateMaterialDTO {
  title_ar: string;
  type: 'pdf' | 'slides' | 'notes' | 'worksheet' | 'video' | 'audio';
  file_url: string;
  file_size: number;
}

export interface MaterialFilters {
  course_id?: string;
  instructor_id?: string;
  type?: string;
  page?: number;
  page_size?: number;
}
