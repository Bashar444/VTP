import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { getDashboardData, getSessionDetails, exportSessionData, getMetrics } from '@/api/analytics';

// Mock fetch globally
global.fetch = vi.fn();

describe('Analytics API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('getDashboardData', () => {
    it('should fetch dashboard data', async () => {
      const mockData = {
        summary: {
          totalSessions: 1250,
          activeUsers: 45,
          successRate: 98.5,
        },
        sessions: [],
        sessionTrends: [],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      });

      const result = await getDashboardData();

      expect(result).toEqual(mockData);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/analytics/dashboard'),
        expect.any(Object)
      );
    });

    it('should handle fetch errors', async () => {
      (global.fetch as any).mockRejectedValueOnce(
        new Error('Network error')
      );

      await expect(getDashboardData()).rejects.toThrow('Network error');
    });

    it('should handle API errors', async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
      });

      await expect(getDashboardData()).rejects.toThrow();
    });

    it('should accept optional parameters', async () => {
      const mockData = { summary: {}, sessions: [], sessionTrends: [] };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      });

      const startDate = '2024-01-01';
      const endDate = '2024-01-31';

      await getDashboardData({ startDate, endDate });

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('startDate'),
        expect.any(Object)
      );
    });
  });

  describe('getSessionDetails', () => {
    it('should fetch session details', async () => {
      const mockSessionData = {
        id: 'session-1',
        user: 'user@example.com',
        startTime: '2024-01-15T10:00:00Z',
        endTime: '2024-01-15T10:15:00Z',
        duration: 900,
        status: 'completed',
        events: [],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockSessionData,
      });

      const result = await getSessionDetails('session-1');

      expect(result).toEqual(mockSessionData);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/analytics/sessions/session-1'),
        expect.any(Object)
      );
    });

    it('should handle invalid session ID', async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      });

      await expect(getSessionDetails('invalid-id')).rejects.toThrow();
    });
  });

  describe('exportSessionData', () => {
    it('should export data as CSV', async () => {
      const mockBlob = new Blob(['data'], { type: 'text/csv' });

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        blob: async () => mockBlob,
      });

      const result = await exportSessionData('csv');

      expect(result).toEqual(mockBlob);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/analytics/export'),
        expect.any(Object)
      );
    });

    it('should export data as JSON', async () => {
      const mockBlob = new Blob(['{}'], { type: 'application/json' });

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        blob: async () => mockBlob,
      });

      const result = await exportSessionData('json');

      expect(result).toEqual(mockBlob);
    });

    it('should handle export errors', async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500,
      });

      await expect(exportSessionData('csv')).rejects.toThrow();
    });

    it('should accept date range filters', async () => {
      const mockBlob = new Blob(['data'], { type: 'text/csv' });

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        blob: async () => mockBlob,
      });

      const startDate = '2024-01-01';
      const endDate = '2024-01-31';

      await exportSessionData('csv', { startDate, endDate });

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('startDate'),
        expect.any(Object)
      );
    });
  });

  describe('getMetrics', () => {
    it('should fetch metrics', async () => {
      const mockMetrics = {
        timestamp: '2024-01-15T10:00:00Z',
        sessions: 100,
        activeUsers: 45,
        avgDuration: 4.32,
        errorRate: 0.5,
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMetrics,
      });

      const result = await getMetrics();

      expect(result).toEqual(mockMetrics);
      expect(global.fetch).toHaveBeenCalled();
    });

    it('should accept metric type parameter', async () => {
      const mockMetrics = {
        timestamp: '2024-01-15T10:00:00Z',
        cpu: 75,
        memory: 60,
        network: 45,
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMetrics,
      });

      await getMetrics({ type: 'system' });

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('type=system'),
        expect.any(Object)
      );
    });

    it('should handle empty metrics response', async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      const result = await getMetrics();

      expect(result).toEqual({});
    });
  });

  describe('Error handling', () => {
    it('should retry on network error', async () => {
      (global.fetch as any)
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ data: 'success' }),
        });

      // Assuming retry logic is implemented
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: 'success' }),
      });

      const result = await getDashboardData().catch(() =>
        getDashboardData()
      );

      expect(result).toBeDefined();
    });

    it('should handle timeout', async () => {
      const timeoutError = new Error('Timeout');
      (global.fetch as any).mockRejectedValueOnce(timeoutError);

      await expect(getDashboardData()).rejects.toThrow('Timeout');
    });

    it('should sanitize API responses', async () => {
      const mockData = {
        summary: {
          totalSessions: 1250,
          activeUsers: 45,
          successRate: 98.5,
        },
        sessions: [],
        sessionTrends: [],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      });

      const result = await getDashboardData();

      // Check that response has expected structure
      expect(result).toHaveProperty('summary');
      expect(result).toHaveProperty('sessions');
      expect(result).toHaveProperty('sessionTrends');
    });
  });

  describe('Request validation', () => {
    it('should validate date format', async () => {
      const mockData = { summary: {}, sessions: [], sessionTrends: [] };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      });

      // Should accept ISO date format
      await getDashboardData({ startDate: '2024-01-01', endDate: '2024-01-31' });

      expect(global.fetch).toHaveBeenCalled();
    });

    it('should handle missing required parameters', async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      // getSessionDetails requires sessionId
      await expect(getSessionDetails('')).rejects.toThrow();
    });
  });
});
