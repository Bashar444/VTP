"use client";
import { useQuery } from '@tanstack/react-query';
import { CourseService } from '@/services/course.service';

export function useFeaturedCourses(limit: number = 4) {
  return useQuery({
    queryKey: ['featured-courses', limit],
    queryFn: async () => {
      const data = await CourseService.getFeaturedCourses(limit);
      return data || [];
    },
    staleTime: 60_000,
    refetchOnWindowFocus: false,
  });
}
