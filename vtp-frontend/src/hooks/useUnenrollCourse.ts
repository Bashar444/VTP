"use client";
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { CourseService } from '@/services/course.service';

export function useUnenrollCourse(courseId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationKey: ['unenroll', courseId],
    mutationFn: async () => {
      await CourseService.unenrollCourse(courseId);
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['courses'] }),
        queryClient.invalidateQueries({ queryKey: ['featured-courses'] }),
      ]);
    },
  });
}
