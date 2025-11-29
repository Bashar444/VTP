import { useState, useEffect } from 'react';
import { Clock, Users, Star, ChevronDown, ChevronUp, Play, Lock, Check } from 'lucide-react';
import CourseService, { Course, Lecture, CourseProgress } from '@/services/course.service';
import { cn } from '@/utils/cn';

interface CourseDetailProps {
  course: Course;
  lectures: Lecture[];
  progress?: CourseProgress;
  isEnrolled?: boolean;
  onEnroll?: () => void;
  onSelectLecture?: (lectureId: string) => void;
  isLoading?: boolean;
  className?: string;
}

export const CourseDetail: React.FC<CourseDetailProps> = ({
  course,
  lectures,
  progress,
  isEnrolled = false,
  onEnroll,
  onSelectLecture,
  isLoading = false,
  className,
}) => {
  return (
    <div className={cn('bg-gray-800 rounded-lg overflow-hidden', className)}>
      {/* Cover Image */}
      <div className="w-full aspect-video bg-gradient-to-b from-gray-700 to-gray-900 overflow-hidden">
        <img
          src={course.coverImage || course.thumbnail}
          alt={course.title}
          className="w-full h-full object-cover"
        />
      </div>

      {/* Course Info */}
      <div className="p-6 lg:p-8">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-3xl lg:text-4xl font-bold text-white mb-3">{course.title}</h1>
          <p className="text-gray-400 text-lg mb-4">{course.description}</p>

          {/* Meta Info */}
          <div className="flex flex-wrap gap-6 text-sm">
            {/* Instructor */}
            <div>
              <p className="text-gray-500">Instructor</p>
              <p className="text-white font-semibold">{course.instructorName}</p>
            </div>

            {/* Rating */}
            <div className="flex items-center gap-2">
              <div className="flex items-center gap-1">
                <Star className="w-5 h-5 text-yellow-400" fill="currentColor" />
                <span className="text-white font-semibold">{course.rating.toFixed(1)}</span>
              </div>
              <span className="text-gray-400">({course.reviewCount} reviews)</span>
            </div>

            {/* Students */}
            <div className="flex items-center gap-2">
              <Users className="w-5 h-5 text-gray-400" />
              <span className="text-gray-400">{course.studentCount} students</span>
            </div>

            {/* Level */}
            <div>
              <p className="text-gray-500">Level</p>
              <p className="text-white font-semibold capitalize">{course.level}</p>
            </div>
          </div>
        </div>

        {/* Course Stats */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6 p-4 bg-gray-900 rounded-lg">
          <div>
            <p className="text-gray-400 text-sm">Duration</p>
            <div className="flex items-center gap-2 text-white font-semibold mt-1">
              <Clock className="w-4 h-4" />
              <span>{Math.round(course.duration / 60)}h</span>
            </div>
          </div>
          <div>
            <p className="text-gray-400 text-sm">Lectures</p>
            <p className="text-white font-semibold mt-1">{course.lectureCount}</p>
          </div>
          <div>
            <p className="text-gray-400 text-sm">Category</p>
            <p className="text-white font-semibold mt-1 capitalize">{course.category}</p>
          </div>
          <div>
            <p className="text-gray-400 text-sm">Price</p>
            <p className="text-white font-semibold mt-1">
              {course.isFree ? 'Free' : `$${course.price}`}
            </p>
          </div>
        </div>

        {/* Progress Bar */}
        {isEnrolled && progress && (
          <div className="mb-6 p-4 bg-blue-900 bg-opacity-30 rounded-lg border border-blue-700">
            <div className="flex items-center justify-between mb-2">
              <span className="text-white font-semibold">Your Progress</span>
              <span className="text-blue-400 font-semibold">
                {Math.round(progress.completionPercentage)}%
              </span>
            </div>
            <div className="w-full bg-gray-700 rounded-full h-2.5">
              <div
                className="bg-blue-500 h-2.5 rounded-full transition-all duration-300"
                style={{ width: `${progress.completionPercentage}%` }}
              />
            </div>
            <p className="text-gray-400 text-sm mt-2">
              {progress.completedLectures} of {progress.totalLectures} lectures completed
            </p>
          </div>
        )}

        {/* Enroll Button */}
        {!isEnrolled && (
          <button
            onClick={onEnroll}
            disabled={isLoading}
            className="w-full py-3 px-6 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 text-white font-bold rounded-lg transition-colors mb-6"
          >
            {isLoading ? 'Enrolling...' : course.isFree ? 'Enroll Now' : `Enroll Now - $${course.price}`}
          </button>
        )}
      </div>
    </div>
  );
};

// Lecture List Component
interface LectureListProps {
  lectures: Lecture[];
  progress?: CourseProgress;
  isEnrolled?: boolean;
  onSelectLecture?: (lectureId: string) => void;
  className?: string;
}

export const LectureList: React.FC<LectureListProps> = ({
  lectures,
  progress,
  isEnrolled = false,
  onSelectLecture,
  className,
}) => {
  const [expandedLectures, setExpandedLectures] = useState<Record<string, boolean>>({});

  const completedLectureIds = new Set<string>();
  if (progress?.currentLecture) {
    // Mark lectures up to current as completed
    const currentIndex = lectures.findIndex(l => l.id === progress.currentLecture);
    for (let i = 0; i <= currentIndex; i++) {
      completedLectureIds.add(lectures[i]?.id || '');
    }
  }

  const toggleExpand = (lectureId: string) => {
    setExpandedLectures(prev => ({
      ...prev,
      [lectureId]: !prev[lectureId],
    }));
  };

  if (lectures.length === 0) {
    return (
      <div className={cn('text-center py-8 text-gray-400', className)}>
        <p>No lectures available yet</p>
      </div>
    );
  }

  return (
    <div className={cn('space-y-3', className)}>
      <h3 className="text-lg font-bold text-white mb-4">Course Content</h3>

      {lectures.map(lecture => (
        <div
          key={lecture.id}
          className={cn(
            'bg-gray-800 rounded-lg overflow-hidden transition-colors',
            isEnrolled ? 'hover:bg-gray-750 cursor-pointer' : ''
          )}
        >
          {/* Lecture Header */}
          <div
            onClick={() => {
              if (isEnrolled) {
                onSelectLecture?.(lecture.id);
              }
              toggleExpand(lecture.id);
            }}
            className={cn(
              'p-4 flex items-center justify-between',
              isEnrolled && 'hover:bg-gray-750'
            )}
          >
            <div className="flex items-center gap-3 flex-1 min-w-0">
              {/* Icon */}
              {!isEnrolled ? (
                <Lock className="w-5 h-5 text-gray-500 flex-shrink-0" />
              ) : completedLectureIds.has(lecture.id) ? (
                <Check className="w-5 h-5 text-green-500 flex-shrink-0" fill="currentColor" />
              ) : (
                <Play className="w-5 h-5 text-blue-400 flex-shrink-0" fill="currentColor" />
              )}

              {/* Lecture Title */}
              <div className="flex-1 min-w-0">
                <h4 className="font-semibold text-white truncate">{lecture.title}</h4>
                <p className="text-sm text-gray-400">{Math.round(lecture.duration / 60)} min</p>
              </div>
            </div>

            {/* Expand Button */}
            <button
              className="p-1 ml-2 text-gray-400 hover:text-white transition-colors"
              onClick={(e) => {
                e.stopPropagation();
                toggleExpand(lecture.id);
              }}
            >
              {expandedLectures[lecture.id] ? (
                <ChevronUp className="w-5 h-5" />
              ) : (
                <ChevronDown className="w-5 h-5" />
              )}
            </button>
          </div>

          {/* Lecture Details (Expanded) */}
          {expandedLectures[lecture.id] && (
            <div className="p-4 border-t border-gray-700 bg-gray-900">
              <p className="text-gray-400 text-sm mb-3">{lecture.description}</p>
              {isEnrolled && (
                <button
                  onClick={() => onSelectLecture?.(lecture.id)}
                  className="w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg transition-colors"
                >
                  <div className="flex items-center justify-center gap-2">
                    <Play className="w-4 h-4" fill="currentColor" />
                    Watch Lecture
                  </div>
                </button>
              )}
            </div>
          )}
        </div>
      ))}
    </div>
  );
};
