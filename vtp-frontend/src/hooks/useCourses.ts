"use client";
import { useQuery } from '@tanstack/react-query';
import { CourseService } from '@/services/course.service';
import type { Course } from '@/services/course.service';

export interface CourseFilters {
  category?: string;
  level?: string;
  search?: string;
  instructor?: string;
  limit?: number;
  offset?: number;
}

export function useCourses(filters?: CourseFilters) {
  return useQuery<{ courses: Course[]; total: number }, Error>({
    queryKey: ['courses', filters],
    queryFn: async () => {
      const result = await CourseService.getCourses(filters);
      // Ensure shape normalization in case backend differs
      if (Array.isArray(result as any)) {
        return { courses: result as unknown as Course[], total: (result as Course[]).length };
      }
      return result;
    },
    staleTime: 30_000,
    refetchOnWindowFocus: false,
  });
}
