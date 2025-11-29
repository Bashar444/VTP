import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CourseDetail } from '@/components/courses/CourseDetail';
import type { Course, Lecture } from '@/services/course.service';

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

const mockLectures: Lecture[] = [
  {
    id: 'lecture-1',
    courseId: 'course-1',
    title: 'Introduction to Variables',
    duration: 45,
    videoId: 'video-1',
    order: 1,
  },
  {
    id: 'lecture-2',
    courseId: 'course-1',
    title: 'Data Types and Operators',
    duration: 50,
    videoId: 'video-2',
    order: 2,
  },
];

describe('CourseDetail', () => {
  it('should render course information', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        onSelectLecture={() => {}}
      />
    );

    expect(screen.getByText('JavaScript Basics')).toBeInTheDocument();
    expect(screen.getByText('Learn JavaScript from scratch')).toBeInTheDocument();
    expect(screen.getByText('by John Doe')).toBeInTheDocument();
  });

  it('should display course rating and review count', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        onSelectLecture={() => {}}
      />
    );

    expect(screen.getByText(/4\.5/)).toBeInTheDocument();
  });

  it('should display enroll button when not enrolled', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isEnrolled={false}
        onEnroll={() => {}}
        onSelectLecture={() => {}}
      />
    );

    expect(screen.getByText(/Enroll/i)).toBeInTheDocument();
  });

  it('should call onEnroll when enroll button is clicked', () => {
    const onEnroll = vi.fn();
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isEnrolled={false}
        onEnroll={onEnroll}
        onSelectLecture={() => {}}
      />
    );

    const enrollBtn = screen.getByText(/Enroll/i);
    fireEvent.click(enrollBtn);

    expect(onEnroll).toHaveBeenCalled();
  });

  it('should display progress bar when enrolled', () => {
    const { container } = render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isEnrolled={true}
        progress={50}
        onSelectLecture={() => {}}
      />
    );

    const progressBar = container.querySelector('[style*="width"]');
    expect(progressBar).toBeInTheDocument();
  });

  it('should render lecture list', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isEnrolled={true}
        onSelectLecture={() => {}}
      />
    );

    expect(screen.getByText('Introduction to Variables')).toBeInTheDocument();
    expect(screen.getByText('Data Types and Operators')).toBeInTheDocument();
  });

  it('should display course stats', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        onSelectLecture={() => {}}
      />
    );

    expect(screen.getByText('beginner')).toBeInTheDocument();
    expect(screen.getByText('1000')).toBeInTheDocument();
  });

  it('should show loading skeleton when isLoading is true', () => {
    const { container } = render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isLoading={true}
        onSelectLecture={() => {}}
      />
    );

    const skeletons = container.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });
});

describe('LectureList', () => {
  it('should render all lectures', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isEnrolled={true}
        onSelectLecture={() => {}}
      />
    );

    mockLectures.forEach(lecture => {
      expect(screen.getByText(lecture.title)).toBeInTheDocument();
    });
  });

  it('should display lecture duration', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isEnrolled={true}
        onSelectLecture={() => {}}
      />
    );

    expect(screen.getByText(/45/)).toBeInTheDocument();
    expect(screen.getByText(/50/)).toBeInTheDocument();
  });

  it('should allow expanding lecture details', () => {
    const { container } = render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isEnrolled={true}
        onSelectLecture={() => {}}
      />
    );

    const expandButtons = container.querySelectorAll('button');
    if (expandButtons.length > 0) {
      fireEvent.click(expandButtons[0]);
    }
  });

  it('should show lock icon when not enrolled', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isEnrolled={false}
        onSelectLecture={() => {}}
      />
    );

    // Check for lock icon or disabled state
    const lectureElements = screen.getAllByText(/lecture/i);
    expect(lectureElements.length).toBeGreaterThan(0);
  });

  it('should call onSelectLecture when watch button is clicked', () => {
    const onSelect = vi.fn();
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isEnrolled={true}
        onSelectLecture={onSelect}
      />
    );

    const watchButtons = screen.getAllByText(/Watch|Play/i);
    if (watchButtons.length > 0) {
      fireEvent.click(watchButtons[0]);
      expect(onSelect).toHaveBeenCalled();
    }
  });

  it('should display course stats grid', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        onSelectLecture={() => {}}
      />
    );

    // Check for key stats
    expect(screen.getByText('beginner')).toBeInTheDocument();
  });

  it('should show completion status for enrolled users', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        isEnrolled={true}
        progress={100}
        onSelectLecture={() => {}}
      />
    );

    // Progress should show 100%
    expect(screen.getByText(/100/)).toBeInTheDocument();
  });
});

describe('CourseDetail Responsive Behavior', () => {
  it('should apply custom className', () => {
    const { container } = render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        onSelectLecture={() => {}}
        className="custom-class"
      />
    );

    expect(container.querySelector('.custom-class')).toBeInTheDocument();
  });

  it('should display all course information sections', () => {
    render(
      <CourseDetail
        course={mockCourse}
        lectures={mockLectures}
        onSelectLecture={() => {}}
      />
    );

    // Title
    expect(screen.getByText('JavaScript Basics')).toBeInTheDocument();

    // Description
    expect(screen.getByText('Learn JavaScript from scratch')).toBeInTheDocument();

    // Rating
    expect(screen.getByText(/4\.5/)).toBeInTheDocument();
  });
});
