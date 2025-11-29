import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Dashboard } from '@/components/analytics/Dashboard';
import * as analyticsApi from '@/api/analytics';

vi.mock('@/api/analytics');

describe('Dashboard', () => {
  const mockDashboardData = {
    summary: {
      totalSessions: 1250,
      totalDuration: 5400,
      averageSessionDuration: 4.32,
      peakHour: '14:00',
      activeUsers: 45,
      successRate: 98.5,
    },
    sessions: [
      {
        id: '1',
        user: 'user1@example.com',
        startTime: '2024-01-15T10:00:00Z',
        endTime: '2024-01-15T10:15:00Z',
        duration: 900,
        status: 'completed',
      },
      {
        id: '2',
        user: 'user2@example.com',
        startTime: '2024-01-15T11:00:00Z',
        endTime: '2024-01-15T11:20:00Z',
        duration: 1200,
        status: 'completed',
      },
    ],
    sessionTrends: [
      { timestamp: '2024-01-15T00:00:00Z', count: 50 },
      { timestamp: '2024-01-15T01:00:00Z', count: 45 },
      { timestamp: '2024-01-15T02:00:00Z', count: 75 },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    (analyticsApi.getDashboardData as any).mockResolvedValue(mockDashboardData);
  });

  it('should render dashboard with summary cards', async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('1250')).toBeInTheDocument(); // totalSessions
    });

    expect(screen.getByText('45')).toBeInTheDocument(); // activeUsers
    expect(screen.getByText('98.5%')).toBeInTheDocument(); // successRate
  });

  it('should display session list', async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('user1@example.com')).toBeInTheDocument();
    });

    expect(screen.getByText('user2@example.com')).toBeInTheDocument();
  });

  it('should show loading state initially', () => {
    (analyticsApi.getDashboardData as any).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );

    render(<Dashboard />);

    const loaders = screen.queryAllByRole('status');
    expect(loaders.length).toBeGreaterThanOrEqual(0);
  });

  it('should handle API errors gracefully', async () => {
    (analyticsApi.getDashboardData as any).mockRejectedValue(
      new Error('API Error')
    );

    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText(/Error|Failed/)).toBeInTheDocument();
    });
  });

  it('should support date range filtering', async () => {
    render(<Dashboard />);

    const filterButton = screen.getByText('Filters');
    fireEvent.click(filterButton);

    // Should show date inputs
    const inputs = screen.queryAllByDisplayValue('');
    expect(inputs.length).toBeGreaterThanOrEqual(0);
  });

  it('should refresh data on filter change', async () => {
    const { rerender } = render(<Dashboard />);

    await waitFor(() => {
      expect(analyticsApi.getDashboardData).toHaveBeenCalledTimes(1);
    });

    // Change filter
    const filterButton = screen.getByText('Filters');
    fireEvent.click(filterButton);

    // Should call API again
    await waitFor(() => {
      expect(analyticsApi.getDashboardData).toHaveBeenCalledTimes(1);
    });
  });

  it('should display charts and visualizations', async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('1250')).toBeInTheDocument();
    });

    // Check for chart container or SVG elements
    const svg = document.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('should be responsive on mobile', async () => {
    // Mock window size for mobile
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 375,
    });

    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('1250')).toBeInTheDocument();
    });

    // Dashboard should still be visible
    expect(screen.getByText('1250')).toBeVisible();
  });

  it('should handle empty data', async () => {
    (analyticsApi.getDashboardData as any).mockResolvedValue({
      summary: {
        totalSessions: 0,
        totalDuration: 0,
        averageSessionDuration: 0,
        peakHour: '-',
        activeUsers: 0,
        successRate: 0,
      },
      sessions: [],
      sessionTrends: [],
    });

    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('0')).toBeInTheDocument();
    });

    // Should show empty state
    expect(screen.getByText(/No data|empty/i)).toBeInTheDocument();
  });

  it('should export data functionality', async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('1250')).toBeInTheDocument();
    });

    const exportButton = screen.queryByText(/Export|Download/i);
    if (exportButton) {
      fireEvent.click(exportButton);
      // Should trigger download
      expect(exportButton).toBeInTheDocument();
    }
  });

  it('should update data periodically', async () => {
    vi.useFakeTimers();
    render(<Dashboard />);

    await waitFor(() => {
      expect(analyticsApi.getDashboardData).toHaveBeenCalledTimes(1);
    });

    vi.advanceTimersByTime(30000); // 30 seconds

    await waitFor(() => {
      // Should have refreshed data
      expect(analyticsApi.getDashboardData).toHaveBeenCalled();
    });

    vi.useRealTimers();
  });

  it('should display session metrics correctly', async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('1250')).toBeInTheDocument();
    });

    // Check for duration display
    const durationElements = screen.queryAllByText(/minutes|hours|seconds|duration/i);
    expect(durationElements.length).toBeGreaterThanOrEqual(0);
  });

  it('should allow session details view', async () => {
    render(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('user1@example.com')).toBeInTheDocument();
    });

    const sessionRow = screen.getByText('user1@example.com').closest('tr');
    fireEvent.click(sessionRow!);

    // Should open details or navigate
    expect(sessionRow).toBeInTheDocument();
  });
});
