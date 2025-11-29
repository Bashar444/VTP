import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Star, Users, Clock, BarChart3, ChevronRight } from 'lucide-react';
import { Course } from '@/services/course.service';
import { cn } from '@/utils/cn';

interface CourseCardProps {
  course: Course;
  onSelect?: (courseId: string) => void;
  showProgress?: boolean;
  progressPercentage?: number;
  variant?: 'default' | 'compact' | 'featured';
  className?: string;
}

export const CourseCard: React.FC<CourseCardProps> = ({
  course,
  onSelect,
  showProgress = false,
  progressPercentage = 0,
  variant = 'default',
  className,
}) => {
  const [imageError, setImageError] = useState(false);
  const t = useTranslations();

  const handleClick = () => {
    onSelect?.(course.id);
  };

  if (variant === 'compact') {
    return (
      <div
        onClick={handleClick}
        className={cn(
          'bg-gray-800 rounded-lg overflow-hidden hover:bg-gray-750 transition-colors cursor-pointer group',
          className
        )}
      >
        <div className="relative w-full aspect-video bg-gray-700 overflow-hidden">
          <img
            src={imageError ? '/placeholder-course.jpg' : course.thumbnail}
            alt={course.title}
            onError={() => setImageError(true)}
            className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-300"
          />
          <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-30 transition-colors" />
        </div>
        <div className="p-3">
          <h3 className="font-semibold text-white line-clamp-2 mb-1">{course.title}</h3>
          <p className="text-sm text-gray-400 mb-2">{course.instructorName}</p>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-1">
              <Star className="w-4 h-4 text-yellow-400" fill="currentColor" />
              <span className="text-sm text-gray-300">{course.rating.toFixed(1)}</span>
            </div>
            <span className="text-xs text-gray-400">{course.studentCount} {t('course.sidebar.students')}</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div
      onClick={handleClick}
      className={cn(
        'bg-gray-800 rounded-lg overflow-hidden hover:bg-gray-750 transition-all duration-300 cursor-pointer group',
        variant === 'featured' && 'ring-2 ring-blue-500',
        className
      )}
    >
      {/* Course Thumbnail */}
      <div className="relative w-full aspect-video bg-gray-700 overflow-hidden">
        <img
          src={imageError ? '/placeholder-course.jpg' : course.thumbnail}
          alt={course.title}
          onError={() => setImageError(true)}
          className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-300"
        />
        <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-40 transition-colors" />

        {/* Level Badge */}
        <div className="absolute top-3 right-3 px-3 py-1 bg-black bg-opacity-70 rounded-full text-xs font-semibold text-white capitalize">
          {course.level}
        </div>

        {/* Play Icon */}
        {variant === 'featured' && (
          <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
            <ChevronRight className="w-12 h-12 text-white" fill="white" />
          </div>
        )}
      </div>

      {/* Course Info */}
      <div className="p-4">
        <h3 className="font-semibold text-white line-clamp-2 mb-2 group-hover:text-blue-400 transition-colors">
          {course.title}
        </h3>

        {/* Instructor */}
        <p className="text-sm text-gray-400 mb-3">{course.instructorName}</p>

        {/* Course Meta */}
        <div className="space-y-2 mb-4">
          {/* Rating and Reviews */}
          <div className="flex items-center gap-4 text-sm">
            <div className="flex items-center gap-1">
              <Star className="w-4 h-4 text-yellow-400" fill="currentColor" />
              <span className="text-gray-300">{course.rating.toFixed(1)}</span>
              <span className="text-gray-500">({course.reviewCount})</span>
            </div>
            <div className="flex items-center gap-1 text-gray-400">
              <Users className="w-4 h-4" />
              <span>{course.studentCount}</span>
            </div>
          </div>

          {/* Duration and Lectures */}
          <div className="flex items-center gap-4 text-sm text-gray-400">
            <div className="flex items-center gap-1">
              <Clock className="w-4 h-4" />
              <span>{Math.round(course.duration / 60)}س</span>
            </div>
            <div className="flex items-center gap-1">
              <BarChart3 className="w-4 h-4" />
              <span>{course.lectureCount} {t('course.sidebar.lectures')}</span>
            </div>
          </div>
        </div>

        {/* Progress Bar */}
        {showProgress && progressPercentage > 0 && (
          <div className="mb-4">
            <div className="flex items-center justify-between mb-1">
              <span className="text-xs font-semibold text-gray-300">{t('course.card.progress')}</span>
              <span className="text-xs text-gray-400">{Math.round(progressPercentage)}%</span>
            </div>
            <div className="w-full bg-gray-700 rounded-full h-2">
              <div
                className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                style={{ width: `${progressPercentage}%` }}
              />
            </div>
          </div>
        )}

        {/* Price or Status */}
        <div className="pt-3 border-t border-gray-700">
          {course.isFree ? (
            <span className="text-sm font-semibold text-green-400">{t('course.card.free')}</span>
          ) : (
            <span className="text-lg font-bold text-white">${course.price}</span>
          )}
        </div>
      </div>
    </div>
  );
};

// Course List Component
interface CourseListProps {
  courses: Course[];
  isLoading?: boolean;
  onCourseSelect?: (courseId: string) => void;
  showProgress?: boolean;
  progressMap?: Record<string, number>;
  gridCols?: number;
  className?: string;
}

export const CourseList: React.FC<CourseListProps> = ({
  courses,
  isLoading = false,
  onCourseSelect,
  showProgress = false,
  progressMap = {},
  gridCols = 3,
  className,
}) => {
  if (isLoading) {
    return (
      <div
        className={cn(
          `grid gap-4`,
          `grid-cols-1 sm:grid-cols-2 lg:grid-cols-${gridCols}`,
          className
        )}
        style={{
          display: 'grid',
          gridTemplateColumns: `repeat(auto-fill, minmax(300px, 1fr))`,
          gap: '1rem',
        }}
      >
        {[...Array(6)].map((_, i) => (
          <div key={i} className="bg-gray-800 rounded-lg overflow-hidden animate-pulse">
            <div className="w-full aspect-video bg-gray-700" />
            <div className="p-4 space-y-3">
              <div className="h-4 bg-gray-700 rounded w-2/3" />
              <div className="h-3 bg-gray-700 rounded w-1/2" />
              <div className="h-2 bg-gray-700 rounded w-full" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (courses.length === 0) {
    return (
      <div className={cn('text-center py-12 text-gray-400', className)}>
        <BarChart3 className="w-12 h-12 mx-auto mb-3 opacity-50" />
        <p className="text-lg">{t('course.list.noCourses')}</p>
        <p className="text-sm mt-1">{t('course.list.tryAdjust')}</p>
      </div>
    );
  }

  return (
    <div
      className={className}
      style={{
        display: 'grid',
        gridTemplateColumns: `repeat(auto-fill, minmax(300px, 1fr))`,
        gap: '1.5rem',
      }}
    >
      {courses.map(course => (
        <CourseCard
          key={course.id}
          course={course}
          onSelect={onCourseSelect}
          showProgress={showProgress}
          progressPercentage={progressMap[course.id] || 0}
        />
      ))}
    </div>
  );
};
