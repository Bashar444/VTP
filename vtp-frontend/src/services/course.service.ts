import { api } from './api.client';

export interface Course {
  id: string;
  title: string;
  description: string;
  instructorId: string;
  instructorName: string;
  thumbnail: string;
  coverImage: string;
  category: string;
  level: 'beginner' | 'intermediate' | 'advanced';
  status: 'draft' | 'published' | 'archived';
  price: number;
  isFree: boolean;
  rating: number;
  reviewCount: number;
  studentCount: number;
  duration: number; // in minutes
  lectureCount: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface Lecture {
  id: string;
  courseId: string;
  title: string;
  description: string;
  duration: number; // in minutes
  videoId: string;
  videoUrl: string;
  order: number;
  isPublished: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export interface Enrollment {
  id: string;
  courseId: string;
  userId: string;
  enrolledAt: Date;
  completionPercentage: number;
  isCompleted: boolean;
  completedAt?: Date;
  lastAccessedAt: Date;
  certificateIssued: boolean;
}

export interface CourseProgress {
  courseId: string;
  userId: string;
  totalLectures: number;
  completedLectures: number;
  completionPercentage: number;
  currentLecture?: string;
  lastAccessedAt: Date;
}

export interface CourseReview {
  id: string;
  courseId: string;
  userId: string;
  userName: string;
  rating: number;
  comment: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface CourseStats {
  courseId: string;
  totalEnrollments: number;
  totalCompletions: number;
  averageRating: number;
  totalReviews: number;
  totalRevenue: number;
  averageCompletion: number;
}

class CourseServiceImpl {
  /**
   * Get all courses with optional filters
   */
  async getCourses(filters?: {
    category?: string;
    level?: string;
    search?: string;
    instructor?: string;
    limit?: number;
    offset?: number;
  }): Promise<{ courses: Course[]; total: number }> {
    const response = await api.get('/courses', {
      params: filters,
    });
    return response.data;
  }

  /**
   * Get featured courses
   */
  async getFeaturedCourses(limit: number = 6): Promise<Course[]> {
    const response = await api.get('/courses/featured', {
      params: { limit },
    });
    return response.data;
  }

  /**
   * Get course by ID
   */
  async getCourseById(courseId: string): Promise<Course> {
    const response = await api.get(`/courses/${courseId}`);
    return response.data;
  }

  /**
   * Get course lectures
   */
  async getCourseLectures(courseId: string): Promise<Lecture[]> {
    const response = await api.get(`/courses/${courseId}/lectures`);
    return response.data;
  }

  /**
   * Get course by instructor
   */
  async getInstructorCourses(instructorId: string): Promise<Course[]> {
    const response = await api.get(`/instructors/${instructorId}/courses`);
    return response.data;
  }

  /**
   * Get user's enrolled courses (uses current authenticated user)
   */
  async getEnrolledCourses(): Promise<Enrollment[]> {
    const response = await api.get(`/api/v1/courses/my-enrollments`);
    return response.data.enrollments || [];
  }

  /**
   * Get user's completed courses (uses current authenticated user)
   */
  async getCompletedCourses(): Promise<Course[]> {
    // TODO: Backend needs to implement completed courses endpoint
    return [];
  }

  /**
   * Enroll in a course
   */
  async enrollCourse(courseId: string): Promise<Enrollment> {
    const response = await api.post(`/courses/${courseId}/enroll`, {
      enrolledAt: new Date(),
    });
    return response.data;
  }

  /**
   * Unenroll from a course
   */
  async unenrollCourse(courseId: string): Promise<void> {
    await api.post(`/courses/${courseId}/unenroll`, {});
  }

  /**
   * Get course progress
   */
  async getCourseProgress(courseId: string): Promise<CourseProgress> {
    const response = await api.get(`/courses/${courseId}/progress`);
    return response.data;
  }

  /**
   * Update lecture progress
   */
  async updateLectureProgress(courseId: string, lectureId: string): Promise<CourseProgress> {
    const response = await api.post(`/courses/${courseId}/lectures/${lectureId}/progress`, {
      completedAt: new Date(),
    });
    return response.data;
  }

  /**
   * Mark lecture as completed
   */
  async markLectureComplete(courseId: string, lectureId: string): Promise<void> {
    await api.post(`/courses/${courseId}/lectures/${lectureId}/complete`, {
      completedAt: new Date(),
    });
  }

  /**
   * Mark course as completed
   */
  async markCourseComplete(courseId: string): Promise<Enrollment> {
    const response = await api.post(`/courses/${courseId}/complete`, {
      completedAt: new Date(),
    });
    return response.data;
  }

  /**
   * Get course reviews
   */
  async getCourseReviews(courseId: string): Promise<CourseReview[]> {
    const response = await api.get(`/courses/${courseId}/reviews`);
    return response.data;
  }

  /**
   * Create course review
   */
  async createReview(
    courseId: string,
    rating: number,
    comment: string
  ): Promise<CourseReview> {
    const response = await api.post(`/courses/${courseId}/reviews`, {
      rating,
      comment,
      createdAt: new Date(),
    });
    return response.data;
  }

  /**
   * Get course statistics (instructor only)
   */
  async getCourseStats(courseId: string): Promise<CourseStats> {
    const response = await api.get(`/courses/${courseId}/stats`);
    return response.data;
  }

  /**
   * Create a new course (instructor only)
   */
  async createCourse(courseData: Partial<Course>): Promise<Course> {
    const response = await api.post('/courses', {
      ...courseData,
      createdAt: new Date(),
    });
    return response.data;
  }

  /**
   * Update course (instructor only)
   */
  async updateCourse(courseId: string, courseData: Partial<Course>): Promise<Course> {
    const response = await api.put(`/courses/${courseId}`, {
      ...courseData,
      updatedAt: new Date(),
    });
    return response.data;
  }

  /**
   * Delete course (instructor only)
   */
  async deleteCourse(courseId: string): Promise<void> {
    await api.delete(`/courses/${courseId}`);
  }

  /**
   * Create lecture (instructor only)
   */
  async createLecture(courseId: string, lectureData: Partial<Lecture>): Promise<Lecture> {
    const response = await api.post(`/courses/${courseId}/lectures`, {
      ...lectureData,
      createdAt: new Date(),
    });
    return response.data;
  }

  /**
   * Update lecture (instructor only)
   */
  async updateLecture(
    courseId: string,
    lectureId: string,
    lectureData: Partial<Lecture>
  ): Promise<Lecture> {
    const response = await api.put(`/courses/${courseId}/lectures/${lectureId}`, {
      ...lectureData,
      updatedAt: new Date(),
    });
    return response.data;
  }

  /**
   * Delete lecture (instructor only)
   */
  async deleteLecture(courseId: string, lectureId: string): Promise<void> {
    await api.delete(`/courses/${courseId}/lectures/${lectureId}`);
  }

  /**
   * Reorder lectures (instructor only)
   */
  async reorderLectures(
    courseId: string,
    lectureIds: string[]
  ): Promise<Lecture[]> {
    const response = await api.post(`/courses/${courseId}/lectures/reorder`, {
      lectureIds,
    });
    return response.data;
  }

  /**
   * Search courses
   */
  async searchCourses(query: string, limit: number = 10): Promise<Course[]> {
    const response = await api.get('/courses/search', {
      params: { q: query, limit },
    });
    return response.data;
  }

  /**
   * Get recommended courses based on user's learning history
   */
  async getRecommendedCourses(limit: number = 5): Promise<Course[]> {
    const response = await api.get('/courses/recommended', {
      params: { limit },
    });
    return response.data;
  }

  /**
   * Get courses by category
   */
  async getCoursesByCategory(category: string): Promise<Course[]> {
    const response = await api.get(`/courses/category/${category}`);
    return response.data;
  }

  /**
   * Get trending courses
   */
  async getTrendingCourses(limit: number = 6): Promise<Course[]> {
    const response = await api.get('/courses/trending', {
      params: { limit },
    });
    return response.data;
  }

  /**
   * Get course certificate (if completed)
   */
  async getCourseCertificate(courseId: string): Promise<{ url: string; issuedAt: Date }> {
    const response = await api.get(`/courses/${courseId}/certificate`);
    return response.data;
  }

  /**
   * Check if user is enrolled in course
   */
  async isEnrolled(courseId: string): Promise<boolean> {
    const response = await api.get(`/courses/${courseId}/is-enrolled`);
    return response.data.enrolled;
  }
}
// Export both named and default for flexibility
export const CourseService = new CourseServiceImpl();
export default CourseService;
