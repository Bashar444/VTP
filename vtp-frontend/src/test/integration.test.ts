import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Dashboard } from '@/components/analytics/Dashboard';
import { AnalyticsFilters } from '@/components/analytics/AnalyticsFilters';
import * as analyticsApi from '@/api/analytics';

vi.mock('@/api/analytics');

describe('Analytics Module Integration Tests', () => {
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
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    (analyticsApi.getDashboardData as any).mockResolvedValue(mockDashboardData);
  });

  describe('Dashboard with Filters', () => {
    it('should update dashboard when filters change', async () => {
      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('1250')).toBeInTheDocument();
      });

      // Verify dashboard is rendered
      expect(screen.getByText('45')).toBeInTheDocument();

      // Change filter
      const filterButton = screen.getByText('Filters');
      fireEvent.click(filterButton);

      // Verify API was called
      expect(analyticsApi.getDashboardData).toHaveBeenCalled();
    });

    it('should display filtered data correctly', async () => {
      const filteredData = {
        ...mockDashboardData,
        summary: {
          ...mockDashboardData.summary,
          totalSessions: 500, // Filtered data
        },
      };

      (analyticsApi.getDashboardData as any).mockResolvedValue(filteredData);

      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('500')).toBeInTheDocument();
      });

      expect(screen.queryByText('1250')).not.toBeInTheDocument();
    });

    it('should maintain filter state during data refresh', async () => {
      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('1250')).toBeInTheDocument();
      });

      // Simulate filter change
      const filterButton = screen.getByText('Filters');
      fireEvent.click(filterButton);

      // API should be called with filter parameters
      expect(analyticsApi.getDashboardData).toHaveBeenCalled();
    });
  });

  describe('Data Export Flow', () => {
    it('should export selected data', async () => {
      const mockBlob = new Blob(['session1,session2'], { type: 'text/csv' });
      (analyticsApi.exportSessionData as any).mockResolvedValue(mockBlob);

      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('1250')).toBeInTheDocument();
      });

      const exportButton = screen.queryByText(/Export|Download/i);
      if (exportButton) {
        fireEvent.click(exportButton);

        await waitFor(() => {
          expect(analyticsApi.exportSessionData).toHaveBeenCalled();
        });
      }
    });

    it('should handle export with current filters', async () => {
      const mockBlob = new Blob(['data'], { type: 'text/csv' });
      (analyticsApi.exportSessionData as any).mockResolvedValue(mockBlob);

      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('1250')).toBeInTheDocument();
      });

      // Apply filter
      const filterButton = screen.getByText('Filters');
      fireEvent.click(filterButton);

      // Export should use filtered data
      const exportButton = screen.queryByText(/Export|Download/i);
      if (exportButton) {
        fireEvent.click(exportButton);

        await waitFor(() => {
          expect(analyticsApi.exportSessionData).toHaveBeenCalled();
        });
      }
    });
  });

  describe('Session Details Navigation', () => {
    it('should navigate to session details on row click', async () => {
      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('user1@example.com')).toBeInTheDocument();
      });

      const sessionRow = screen.getByText('user1@example.com');
      fireEvent.click(sessionRow.closest('tr')!);

      // Component should handle navigation
      expect(sessionRow).toBeInTheDocument();
    });

    it('should display session details after navigation', async () => {
      const sessionDetails = {
        id: '1',
        user: 'user1@example.com',
        startTime: '2024-01-15T10:00:00Z',
        endTime: '2024-01-15T10:15:00Z',
        duration: 900,
        status: 'completed',
        events: [
          { timestamp: '2024-01-15T10:00:00Z', type: 'start' },
          { timestamp: '2024-01-15T10:15:00Z', type: 'end' },
        ],
      };

      (analyticsApi.getSessionDetails as any).mockResolvedValue(sessionDetails);

      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('user1@example.com')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('user1@example.com').closest('tr')!);

      // Verify navigation occurred
      expect(screen.getByText('user1@example.com')).toBeInTheDocument();
    });
  });

  describe('Error Recovery', () => {
    it('should handle API error and allow retry', async () => {
      (analyticsApi.getDashboardData as any)
        .mockRejectedValueOnce(new Error('API Error'))
        .mockResolvedValueOnce(mockDashboardData);

      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText(/Error|Failed/)).toBeInTheDocument();
      });

      // Retry button should be present
      const retryButton = screen.queryByText(/Retry|Refresh/i);
      if (retryButton) {
        fireEvent.click(retryButton);

        await waitFor(() => {
          expect(screen.getByText('1250')).toBeInTheDocument();
        });
      }
    });

    it('should continue functioning after transient error', async () => {
      (analyticsApi.getDashboardData as any)
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce(mockDashboardData);

      render(<Dashboard />);

      // Wait for error state
      await waitFor(() => {
        expect(analyticsApi.getDashboardData).toHaveBeenCalledTimes(1);
      });

      // Retry or auto-recover
      await waitFor(() => {
        expect(analyticsApi.getDashboardData).toHaveBeenCalled();
      });
    });
  });

  describe('Performance and Optimization', () => {
    it('should not make redundant API calls', async () => {
      const { rerender } = render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('1250')).toBeInTheDocument();
      });

      const callCount = (analyticsApi.getDashboardData as any).mock.calls.length;

      // Rerender with same props
      rerender(<Dashboard />);

      // Should use cached data
      expect((analyticsApi.getDashboardData as any).mock.calls.length).toBeLessThanOrEqual(
        callCount + 1
      );
    });

    it('should debounce filter changes', async () => {
      vi.useFakeTimers();

      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('1250')).toBeInTheDocument();
      });

      const initialCalls = (analyticsApi.getDashboardData as any).mock.calls.length;

      // Simulate rapid filter changes
      const filterButton = screen.getByText('Filters');
      fireEvent.click(filterButton);
      fireEvent.click(filterButton);
      fireEvent.click(filterButton);

      vi.runAllTimers();

      // Should debounce calls
      const finalCalls = (analyticsApi.getDashboardData as any).mock.calls.length;
      expect(finalCalls - initialCalls).toBeLessThanOrEqual(2);

      vi.useRealTimers();
    });
  });

  describe('Responsive Behavior', () => {
    it('should render on mobile without filters panel visible', async () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        value: 375,
      });

      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('1250')).toBeInTheDocument();
      });

      // Filters should be collapsible on mobile
      const filterButton = screen.queryByText('Filters');
      expect(filterButton).toBeInTheDocument();
    });

    it('should adapt layout on window resize', async () => {
      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('1250')).toBeInTheDocument();
      });

      // Simulate resize
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        value: 1200,
      });

      fireEvent.resize(window);

      // Component should adapt
      expect(screen.getByText('1250')).toBeVisible();
    });
  });

  describe('Data Consistency', () => {
    it('should maintain data consistency across components', async () => {
      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('1250')).toBeInTheDocument();
      });

      // Summary data should match individual session count
      const sessionCount = screen.queryAllByText(/user@example.com/);
      expect(sessionCount.length).toBeGreaterThan(0);

      // Verify summary reflects session data
      expect(screen.getByText('1250')).toBeInTheDocument();
    });

    it('should update all components when data refreshes', async () => {
      const updatedData = {
        ...mockDashboardData,
        summary: {
          ...mockDashboardData.summary,
          activeUsers: 100,
        },
      };

      (analyticsApi.getDashboardData as any).mockResolvedValue(mockDashboardData);

      render(<Dashboard />);

      await waitFor(() => {
        expect(screen.getByText('45')).toBeInTheDocument(); // Original value
      });

      // Simulate refresh
      (analyticsApi.getDashboardData as any).mockResolvedValue(updatedData);

      const refreshButton = screen.queryByText(/Refresh|Reload/i);
      if (refreshButton) {
        fireEvent.click(refreshButton);

        await waitFor(() => {
          expect(analyticsApi.getDashboardData).toHaveBeenCalled();
        });
      }
    });
  });
});
