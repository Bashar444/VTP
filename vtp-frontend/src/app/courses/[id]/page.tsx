'use client';
export const dynamic = 'force-dynamic';

import { useState, useEffect } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/store';
import { CourseService } from '@/services/course.service';
import { useUnenrollCourse } from '@/hooks/useUnenrollCourse';
import { CourseDetail } from '@/components/courses/CourseDetail';
import { EnrollmentForm } from '@/components/courses/CourseFilters';
import type { Course, Lecture } from '@/services/course.service';
import { AlertCircle, ArrowLeft } from 'lucide-react';

export default function CourseDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();
  const courseId = params.id as string;

  const [course, setCourse] = useState<Course | null>(null);
  const [lectures, setLectures] = useState<Lecture[]>([]);
  const [isEnrolled, setIsEnrolled] = useState(false);
  const [progress, setProgress] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showEnrollForm, setShowEnrollForm] = useState(false);
  const [isEnrolling, setIsEnrolling] = useState(false);

  const queryClient = useQueryClient();
  // Unenroll mutation & handler
  const unenrollMutation = useUnenrollCourse(courseId);
  const handleUnenroll = () => {
    unenrollMutation.mutate(undefined, {
      onSuccess: () => {
        setIsEnrolled(false);
        setProgress(0);
      },
      onError: (err: any) => {
        setError(err?.message || 'Failed to unenroll');
      },
    mutationKey: ['enroll', courseId],
    mutationFn: async () => {
      const enrollment = await CourseService.enrollCourse(courseId);
      return enrollment;
    },
    onMutate: async () => {
      setIsEnrolling(true);
      setError(null);
    },
    onSuccess: async () => {
      setIsEnrolled(true);
      setShowEnrollForm(false);
      setProgress(0);
      // Invalidate courses lists and featured caches if present
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['courses'] }),
        queryClient.invalidateQueries({ queryKey: ['featured-courses'] }),
      ]);
    },
    onError: (err: any) => {
      setError(err?.message || 'Failed to enroll in course');
    },
    });
  };

  const enrollMutation = useMutation({
    mutationKey: ['enroll', courseId],
    mutationFn: async () => {
      const enrollment = await CourseService.enrollCourse(courseId);
      return enrollment;
    },
    onMutate: async () => {
      setIsEnrolling(true);
      setError(null);
    },
    onSuccess: async () => {
      setIsEnrolled(true);
      setShowEnrollForm(false);
      setProgress(0);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['courses'] }),
        queryClient.invalidateQueries({ queryKey: ['featured-courses'] }),
      ]);
    },
    onError: (err: any) => {
      setError(err?.message || 'Failed to enroll in course');
    },
    onSettled: () => {
      setIsEnrolling(false);
    },
  });

  // Fetch course and lectures
  useEffect(() => {
    const fetchCourseData = async () => {
      try {
        setIsLoading(true);
        const [courseData, lecturesData] = await Promise.all([
          CourseService.getCourseById(courseId),
          CourseService.getCourseLectures(courseId),
        ]);

        // Ensure plain objects for client components
        setCourse(JSON.parse(JSON.stringify(courseData)));
        setLectures(JSON.parse(JSON.stringify(lecturesData)));

        // Check if user is enrolled and get progress
        if (user) {
          const enrolled = await CourseService.isEnrolled(courseId);
          setIsEnrolled(enrolled);

          if (enrolled) {
            const progressData = await CourseService.getCourseProgress(courseId);
            setProgress(progressData.percentageComplete || 0);
          }
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load course');
      } finally {
        setIsLoading(false);
      }
    };

    if (courseId) {
      fetchCourseData();
    }
  }, [courseId, user]);

  const handleEnroll = () => {
    if (!user) {
      router.push('/auth/login');
      return;
    }
    enrollMutation.mutate();
  };

  const handleSelectLecture = (lectureId: string) => {
    const lecture = lectures.find(l => l.id === lectureId);
    if (lecture?.videoId) {
      router.push(`/watch/${lecture.videoId}`);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-900 pt-24 pb-12">
        <div className="container mx-auto px-4">
          <div className="animate-pulse space-y-8">
            <div className="h-80 bg-gray-800 rounded-lg" />
            <div className="space-y-4">
              <div className="h-8 bg-gray-800 rounded w-3/4" />
              <div className="h-4 bg-gray-800 rounded w-full" />
              <div className="h-4 bg-gray-800 rounded w-5/6" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !course) {
    return (
      <div className="min-h-screen bg-gray-900 pt-24 pb-12">
        <div className="container mx-auto px-4">
          <button
            onClick={() => router.back()}
            className="mb-6 flex items-center gap-2 text-blue-400 hover:text-blue-300"
          >
            <ArrowLeft className="w-5 h-5" />
            Back
          </button>

          <div className="bg-red-900/20 border border-red-700 rounded-lg p-6 flex items-center gap-4">
            <AlertCircle className="w-6 h-6 text-red-400 flex-shrink-0" />
            <p className="text-red-400">{error || 'Course not found'}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        {/* Back Button */}
        <button
          onClick={() => router.back()}
          className="mb-6 flex items-center gap-2 text-blue-400 hover:text-blue-300 transition-colors"
        >
          <ArrowLeft className="w-5 h-5" />
          Back to Courses
        </button>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Course Details - 2 columns on desktop */}
          <div className="lg:col-span-2">
            <CourseDetail
              course={course}
              lectures={lectures}
              progress={isEnrolled ? progress : undefined}
              isEnrolled={isEnrolled}
              onEnroll={() => setShowEnrollForm(!showEnrollForm)}
              onSelectLecture={handleSelectLecture}
            />
          </div>

          {/* Sidebar - Enrollment or Info */}
          <div className="lg:col-span-1">
            <div className="sticky top-24 space-y-6">
              {/* Enrollment Form or Summary */}
              {!isEnrolled ? (
                showEnrollForm ? (
                  <EnrollmentForm
                    courseId={courseId}
                    courseName={course.title}
                    coursePrice={course.price || 0}
                    isFree={!course.price}
                    onEnroll={handleEnroll}
                    onCancel={() => setShowEnrollForm(false)}
                    isLoading={isEnrolling}
                  />
                ) : (
                  <div className="bg-gray-800 rounded-lg p-6">
                    <div className="text-center space-y-4">
                      <p className="text-gray-400">
                        {course.price ? `This course costs $${course.price}` : 'This course is free'}
                      </p>
                      <button
                        onClick={() => setShowEnrollForm(true)}
                        className="w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-lg transition-colors disabled:opacity-50"
                        disabled={isEnrolling}
                      >
                        {isEnrolling ? 'Processing…' : 'Enroll Now'}
                      </button>
                      {error && (
                        <p className="text-xs text-red-400">{error}</p>
                      )}
                    </div>
                  </div>
                )
              ) : (
                <div className="bg-gray-800 rounded-lg p-6 space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="text-lg font-bold text-white">Your Progress</h3>
                    <button
                      onClick={handleUnenroll}
                      disabled={unenrollMutation.isLoading}
                      className="text-xs px-3 py-1 rounded bg-red-600 hover:bg-red-700 text-white disabled:opacity-50"
                    >
                      {unenrollMutation.isLoading ? 'Leaving…' : 'Unenroll'}
                    </button>
                  </div>
                  <div className="space-y-4">
                    <div>
                      <div className="flex justify-between mb-2">
                        <span className="text-gray-400">Overall</span>
                        <span className="text-white font-semibold">{Math.round(progress)}%</span>
                      </div>
                      <div className="w-full bg-gray-700 rounded-full h-2">
                        <div
                          className="bg-gradient-to-r from-blue-500 to-purple-600 h-2 rounded-full transition-all"
                          style={{ width: `${progress}%` }}
                        />
                      </div>
                    </div>
                    <p className="text-sm text-gray-400">
                      Keep learning! {Math.round(100 - progress)}% of the course remaining
                    </p>
                    {error && (
                      <p className="text-xs text-red-400">{error}</p>
                    )}
                  </div>
                </div>
              )}

              {/* Course Info Card */}
              <div className="bg-gray-800 rounded-lg p-6">
                <h3 className="text-lg font-bold text-white mb-4">Course Info</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Level</p>
                    <p className="text-white capitalize">{course.level || 'Beginner'}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Students</p>
                    <p className="text-white">{course.students || 0} enrolled</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Duration</p>
                    <p className="text-white">{course.duration || 0} hours</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Lectures</p>
                    <p className="text-white">{lectures.length} lectures</p>
                  </div>
                </div>
              </div>

              {/* Instructor Card */}
              {course.instructor && (
                <div className="bg-gray-800 rounded-lg p-6">
                  <h3 className="text-lg font-bold text-white mb-4">Instructor</h3>
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                      <span className="text-white font-bold">
                        {course.instructor.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div>
                      <p className="text-white font-semibold">{course.instructor}</p>
                      <p className="text-gray-400 text-sm">Instructor</p>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
