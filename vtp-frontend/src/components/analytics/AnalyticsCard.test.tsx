import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import {
  AnalyticsCard,
  StatGrid,
  InsightCard,
  PerformanceIndicator,
} from '@/components/analytics/AnalyticsCard';
import { Activity } from 'lucide-react';

describe('AnalyticsCard', () => {
  it('should render card with title and value', () => {
    render(
      <AnalyticsCard
        title="Active Users"
        value={450}
        suffix="users"
      />
    );

    expect(screen.getByText('Active Users')).toBeInTheDocument();
    expect(screen.getByText('450')).toBeInTheDocument();
    expect(screen.getByText('users')).toBeInTheDocument();
  });

  it('should display trend information', () => {
    render(
      <AnalyticsCard
        title="Revenue"
        value="$15,000"
        trend={{
          value: 23,
          isPositive: true,
          period: 'vs last month',
        }}
      />
    );

    expect(screen.getByText(/23%/)).toBeInTheDocument();
    expect(screen.getByText('vs last month')).toBeInTheDocument();
  });

  it('should render icon when provided', () => {
    const { container } = render(
      <AnalyticsCard
        title="Active Users"
        value={450}
        icon={<Activity className="w-6 h-6" />}
      />
    );

    expect(container.querySelector('svg')).toBeInTheDocument();
  });

  it('should handle click event', () => {
    const handleClick = vi.fn();
    render(
      <AnalyticsCard
        title="Active Users"
        value={450}
        onClick={handleClick}
      />
    );

    const card = screen.getByText('Active Users').closest('div');
    fireEvent.click(card!);

    expect(handleClick).toHaveBeenCalled();
  });

  it('should show loading skeleton', () => {
    const { container } = render(
      <AnalyticsCard
        title="Active Users"
        value={450}
        isLoading={true}
      />
    );

    expect(container.querySelector('.animate-pulse')).toBeInTheDocument();
  });

  it('should display negative trend', () => {
    render(
      <AnalyticsCard
        title="Churn Rate"
        value="5%"
        trend={{
          value: 2,
          isPositive: false,
          period: 'improvement',
        }}
      />
    );

    const trendText = screen.getByText(/2%/);
    expect(trendText).toBeInTheDocument();
  });
});

describe('StatGrid', () => {
  const mockStats = [
    {
      title: 'Total Students',
      value: 1000,
      trend: { value: 12, isPositive: true, period: 'vs last month' },
    },
    {
      title: 'Active Users',
      value: 450,
      trend: { value: 8, isPositive: true, period: 'this week' },
    },
  ];

  it('should render multiple stat cards', () => {
    render(<StatGrid stats={mockStats} />);

    expect(screen.getByText('Total Students')).toBeInTheDocument();
    expect(screen.getByText('Active Users')).toBeInTheDocument();
  });

  it('should apply correct grid columns', () => {
    const { container } = render(<StatGrid stats={mockStats} columns={2} />);

    expect(container.querySelector('.grid-cols-1')).toBeInTheDocument();
  });

  it('should handle card clicks', () => {
    const handleClick = vi.fn();
    render(<StatGrid stats={mockStats} onCardClick={handleClick} />);

    const card = screen.getByText('Total Students').closest('div');
    fireEvent.click(card!);

    expect(handleClick).toHaveBeenCalledWith(0);
  });

  it('should show loading skeletons', () => {
    const { container } = render(
      <StatGrid stats={mockStats} isLoading={true} />
    );

    const skeletons = container.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });
});

describe('InsightCard', () => {
  it('should render insight with title and description', () => {
    render(
      <InsightCard
        title="High Engagement"
        description="75% of users are actively engaged"
        variant="success"
      />
    );

    expect(screen.getByText('High Engagement')).toBeInTheDocument();
    expect(screen.getByText('75% of users are actively engaged')).toBeInTheDocument();
  });

  it('should display action button when provided', () => {
    const handleAction = vi.fn();
    render(
      <InsightCard
        title="Growth Opportunity"
        description="Consider adding new course content"
        actionLabel="View Recommendations"
        onAction={handleAction}
        variant="info"
      />
    );

    const button = screen.getByText('View Recommendations');
    fireEvent.click(button);

    expect(handleAction).toHaveBeenCalled();
  });

  it('should apply correct variant styling', () => {
    const { container } = render(
      <InsightCard
        title="Warning"
        description="Check system health"
        variant="warning"
      />
    );

    expect(container.querySelector('.bg-amber-900')).toBeInTheDocument();
  });

  it('should render icon when provided', () => {
    const { container } = render(
      <InsightCard
        title="Insight"
        description="Some insight"
        icon={<Activity className="w-5 h-5" />}
      />
    );

    expect(container.querySelector('svg')).toBeInTheDocument();
  });
});

describe('PerformanceIndicator', () => {
  it('should render label and value', () => {
    render(
      <PerformanceIndicator
        label="Completion Rate"
        value={75}
        unit="%"
      />
    );

    expect(screen.getByText('Completion Rate')).toBeInTheDocument();
    expect(screen.getByText(/75%/)).toBeInTheDocument();
  });

  it('should display target value when provided', () => {
    render(
      <PerformanceIndicator
        label="Revenue"
        value={15000}
        target={20000}
        unit="$"
      />
    );

    expect(screen.getByText(/15000/)).toBeInTheDocument();
  });

  it('should show progress bar with correct width', () => {
    const { container } = render(
      <PerformanceIndicator
        label="Test"
        value={50}
        target={100}
      />
    );

    const progressBar = container.querySelector('div[style*="width"]');
    expect(progressBar).toBeInTheDocument();
  });

  it('should change color based on performance', () => {
    const { container: container1 } = render(
      <PerformanceIndicator
        label="Good"
        value={85}
        target={100}
        color="green"
      />
    );

    const { container: container2 } = render(
      <PerformanceIndicator
        label="Bad"
        value={25}
        target={100}
        color="red"
      />
    );

    expect(container1.querySelector('.bg-green-600')).toBeInTheDocument();
    expect(container2.querySelector('.bg-red-600')).toBeInTheDocument();
  });

  it('should handle different size options', () => {
    const { rerender } = render(
      <PerformanceIndicator
        label="Test"
        value={50}
        size="sm"
      />
    );

    const value = screen.getByText(/50/);
    expect(value).toBeInTheDocument();

    rerender(
      <PerformanceIndicator
        label="Test"
        value={50}
        size="lg"
      />
    );

    expect(screen.getByText(/50/)).toBeInTheDocument();
  });
});
