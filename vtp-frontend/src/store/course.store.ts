import { create } from 'zustand';
import type { Course, Enrollment } from '@/types/api';

interface CourseState {
  courses: Course[];
  enrollments: Enrollment[];
  selectedCourse: Course | null;
  isLoading: boolean;

  // Actions
  setCourses: (courses: Course[]) => void;
  setEnrollments: (enrollments: Enrollment[]) => void;
  setSelectedCourse: (course: Course | null) => void;
  setLoading: (loading: boolean) => void;
  addCourse: (course: Course) => void;
  removeCourse: (courseId: string) => void;
}

export const useCourseStore = create<CourseState>((set) => ({
  courses: [],
  enrollments: [],
  selectedCourse: null,
  isLoading: false,

  setCourses: (courses: Course[]) => set({ courses }),
  setEnrollments: (enrollments: Enrollment[]) => set({ enrollments }),
  setSelectedCourse: (course: Course | null) => set({ selectedCourse: course }),
  setLoading: (loading: boolean) => set({ isLoading: loading }),
  addCourse: (course: Course) =>
    set((state) => ({
      courses: [...state.courses, course],
    })),
  removeCourse: (courseId: string) =>
    set((state) => ({
      courses: state.courses.filter((c) => c.id !== courseId),
    })),
}));
