"use client";

import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import Link from 'next/link';

export default function MyCoursesPage() {
  const { user } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!user) {
      router.push('/login');
    }
  }, [user, router]);

  if (!user) return null;

  return (
    <div className="min-h-screen bg-gray-50 pt-24 pb-12">
      <div className="max-w-7xl mx-auto px-4">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">My Courses</h1>

        <div className="bg-white rounded-lg shadow p-8 text-center">
          <p className="text-gray-600 mb-4">
            You haven't enrolled in any courses yet.
          </p>
          <Link
            href="/courses"
            className="inline-block bg-indigo-600 text-white px-6 py-3 rounded-md hover:bg-indigo-700"
          >
            Browse Courses
          </Link>
        </div>
      </div>
    </div>
  );
}
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
                <p className="text-gray-400 mb-6">{t('myCourses.empty.start')}</p>
                <button
                  onClick={() => router.push('/courses')}
                  className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-lg transition-colors"
                >
                  {t('myCourses.empty.browse')}
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
                        {t('myCourses.completed.badge')}
                      </div>
                    </div>

                    {/* Course Info */}
                    <div className="p-4">
                      <h3 className="text-lg font-bold text-white mb-2 line-clamp-2">
                        {course.title}
                      </h3>

                      {/* Instructor */}
                      <p className="text-sm text-gray-400 mb-4">
                        {t('myCourses.instructor.by')} {course.instructor || t('myCourses.instructor.unknown')}
                      </p>

                      {/* Footer */}
                      <button
                        onClick={() => handleViewCourse(course.id)}
                        className="w-full py-2 px-4 bg-green-600 hover:bg-green-700 text-white font-semibold rounded transition-colors"
                      >
                        {t('myCourses.view.certificate')}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-16">
                <CheckCircle className="w-16 h-16 text-gray-600 mx-auto mb-4" />
                <h3 className="text-xl font-semibold text-gray-300 mb-2">{t('myCourses.completed.noneTitle')}</h3>
                <p className="text-gray-400 mb-6">{t('myCourses.completed.noneDesc')}</p>
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
