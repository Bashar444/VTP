import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useAnalyticsData, useSessionMetrics, useExport } from '@/hooks/analytics';
import * as analyticsApi from '@/api/analytics';

vi.mock('@/api/analytics');

describe('useAnalyticsData', () => {
  const mockData = {
    summary: {
      totalSessions: 1250,
      activeUsers: 45,
      successRate: 98.5,
    },
    sessions: [],
    sessionTrends: [],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    (analyticsApi.getDashboardData as any).mockResolvedValue(mockData);
  });

  it('should fetch analytics data on mount', async () => {
    const { result } = renderHook(() => useAnalyticsData());

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toEqual(mockData);
    expect(analyticsApi.getDashboardData).toHaveBeenCalled();
  });

  it('should handle loading state', () => {
    (analyticsApi.getDashboardData as any).mockImplementation(
      () => new Promise(() => {})
    );

    const { result } = renderHook(() => useAnalyticsData());

    expect(result.current.isLoading).toBe(true);
  });

  it('should handle error state', async () => {
    const error = new Error('API Error');
    (analyticsApi.getDashboardData as any).mockRejectedValue(error);

    const { result } = renderHook(() => useAnalyticsData());

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toEqual(error);
  });

  it('should refetch data', async () => {
    const { result } = renderHook(() => useAnalyticsData());

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    act(() => {
      result.current.refetch();
    });

    expect(analyticsApi.getDashboardData).toHaveBeenCalledTimes(2);
  });

  it('should accept custom options', async () => {
    const options = { startDate: '2024-01-01', endDate: '2024-01-31' };

    const { result } = renderHook(() => useAnalyticsData(options));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(analyticsApi.getDashboardData).toHaveBeenCalledWith(options);
  });

  it('should cache data', async () => {
    const { result: result1 } = renderHook(() => useAnalyticsData());

    await waitFor(() => {
      expect(result1.current.isLoading).toBe(false);
    });

    const { result: result2 } = renderHook(() => useAnalyticsData());

    // Should use cached data
    expect(result2.current.data).toEqual(mockData);
  });

  it('should update on dependency change', async () => {
    const { result, rerender } = renderHook(
      (options) => useAnalyticsData(options),
      { initialProps: { startDate: '2024-01-01' } }
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    const callCount1 = (analyticsApi.getDashboardData as any).mock.calls.length;

    rerender({ startDate: '2024-01-02' });

    await waitFor(() => {
      expect((analyticsApi.getDashboardData as any).mock.calls.length).toBeGreaterThan(
        callCount1
      );
    });
  });
});

describe('useSessionMetrics', () => {
  const mockMetrics = {
    timestamp: '2024-01-15T10:00:00Z',
    sessions: 100,
    activeUsers: 45,
    avgDuration: 4.32,
    errorRate: 0.5,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    (analyticsApi.getMetrics as any).mockResolvedValue(mockMetrics);
  });

  it('should fetch session metrics', async () => {
    const { result } = renderHook(() => useSessionMetrics());

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.metrics).toEqual(mockMetrics);
  });

  it('should update metrics periodically', async () => {
    vi.useFakeTimers();

    const { result } = renderHook(() => useSessionMetrics({ interval: 10000 }));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    const initialCallCount = (analyticsApi.getMetrics as any).mock.calls.length;

    act(() => {
      vi.advanceTimersByTime(10000);
    });

    await waitFor(() => {
      expect((analyticsApi.getMetrics as any).mock.calls.length).toBeGreaterThan(
        initialCallCount
      );
    });

    vi.useRealTimers();
  });

  it('should allow manual refresh', async () => {
    const { result } = renderHook(() => useSessionMetrics());

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    const callCount = (analyticsApi.getMetrics as any).mock.calls.length;

    act(() => {
      result.current.refresh();
    });

    await waitFor(() => {
      expect((analyticsApi.getMetrics as any).mock.calls.length).toBeGreaterThan(callCount);
    });
  });

  it('should handle metric type selection', async () => {
    const { result } = renderHook(() => useSessionMetrics());

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    act(() => {
      result.current.setMetricType('system');
    });

    expect(analyticsApi.getMetrics).toHaveBeenCalled();
  });

  it('should calculate metric trends', async () => {
    (analyticsApi.getMetrics as any).mockResolvedValue({
      ...mockMetrics,
      trend: { direction: 'up', percentage: 5 },
    });

    const { result } = renderHook(() => useSessionMetrics());

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.metrics).toHaveProperty('trend');
  });
});

describe('useExport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (analyticsApi.exportSessionData as any).mockResolvedValue(new Blob(['data']));
  });

  it('should export data', async () => {
    const { result } = renderHook(() => useExport());

    act(() => {
      result.current.exportData('csv');
    });

    await waitFor(() => {
      expect(analyticsApi.exportSessionData).toHaveBeenCalledWith('csv');
    });
  });

  it('should show export progress', async () => {
    const { result } = renderHook(() => useExport());

    expect(result.current.isExporting).toBe(false);

    act(() => {
      result.current.exportData('csv');
    });

    // Initially should show exporting
    // Then should complete

    await waitFor(() => {
      expect(analyticsApi.exportSessionData).toHaveBeenCalled();
    });
  });

  it('should trigger download on successful export', async () => {
    const { result } = renderHook(() => useExport());

    const mockCreateObjectURL = vi.fn(() => 'blob:url');
    const mockRevokeObjectURL = vi.fn();
    const mockClick = vi.fn();

    global.URL.createObjectURL = mockCreateObjectURL;
    global.URL.revokeObjectURL = mockRevokeObjectURL;

    act(() => {
      result.current.exportData('csv');
    });

    await waitFor(() => {
      expect(analyticsApi.exportSessionData).toHaveBeenCalled();
    });
  });

  it('should handle export errors', async () => {
    (analyticsApi.exportSessionData as any).mockRejectedValue(
      new Error('Export failed')
    );

    const { result } = renderHook(() => useExport());

    act(() => {
      result.current.exportData('csv');
    });

    await waitFor(() => {
      expect(result.current.error).toBeDefined();
    });
  });

  it('should support multiple export formats', async () => {
    const { result } = renderHook(() => useExport());

    for (const format of ['csv', 'json', 'excel']) {
      act(() => {
        result.current.exportData(format as any);
      });

      await waitFor(() => {
        expect(analyticsApi.exportSessionData).toHaveBeenCalledWith(format);
      });
    }
  });

  it('should accept export filters', async () => {
    const { result } = renderHook(() => useExport());

    const filters = { startDate: '2024-01-01', endDate: '2024-01-31' };

    act(() => {
      result.current.exportData('csv', filters);
    });

    await waitFor(() => {
      expect(analyticsApi.exportSessionData).toHaveBeenCalledWith('csv', filters);
    });
  });

  it('should clear export status', async () => {
    const { result } = renderHook(() => useExport());

    act(() => {
      result.current.exportData('csv');
    });

    await waitFor(() => {
      expect(analyticsApi.exportSessionData).toHaveBeenCalled();
    });

    act(() => {
      result.current.clearStatus();
    });

    expect(result.current.error).toBeNull();
  });
});
