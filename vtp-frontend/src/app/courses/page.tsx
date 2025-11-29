"use client";
export const dynamic = 'force-dynamic';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/store';
import { CourseService } from '@/services/course.service';
import { useCourses } from '@/hooks/useCourses';
import { CourseCard, CourseList } from '@/components/courses/CourseCard';
import { CourseFilters, CourseFilterState } from '@/components/courses/CourseFilters';
import type { Course } from '@/services/course.service';

export default function CoursesPage() {
  const router = useRouter();
  const { user } = useAuth();
  const [courses, setCourses] = useState<Course[]>([]);
  const { data, isLoading, error, refetch } = useCourses();
  const [filteredCourses, setFilteredCourses] = useState<Course[]>([]);
  const [refreshing, setRefreshing] = useState(false);
  const [filters, setFilters] = useState<CourseFilterState>({});

  // Fetch courses
  useEffect(() => {
    if (data) {
      // Sanitize fetched course data to ensure plain objects for Next.js client serialization
      const plainCourses: Course[] = data.courses.map(c => JSON.parse(JSON.stringify(c)) as Course);
      setCourses(plainCourses);
      setFilteredCourses(plainCourses);
    }
  }, [data]);

  // Apply filters
  useEffect(() => {
    let filtered = courses;

    // Search filter
    if (filters.search) {
      const query = filters.search.toLowerCase();
      filtered = filtered.filter(
        course =>
          course.title.toLowerCase().includes(query) ||
          course.description.toLowerCase().includes(query)
      );
    }

    // Category filter
    if (filters.category) {
      filtered = filtered.filter(c => c.category?.toLowerCase() === filters.category);
    }

    // Level filter
    if (filters.level) {
      filtered = filtered.filter(c => c.level?.toLowerCase() === filters.level);
    }

    // Rating filter
    if (filters.rating) {
      filtered = filtered.filter(c => (c.rating || 0) >= filters.rating);
    }

    // Sort
    if (filters.sortBy) {
      switch (filters.sortBy) {
        case 'popular':
          filtered.sort((a, b) => (b.students || 0) - (a.students || 0));
          break;
        case 'highest-rated':
          filtered.sort((a, b) => (b.rating || 0) - (a.rating || 0));
          break;
        case 'price-low':
          filtered.sort((a, b) => (a.price || 0) - (b.price || 0));
          break;
        case 'price-high':
          filtered.sort((a, b) => (b.price || 0) - (a.price || 0));
          break;
        case 'newest':
        default:
          // Assuming courses have a createdAt field
          break;
      }
    }

    setFilteredCourses(filtered);
  }, [courses, filters]);

  const handleCourseSelect = (courseId: string) => {
    router.push(`/courses/${courseId}`);
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await refetch();
    } finally {
      setRefreshing(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        {/* Page Header */}
        <div className="mb-8 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
          <div>
            <h1 className="text-4xl font-bold text-white mb-2">Explore Courses</h1>
            <p className="text-gray-400">{filteredCourses.length} courses available</p>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={handleRefresh}
              disabled={refreshing || isLoading}
              className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium"
            >
              {refreshing ? 'Refreshing…' : 'Refresh'}
            </button>
            <button
              onClick={() => setFilters({})}
              className="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-white text-sm font-medium"
            >
              Reset Filters
            </button>
          </div>
        </div>

        {/* Main Content */}
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          {/* Sidebar Filters */}
          <div className="lg:col-span-1">
            <div className="sticky top-24">
              <CourseFilters
                onFilterChange={setFilters}
                onSearch={query => setFilters(prev => ({ ...prev, search: query }))}
              />
            </div>
          </div>

          {/* Courses Grid */}
          <div className="lg:col-span-3">
            {error && (
              <div className="bg-red-900/20 border border-red-700 rounded-lg p-4 mb-6">
                <p className="text-red-400">{(error as Error).message}</p>
              </div>
            )}

            {isLoading ? (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {[...Array(6)].map((_, i) => (
                  <div key={i} className="bg-gray-800 rounded-lg animate-pulse h-80" />
                ))}
              </div>
            ) : filteredCourses.length > 0 ? (
              <CourseList
                courses={filteredCourses}
                onCourseSelect={handleCourseSelect}
                gridCols="grid-cols-1 md:grid-cols-2"
              />
            ) : (
              <div className="text-center py-12">
                <p className="text-gray-400 text-lg mb-4">No courses found</p>
                <button
                  onClick={() => setFilters({})}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
                >
                  Clear Filters
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
