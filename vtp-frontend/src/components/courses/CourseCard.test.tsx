import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CourseCard, CourseList } from '@/components/courses/CourseCard';
import type { Course } from '@/services/course.service';

const mockCourse: Course = {
  id: 'course-1',
  title: 'JavaScript Basics',
  description: 'Learn JavaScript from scratch',
  instructor: 'John Doe',
  thumbnail: 'https://example.com/thumbnail.jpg',
  category: 'programming',
  level: 'beginner',
  status: 'published',
  price: 29.99,
  rating: 4.5,
  students: 1000,
  duration: 10,
  lectures: 25,
};

describe('CourseCard', () => {
  it('should render course card with all information', () => {
    render(
      <CourseCard
        course={mockCourse}
        onSelect={() => {}}
      />
    );

    expect(screen.getByText('JavaScript Basics')).toBeInTheDocument();
    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText(/4\.5/)).toBeInTheDocument();
  });

  it('should call onSelect when card is clicked', () => {
    const onSelect = vi.fn();
    render(
      <CourseCard
        course={mockCourse}
        onSelect={onSelect}
      />
    );

    const card = screen.getByText('JavaScript Basics').closest('div');
    fireEvent.click(card!);

    expect(onSelect).toHaveBeenCalledWith('course-1');
  });

  it('should display progress bar when showProgress is true', () => {
    render(
      <CourseCard
        course={mockCourse}
        onSelect={() => {}}
        showProgress
        progressPercentage={50}
      />
    );

    const progressBar = document.querySelector('[style*="width"]');
    expect(progressBar).toBeInTheDocument();
  });

  it('should render compact variant', () => {
    const { container } = render(
      <CourseCard
        course={mockCourse}
        onSelect={() => {}}
        variant="compact"
      />
    );

    expect(container.querySelector('.h-32')).toBeInTheDocument();
  });

  it('should render featured variant with ring border', () => {
    const { container } = render(
      <CourseCard
        course={mockCourse}
        onSelect={() => {}}
        variant="featured"
      />
    );

    expect(container.querySelector('.ring-2')).toBeInTheDocument();
  });

  it('should display free badge for free courses', () => {
    const freeCourse = { ...mockCourse, price: 0 };
    render(
      <CourseCard
        course={freeCourse}
        onSelect={() => {}}
      />
    );

    expect(screen.getByText('Free')).toBeInTheDocument();
  });

  it('should display price for paid courses', () => {
    render(
      <CourseCard
        course={mockCourse}
        onSelect={() => {}}
      />
    );

    expect(screen.getByText('$29.99')).toBeInTheDocument();
  });
});

describe('CourseList', () => {
  const mockCourses = [mockCourse, { ...mockCourse, id: 'course-2', title: 'React Advanced' }];

  it('should render multiple course cards', () => {
    render(
      <CourseList
        courses={mockCourses}
        onCourseSelect={() => {}}
      />
    );

    expect(screen.getByText('JavaScript Basics')).toBeInTheDocument();
    expect(screen.getByText('React Advanced')).toBeInTheDocument();
  });

  it('should display loading skeleton when isLoading is true', () => {
    const { container } = render(
      <CourseList
        courses={[]}
        isLoading={true}
        onCourseSelect={() => {}}
      />
    );

    const skeletons = container.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('should display empty state when no courses', () => {
    render(
      <CourseList
        courses={[]}
        isLoading={false}
        onCourseSelect={() => {}}
      />
    );

    expect(screen.getByText(/No courses available/i)).toBeInTheDocument();
  });

  it('should call onCourseSelect when course is selected', () => {
    const onSelect = vi.fn();
    render(
      <CourseList
        courses={mockCourses}
        onCourseSelect={onSelect}
      />
    );

    const card = screen.getByText('JavaScript Basics').closest('button');
    fireEvent.click(card!);

    expect(onSelect).toHaveBeenCalledWith('course-1');
  });

  it('should display progress for enrolled courses', () => {
    const { container } = render(
      <CourseList
        courses={mockCourses}
        onCourseSelect={() => {}}
        showProgress={true}
        progressMap={{ 'course-1': 50, 'course-2': 75 }}
      />
    );

    const progressBars = container.querySelectorAll('[style*="width"]');
    expect(progressBars.length).toBeGreaterThan(0);
  });

  it('should apply custom grid columns', () => {
    const { container } = render(
      <CourseList
        courses={mockCourses}
        onCourseSelect={() => {}}
        gridCols="grid-cols-1"
      />
    );

    expect(container.querySelector('.grid-cols-1')).toBeInTheDocument();
  });
});
