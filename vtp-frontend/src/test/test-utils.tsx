import { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { vi } from 'vitest';

// Mock providers wrapper
const AllTheProviders = ({ children }: { children: React.ReactNode }) => {
  return <>{children}</>;
};

const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>,
) => render(ui, { wrapper: AllTheProviders, ...options });

export * from '@testing-library/react';
export { customRender as render };

// Utilities for async operations
export const waitForAsync = () =>
  new Promise((resolve) => setTimeout(resolve, 0));

// Mock data generators
export const generateMockSession = (overrides = {}) => ({
  id: 'session-' + Math.random().toString(36).substr(2, 9),
  user: 'user@example.com',
  startTime: new Date().toISOString(),
  endTime: new Date(Date.now() + 900000).toISOString(),
  duration: 900,
  status: 'completed',
  ...overrides,
});

export const generateMockMetrics = (overrides = {}) => ({
  timestamp: new Date().toISOString(),
  sessions: 100,
  activeUsers: 45,
  avgDuration: 4.32,
  errorRate: 0.5,
  ...overrides,
});

export const generateMockDashboardData = (overrides = {}) => ({
  summary: {
    totalSessions: 1250,
    totalDuration: 5400,
    averageSessionDuration: 4.32,
    peakHour: '14:00',
    activeUsers: 45,
    successRate: 98.5,
  },
  sessions: [
    generateMockSession(),
    generateMockSession(),
  ],
  sessionTrends: [
    { timestamp: new Date().toISOString(), count: 50 },
    { timestamp: new Date(Date.now() - 3600000).toISOString(), count: 45 },
  ],
  ...overrides,
});

// Mock API response builder
export const mockApiResponse = <T,>(data: T, status = 200) => ({
  ok: status >= 200 && status < 300,
  status,
  json: async () => data,
  blob: async () => new Blob([JSON.stringify(data)]),
  text: async () => JSON.stringify(data),
  headers: new Map(),
});

// Wait for element with custom predicate
export const waitForElement = async (
  predicate: () => boolean,
  timeout = 3000
) => {
  const startTime = Date.now();
  while (Date.now() - startTime < timeout) {
    if (predicate()) {
      return true;
    }
    await waitForAsync();
  }
  throw new Error('Element not found within timeout');
};

// Fetch mock helpers
export const setupFetchMock = (responses: Record<string, any>) => {
  global.fetch = vi.fn((url: string) => {
    const key = Object.keys(responses).find((k) =>
      typeof url === 'string' ? url.includes(k) : false
    );
    if (key) {
      return Promise.resolve(mockApiResponse(responses[key]));
    }
    return Promise.reject(new Error(`No mock for ${url}`));
  });
};

// Performance testing utilities
export const measureComponentRender = (
  renderFn: () => void
): { duration: number; cleanup: () => void } => {
  const startTime = performance.now();
  const cleanup = renderFn();
  const duration = performance.now() - startTime;
  return { duration, cleanup: cleanup || (() => {}) };
};

// Export test expectations
export const expectSessionStructure = (session: any) => {
  expect(session).toHaveProperty('id');
  expect(session).toHaveProperty('user');
  expect(session).toHaveProperty('startTime');
  expect(session).toHaveProperty('endTime');
  expect(session).toHaveProperty('duration');
  expect(session).toHaveProperty('status');
};

export const expectMetricsStructure = (metrics: any) => {
  expect(metrics).toHaveProperty('timestamp');
  expect(metrics).toHaveProperty('sessions');
  expect(metrics).toHaveProperty('activeUsers');
  expect(metrics).toHaveProperty('avgDuration');
};

export const expectDashboardDataStructure = (data: any) => {
  expect(data).toHaveProperty('summary');
  expect(data).toHaveProperty('sessions');
  expect(data).toHaveProperty('sessionTrends');
  expect(Array.isArray(data.sessions)).toBe(true);
  expect(Array.isArray(data.sessionTrends)).toBe(true);
};

// Error simulation helpers
export const simulateNetworkError = () => {
  (global.fetch as any).mockRejectedValueOnce(
    new Error('Network error')
  );
};

export const simulateApiError = (status = 500) => {
  (global.fetch as any).mockResolvedValueOnce({
    ok: false,
    status,
    json: async () => ({ error: 'API Error' }),
  });
};

export const simulateTimeout = () => {
  (global.fetch as any).mockImplementationOnce(
    () => new Promise(() => {}) // Never resolves
  );
};
