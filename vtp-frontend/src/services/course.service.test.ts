import { describe, it, expect, beforeEach, vi } from 'vitest';
import { CourseService } from '@/services/course.service';
import { api } from '@/services/api.client';

vi.mock('@/services/api.client');

const mockCourse = {
  id: 'course-1',
  title: 'JavaScript Basics',
  description: 'Learn JavaScript from scratch',
  instructor: 'John Doe',
  thumbnail: 'https://example.com/thumbnail.jpg',
  category: 'programming',
  level: 'beginner',
  status: 'published' as const,
  price: 29.99,
  rating: 4.5,
  students: 1000,
  duration: 10,
  lectures: 25,
};

const mockLecture = {
  id: 'lecture-1',
  courseId: 'course-1',
  title: 'Introduction to Variables',
  duration: 45,
  videoId: 'video-1',
  order: 1,
};

describe('CourseService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getCourses', () => {
    it('should fetch all courses', async () => {
      const mockCourses = [mockCourse];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockCourses });

      const result = await CourseService.getCourses();

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses');
      expect(result).toEqual(mockCourses);
    });
  });

  describe('getFeaturedCourses', () => {
    it('should fetch featured courses', async () => {
      const mockFeatured = [mockCourse];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockFeatured });

      const result = await CourseService.getFeaturedCourses();

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/featured');
      expect(result).toEqual(mockFeatured);
    });
  });

  describe('getCourseById', () => {
    it('should fetch a course by ID', async () => {
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockCourse });

      const result = await CourseService.getCourseById('course-1');

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/course-1');
      expect(result).toEqual(mockCourse);
    });
  });

  describe('getCourseLectures', () => {
    it('should fetch lectures for a course', async () => {
      const mockLectures = [mockLecture];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockLectures });

      const result = await CourseService.getCourseLectures('course-1');

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/course-1/lectures');
      expect(result).toEqual(mockLectures);
    });
  });

  describe('getEnrolledCourses', () => {
    it('should fetch enrolled courses for the user', async () => {
      const mockCourses = [{ ...mockCourse, progress: 50 }];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockCourses });

      const result = await CourseService.getEnrolledCourses();

      expect(api.get).toHaveBeenCalledWith('/api/v1/users/enrolled-courses');
      expect(result).toEqual(mockCourses);
    });
  });

  describe('getCompletedCourses', () => {
    it('should fetch completed courses for the user', async () => {
      const mockCourses = [mockCourse];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockCourses });

      const result = await CourseService.getCompletedCourses();

      expect(api.get).toHaveBeenCalledWith('/api/v1/users/completed-courses');
      expect(result).toEqual(mockCourses);
    });
  });

  describe('enrollCourse', () => {
    it('should enroll user in a course', async () => {
      vi.mocked(api.post).mockResolvedValueOnce({ data: { success: true } });

      const result = await CourseService.enrollCourse('course-1');

      expect(api.post).toHaveBeenCalledWith('/api/v1/courses/course-1/enroll', {});
      expect(result).toEqual({ success: true });
    });
  });

  describe('unenrollCourse', () => {
    it('should unenroll user from a course', async () => {
      vi.mocked(api.post).mockResolvedValueOnce({ data: { success: true } });

      const result = await CourseService.unenrollCourse('course-1');

      expect(api.post).toHaveBeenCalledWith('/api/v1/courses/course-1/unenroll', {});
      expect(result).toEqual({ success: true });
    });
  });

  describe('getCourseProgress', () => {
    it('should fetch course progress for the user', async () => {
      const mockProgress = {
        courseId: 'course-1',
        completedLectures: 10,
        totalLectures: 25,
        percentageComplete: 40,
      };
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockProgress });

      const result = await CourseService.getCourseProgress('course-1');

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/course-1/progress');
      expect(result).toEqual(mockProgress);
    });
  });

  describe('updateLectureProgress', () => {
    it('should update progress for a lecture', async () => {
      vi.mocked(api.post).mockResolvedValueOnce({ data: { success: true } });

      const result = await CourseService.updateLectureProgress('lecture-1', {
        watched: true,
      });

      expect(api.post).toHaveBeenCalledWith('/api/v1/lectures/lecture-1/progress', {
        watched: true,
      });
      expect(result).toEqual({ success: true });
    });
  });

  describe('markLectureComplete', () => {
    it('should mark a lecture as complete', async () => {
      vi.mocked(api.post).mockResolvedValueOnce({ data: { success: true } });

      const result = await CourseService.markLectureComplete('lecture-1');

      expect(api.post).toHaveBeenCalledWith('/api/v1/lectures/lecture-1/complete', {});
      expect(result).toEqual({ success: true });
    });
  });

  describe('markCourseComplete', () => {
    it('should mark a course as complete', async () => {
      vi.mocked(api.post).mockResolvedValueOnce({ data: { success: true } });

      const result = await CourseService.markCourseComplete('course-1');

      expect(api.post).toHaveBeenCalledWith('/api/v1/courses/course-1/complete', {});
      expect(result).toEqual({ success: true });
    });
  });

  describe('getCourseReviews', () => {
    it('should fetch reviews for a course', async () => {
      const mockReviews = [
        {
          id: 'review-1',
          courseId: 'course-1',
          userId: 'user-1',
          rating: 5,
          comment: 'Great course!',
          createdAt: new Date().toISOString(),
        },
      ];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockReviews });

      const result = await CourseService.getCourseReviews('course-1');

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/course-1/reviews');
      expect(result).toEqual(mockReviews);
    });
  });

  describe('createReview', () => {
    it('should create a review for a course', async () => {
      const mockReview = {
        id: 'review-1',
        courseId: 'course-1',
        userId: 'user-1',
        rating: 5,
        comment: 'Great course!',
        createdAt: new Date().toISOString(),
      };
      vi.mocked(api.post).mockResolvedValueOnce({ data: mockReview });

      const result = await CourseService.createReview('course-1', {
        rating: 5,
        comment: 'Great course!',
      });

      expect(api.post).toHaveBeenCalledWith('/api/v1/courses/course-1/reviews', {
        rating: 5,
        comment: 'Great course!',
      });
      expect(result).toEqual(mockReview);
    });
  });

  describe('getCourseStats', () => {
    it('should fetch course statistics', async () => {
      const mockStats = {
        courseId: 'course-1',
        enrollmentCount: 1000,
        completionRate: 45,
        averageRating: 4.5,
      };
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockStats });

      const result = await CourseService.getCourseStats('course-1');

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/course-1/stats');
      expect(result).toEqual(mockStats);
    });
  });

  describe('searchCourses', () => {
    it('should search for courses', async () => {
      const mockResults = [mockCourse];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockResults });

      const result = await CourseService.searchCourses('JavaScript');

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/search', {
        params: { q: 'JavaScript' },
      });
      expect(result).toEqual(mockResults);
    });
  });

  describe('getRecommendedCourses', () => {
    it('should fetch recommended courses', async () => {
      const mockCourses = [mockCourse];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockCourses });

      const result = await CourseService.getRecommendedCourses();

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/recommended');
      expect(result).toEqual(mockCourses);
    });
  });

  describe('getCoursesByCategory', () => {
    it('should fetch courses by category', async () => {
      const mockCourses = [mockCourse];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockCourses });

      const result = await CourseService.getCoursesByCategory('programming');

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/category/programming');
      expect(result).toEqual(mockCourses);
    });
  });

  describe('getTrendingCourses', () => {
    it('should fetch trending courses', async () => {
      const mockCourses = [mockCourse];
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockCourses });

      const result = await CourseService.getTrendingCourses();

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/trending');
      expect(result).toEqual(mockCourses);
    });
  });

  describe('getCourseCertificate', () => {
    it('should fetch course certificate', async () => {
      const mockCert = {
        id: 'cert-1',
        courseId: 'course-1',
        userId: 'user-1',
        issuedAt: new Date().toISOString(),
        certificateUrl: 'https://example.com/cert.pdf',
      };
      vi.mocked(api.get).mockResolvedValueOnce({ data: mockCert });

      const result = await CourseService.getCourseCertificate('course-1');

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/course-1/certificate');
      expect(result).toEqual(mockCert);
    });
  });

  describe('isEnrolled', () => {
    it('should check if user is enrolled in course', async () => {
      vi.mocked(api.get).mockResolvedValueOnce({ data: { enrolled: true } });

      const result = await CourseService.isEnrolled('course-1');

      expect(api.get).toHaveBeenCalledWith('/api/v1/courses/course-1/enrollment-status');
      expect(result).toEqual(true);
    });
  });

  describe('createCourse (Instructor)', () => {
    it('should create a new course', async () => {
      vi.mocked(api.post).mockResolvedValueOnce({ data: mockCourse });

      const result = await CourseService.createCourse({
        title: 'JavaScript Basics',
        description: 'Learn JavaScript from scratch',
        category: 'programming',
      });

      expect(api.post).toHaveBeenCalledWith('/api/v1/instructor/courses', {
        title: 'JavaScript Basics',
        description: 'Learn JavaScript from scratch',
        category: 'programming',
      });
      expect(result).toEqual(mockCourse);
    });
  });

  describe('updateCourse (Instructor)', () => {
    it('should update a course', async () => {
      vi.mocked(api.put).mockResolvedValueOnce({ data: mockCourse });

      const result = await CourseService.updateCourse('course-1', {
        title: 'Advanced JavaScript',
      });

      expect(api.put).toHaveBeenCalledWith('/api/v1/instructor/courses/course-1', {
        title: 'Advanced JavaScript',
      });
      expect(result).toEqual(mockCourse);
    });
  });

  describe('deleteCourse (Instructor)', () => {
    it('should delete a course', async () => {
      vi.mocked(api.delete).mockResolvedValueOnce({ data: { success: true } });

      const result = await CourseService.deleteCourse('course-1');

      expect(api.delete).toHaveBeenCalledWith('/api/v1/instructor/courses/course-1');
      expect(result).toEqual({ success: true });
    });
  });
});
