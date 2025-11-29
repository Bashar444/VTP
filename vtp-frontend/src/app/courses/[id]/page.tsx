'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/stores/auth';
import { CourseService } from '@/services/course.service';
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

  // Fetch course and lectures
  useEffect(() => {
    const fetchCourseData = async () => {
      try {
        setIsLoading(true);
        const [courseData, lecturesData] = await Promise.all([
          CourseService.getCourseById(courseId),
          CourseService.getCourseLectures(courseId),
        ]);

        setCourse(courseData);
        setLectures(lecturesData);

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

  const handleEnroll = async () => {
    if (!user) {
      router.push('/auth/login');
      return;
    }

    try {
      setIsEnrolling(true);
      await CourseService.enrollCourse(courseId);
      setIsEnrolled(true);
      setShowEnrollForm(false);
      setProgress(0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to enroll in course');
    } finally {
      setIsEnrolling(false);
    }
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
                    <div className="text-center">
                      <p className="text-gray-400 mb-4">
                        {course.price
                          ? `This course costs $${course.price}`
                          : 'This course is free'}
                      </p>
                      <button
                        onClick={() => setShowEnrollForm(true)}
                        className="w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-lg transition-colors"
                      >
                        Enroll Now
                      </button>
                    </div>
                  </div>
                )
              ) : (
                <div className="bg-gray-800 rounded-lg p-6">
                  <h3 className="text-lg font-bold text-white mb-4">Your Progress</h3>
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
