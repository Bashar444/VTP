import { describe, it, expect } from 'vitest';
import {
  formatDuration,
  formatSessionTime,
  calculateMetrics,
  parseAnalyticsResponse,
  sanitizeSessionData,
  groupSessionsByDate,
  calculateSessionTrends,
} from '@/utils/analytics';

describe('Analytics Utilities', () => {
  describe('formatDuration', () => {
    it('should format seconds to readable duration', () => {
      expect(formatDuration(45)).toBe('45 seconds');
      expect(formatDuration(90)).toBe('1 minute 30 seconds');
      expect(formatDuration(3600)).toBe('1 hour');
      expect(formatDuration(3661)).toBe('1 hour 1 minute 1 second');
    });

    it('should handle zero duration', () => {
      expect(formatDuration(0)).toBe('0 seconds');
    });

    it('should handle large durations', () => {
      expect(formatDuration(86400)).toBe('1 day');
      expect(formatDuration(86461)).toMatch(/day.*hour.*minute.*second/);
    });

    it('should format decimal seconds', () => {
      const result = formatDuration(4.32 * 60); // 4.32 minutes
      expect(result).toContain('minute');
    });
  });

  describe('formatSessionTime', () => {
    it('should format ISO timestamp to readable format', () => {
      const isoTime = '2024-01-15T10:30:45Z';
      const result = formatSessionTime(isoTime);

      expect(result).toMatch(/Jan|15|10:30|45/);
    });

    it('should handle different date formats', () => {
      const timestamp = '2024-01-15T10:30:45Z';
      const result = formatSessionTime(timestamp, 'short');

      expect(result.length).toBeLessThan('Monday, January 15, 2024 10:30:45 AM'.length);
    });

    it('should format relative time', () => {
      const pastTime = new Date(Date.now() - 3600000).toISOString(); // 1 hour ago
      const result = formatSessionTime(pastTime, 'relative');

      expect(result).toMatch(/ago|hour/i);
    });
  });

  describe('calculateMetrics', () => {
    it('should calculate session metrics', () => {
      const sessions = [
        { duration: 900, status: 'completed' },
        { duration: 1200, status: 'completed' },
        { duration: 600, status: 'failed' },
      ];

      const metrics = calculateMetrics(sessions);

      expect(metrics).toHaveProperty('totalSessions');
      expect(metrics).toHaveProperty('averageDuration');
      expect(metrics).toHaveProperty('successCount');
      expect(metrics.totalSessions).toBe(3);
      expect(metrics.averageDuration).toBeCloseTo(900);
    });

    it('should handle empty sessions', () => {
      const metrics = calculateMetrics([]);

      expect(metrics.totalSessions).toBe(0);
      expect(metrics.averageDuration).toBe(0);
      expect(metrics.successCount).toBe(0);
    });

    it('should calculate success rate', () => {
      const sessions = [
        { duration: 900, status: 'completed' },
        { duration: 1200, status: 'completed' },
        { duration: 600, status: 'failed' },
        { duration: 300, status: 'failed' },
      ];

      const metrics = calculateMetrics(sessions);

      expect(metrics.successRate).toBe(50); // 2 out of 4
    });

    it('should identify peak metrics', () => {
      const sessions = [
        { duration: 900, status: 'completed', timestamp: '2024-01-15T10:00:00Z' },
        { duration: 1200, status: 'completed', timestamp: '2024-01-15T10:00:00Z' },
        { duration: 600, status: 'completed', timestamp: '2024-01-15T14:00:00Z' },
      ];

      const metrics = calculateMetrics(sessions);

      expect(metrics).toHaveProperty('peakHour');
    });
  });

  describe('parseAnalyticsResponse', () => {
    it('should parse valid analytics response', () => {
      const response = {
        summary: { totalSessions: 100, activeUsers: 50 },
        sessions: [{ id: '1', duration: 900 }],
        sessionTrends: [{ timestamp: '2024-01-15T10:00:00Z', count: 10 }],
      };

      const parsed = parseAnalyticsResponse(response);

      expect(parsed).toEqual(response);
    });

    it('should validate required fields', () => {
      const invalidResponse = {
        summary: { totalSessions: 100 },
        // Missing sessions and sessionTrends
      };

      expect(() => {
        parseAnalyticsResponse(invalidResponse);
      }).toThrow();
    });

    it('should handle and normalize data types', () => {
      const response = {
        summary: { totalSessions: '100', activeUsers: 50 },
        sessions: [],
        sessionTrends: [],
      };

      const parsed = parseAnalyticsResponse(response);

      expect(typeof parsed.summary.totalSessions).toBe('number');
    });

    it('should handle missing optional fields gracefully', () => {
      const response = {
        summary: { totalSessions: 100 },
        sessions: [],
        sessionTrends: [],
      };

      const parsed = parseAnalyticsResponse(response);

      expect(parsed.summary.totalSessions).toBe(100);
      expect(parsed.sessions).toEqual([]);
    });
  });

  describe('sanitizeSessionData', () => {
    it('should remove sensitive data', () => {
      const session = {
        id: '123',
        user: 'user@example.com',
        apiKey: 'secret-key',
        token: 'bearer-token',
        duration: 900,
      };

      const sanitized = sanitizeSessionData(session);

      expect(sanitized).not.toHaveProperty('apiKey');
      expect(sanitized).not.toHaveProperty('token');
      expect(sanitized).toHaveProperty('id');
      expect(sanitized).toHaveProperty('duration');
    });

    it('should handle nested sensitive data', () => {
      const session = {
        id: '123',
        user: 'user@example.com',
        config: {
          password: 'secret',
          apiKey: 'key',
        },
        duration: 900,
      };

      const sanitized = sanitizeSessionData(session);

      expect(sanitized.config).not.toHaveProperty('password');
      expect(sanitized.config).not.toHaveProperty('apiKey');
    });

    it('should preserve non-sensitive data', () => {
      const session = {
        id: '123',
        user: 'user@example.com',
        status: 'completed',
        duration: 900,
        startTime: '2024-01-15T10:00:00Z',
        endTime: '2024-01-15T10:15:00Z',
      };

      const sanitized = sanitizeSessionData(session);

      expect(sanitized.id).toBe('123');
      expect(sanitized.user).toBe('user@example.com');
      expect(sanitized.status).toBe('completed');
      expect(sanitized.duration).toBe(900);
    });
  });

  describe('groupSessionsByDate', () => {
    it('should group sessions by date', () => {
      const sessions = [
        { id: '1', timestamp: '2024-01-15T10:00:00Z', duration: 900 },
        { id: '2', timestamp: '2024-01-15T14:00:00Z', duration: 1200 },
        { id: '3', timestamp: '2024-01-16T10:00:00Z', duration: 600 },
      ];

      const grouped = groupSessionsByDate(sessions);

      expect(grouped).toHaveProperty('2024-01-15');
      expect(grouped).toHaveProperty('2024-01-16');
      expect(grouped['2024-01-15']).toHaveLength(2);
      expect(grouped['2024-01-16']).toHaveLength(1);
    });

    it('should handle empty sessions', () => {
      const grouped = groupSessionsByDate([]);

      expect(Object.keys(grouped)).toHaveLength(0);
    });

    it('should handle different timestamp formats', () => {
      const sessions = [
        { id: '1', timestamp: '2024-01-15T10:00:00Z', duration: 900 },
        { id: '2', timestamp: '2024-01-15 14:30:00', duration: 1200 },
      ];

      const grouped = groupSessionsByDate(sessions);

      expect(Object.keys(grouped).length).toBeGreaterThan(0);
    });
  });

  describe('calculateSessionTrends', () => {
    it('should calculate session trends', () => {
      const sessions = [
        { id: '1', timestamp: '2024-01-15T10:00:00Z' },
        { id: '2', timestamp: '2024-01-15T10:00:00Z' },
        { id: '3', timestamp: '2024-01-15T11:00:00Z' },
      ];

      const trends = calculateSessionTrends(sessions, 'hourly');

      expect(trends).toHaveLength(2);
      expect(trends[0].count).toBe(2);
      expect(trends[1].count).toBe(1);
    });

    it('should calculate daily trends', () => {
      const sessions = [
        { id: '1', timestamp: '2024-01-15T10:00:00Z' },
        { id: '2', timestamp: '2024-01-15T14:00:00Z' },
        { id: '3', timestamp: '2024-01-16T10:00:00Z' },
      ];

      const trends = calculateSessionTrends(sessions, 'daily');

      expect(trends).toHaveLength(2);
      expect(trends[0].count).toBe(2);
      expect(trends[1].count).toBe(1);
    });

    it('should handle empty sessions', () => {
      const trends = calculateSessionTrends([], 'daily');

      expect(trends).toEqual([]);
    });

    it('should format trend data correctly', () => {
      const sessions = [
        { id: '1', timestamp: '2024-01-15T10:00:00Z' },
        { id: '2', timestamp: '2024-01-15T10:00:00Z' },
      ];

      const trends = calculateSessionTrends(sessions, 'hourly');

      expect(trends[0]).toHaveProperty('timestamp');
      expect(trends[0]).toHaveProperty('count');
      expect(typeof trends[0].timestamp).toBe('string');
      expect(typeof trends[0].count).toBe('number');
    });
  });
});
