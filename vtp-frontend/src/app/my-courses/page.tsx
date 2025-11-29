'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/store/auth';
import { CourseService } from '@/services/course.service';
import { CourseCard } from '@/components/courses/CourseCard';
import type { Course } from '@/services/course.service';
import { Play, BookOpen, CheckCircle } from 'lucide-react';

interface EnrolledCourse extends Course {
  progress?: number;
  lastWatchedDate?: string;
}

export default function MyCoursesPage() {
  const router = useRouter();
  const { user } = useAuth();
  const [enrolledCourses, setEnrolledCourses] = useState<EnrolledCourse[]>([]);
  const [completedCourses, setCompletedCourses] = useState<Course[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'in-progress' | 'completed'>('in-progress');

  useEffect(() => {
    // Redirect if not authenticated
    if (!user) {
      router.push('/auth/login');
      return;
    }

    const fetchCourses = async () => {
      try {
        setIsLoading(true);
        const [enrolled, completed] = await Promise.all([
          CourseService.getEnrolledCourses(),
          CourseService.getCompletedCourses(),
        ]);

        setEnrolledCourses(enrolled);
        setCompletedCourses(completed);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load courses');
      } finally {
        setIsLoading(false);
      }
    };

    fetchCourses();
  }, [user, router]);

  const handleContinueWatching = (courseId: string) => {
    // In a real app, this would fetch the last watched video
    router.push(`/courses/${courseId}`);
  };

  const handleViewCourse = (courseId: string) => {
    router.push(`/courses/${courseId}`);
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-900 pt-24 pb-12">
        <div className="container mx-auto px-4">
          <h1 className="text-4xl font-bold text-white mb-8">My Courses</h1>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="bg-gray-800 rounded-lg animate-pulse h-80" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        {/* Page Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-white mb-2">My Courses</h1>
          <p className="text-gray-400">
            {enrolledCourses.length + completedCourses.length} courses total
          </p>
        </div>

        {error && (
          <div className="bg-red-900/20 border border-red-700 rounded-lg p-4 mb-6">
            <p className="text-red-400">{error}</p>
          </div>
        )}

        {/* Tabs */}
        <div className="mb-8 border-b border-gray-700">
          <div className="flex gap-8">
            <button
              onClick={() => setActiveTab('in-progress')}
              className={`py-4 px-2 font-semibold border-b-2 transition-colors ${
                activeTab === 'in-progress'
                  ? 'text-blue-400 border-blue-400'
                  : 'text-gray-400 border-transparent hover:text-white'
              }`}
            >
              <div className="flex items-center gap-2">
                <Play className="w-5 h-5" />
                In Progress ({enrolledCourses.length})
              </div>
            </button>
            <button
              onClick={() => setActiveTab('completed')}
              className={`py-4 px-2 font-semibold border-b-2 transition-colors ${
                activeTab === 'completed'
                  ? 'text-blue-400 border-blue-400'
                  : 'text-gray-400 border-transparent hover:text-white'
              }`}
            >
              <div className="flex items-center gap-2">
                <CheckCircle className="w-5 h-5" />
                Completed ({completedCourses.length})
              </div>
            </button>
          </div>
        </div>

        {/* In Progress Courses */}
        {activeTab === 'in-progress' && (
          <div>
            {enrolledCourses.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {enrolledCourses.map(course => (
                  <div
                    key={course.id}
                    className="bg-gray-800 rounded-lg overflow-hidden hover:ring-2 hover:ring-blue-500 transition-all"
                  >
                    {/* Course Card */}
                    <div className="relative h-48 bg-gray-700 overflow-hidden group">
                      {course.thumbnail && (
                        <img
                          src={course.thumbnail}
                          alt={course.title}
                          className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-300"
                        />
                      )}
                      <div className="absolute inset-0 bg-black/40 group-hover:bg-black/60 transition-colors" />
                      <button
                        onClick={() => handleContinueWatching(course.id)}
                        className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                      >
                        <div className="w-14 h-14 rounded-full bg-blue-600 flex items-center justify-center">
                          <Play className="w-6 h-6 text-white fill-white ml-1" />
                        </div>
                      </button>
                    </div>

                    {/* Course Info */}
                    <div className="p-4">
                      <h3 className="text-lg font-bold text-white mb-2 line-clamp-2">
                        {course.title}
                      </h3>

                      {/* Progress Bar */}
                      <div className="mb-4">
                        <div className="flex justify-between items-center mb-2">
                          <span className="text-sm text-gray-400">Progress</span>
                          <span className="text-sm font-semibold text-blue-400">
                            {Math.round(course.progress || 0)}%
                          </span>
                        </div>
                        <div className="w-full bg-gray-700 rounded-full h-2">
                          <div
                            className="bg-gradient-to-r from-blue-500 to-purple-600 h-2 rounded-full transition-all"
                            style={{ width: `${course.progress || 0}%` }}
                          />
                        </div>
                      </div>

                      {/* Footer */}
                      <div className="flex items-center justify-between pt-4 border-t border-gray-700">
                        <span className="text-sm text-gray-400">
                          {course.lastWatchedDate ? `Last watched: ${course.lastWatchedDate}` : 'Not started'}
                        </span>
                        <button
                          onClick={() => handleContinueWatching(course.id)}
                          className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded transition-colors"
                        >
                          Continue
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-16">
                <BookOpen className="w-16 h-16 text-gray-600 mx-auto mb-4" />
                <h3 className="text-xl font-semibold text-gray-300 mb-2">No courses yet</h3>
                <p className="text-gray-400 mb-6">Start learning by enrolling in a course</p>
                <button
                  onClick={() => router.push('/courses')}
                  className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-lg transition-colors"
                >
                  Browse Courses
                </button>
              </div>
            )}
          </div>
        )}

        {/* Completed Courses */}
        {activeTab === 'completed' && (
          <div>
            {completedCourses.length > 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {completedCourses.map(course => (
                  <div
                    key={course.id}
                    className="bg-gray-800 rounded-lg overflow-hidden hover:ring-2 hover:ring-green-500 transition-all"
                  >
                    {/* Course Card */}
                    <div className="relative h-48 bg-gray-700 overflow-hidden group">
                      {course.thumbnail && (
                        <img
                          src={course.thumbnail}
                          alt={course.title}
                          className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-300"
                        />
                      )}
                      <div className="absolute inset-0 bg-black/40 group-hover:bg-black/60 transition-colors" />
                      <div className="absolute top-3 right-3 bg-green-600 text-white px-3 py-1 rounded-full text-sm font-semibold flex items-center gap-1">
                        <CheckCircle className="w-4 h-4" />
                        Completed
                      </div>
                    </div>

                    {/* Course Info */}
                    <div className="p-4">
                      <h3 className="text-lg font-bold text-white mb-2 line-clamp-2">
                        {course.title}
                      </h3>

                      {/* Instructor */}
                      <p className="text-sm text-gray-400 mb-4">
                        by {course.instructor || 'Unknown Instructor'}
                      </p>

                      {/* Footer */}
                      <button
                        onClick={() => handleViewCourse(course.id)}
                        className="w-full py-2 px-4 bg-green-600 hover:bg-green-700 text-white font-semibold rounded transition-colors"
                      >
                        View Certificate
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-16">
                <CheckCircle className="w-16 h-16 text-gray-600 mx-auto mb-4" />
                <h3 className="text-xl font-semibold text-gray-300 mb-2">No completed courses</h3>
                <p className="text-gray-400 mb-6">Complete a course to see it here</p>
                <button
                  onClick={() => router.push('/courses')}
                  className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-lg transition-colors"
                >
                  Browse Courses
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
