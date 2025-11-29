# Analytics Module Testing Guide

## Overview

This guide covers the comprehensive test suite for the analytics module, including unit tests, component tests, integration tests, and best practices.

## Test Structure

```
src/
├── components/
│   └── analytics/
│       ├── AnalyticsFilters.test.tsx      # Filter component tests
│       └── Dashboard.test.tsx              # Dashboard component tests
├── api/
│   └── analytics.test.ts                  # API endpoint tests
├── hooks/
│   └── analytics.test.ts                  # Custom hook tests
├── utils/
│   └── analytics.test.ts                  # Utility function tests
└── test/
    ├── setup.ts                           # Test environment setup
    └── test-utils.tsx                     # Testing utilities and helpers
```

## Running Tests

### Run all tests
```bash
npm test
# or
npm run test:watch
```

### Run specific test file
```bash
npm test -- AnalyticsFilters.test.tsx
```

### Run tests with coverage
```bash
npm run test:coverage
```

### Run tests in CI mode
```bash
npm run test:ci
```

## Test Categories

### 1. Component Tests (`AnalyticsFilters.test.tsx`, `Dashboard.test.tsx`)

Tests for React components including:
- Rendering and DOM structure
- User interactions (clicks, form inputs)
- Props handling and state management
- Loading and error states
- Responsive behavior

**Example Test:**
```typescript
it('should render filters panel', () => {
  render(<AnalyticsFilters />);
  expect(screen.getByText('Filters')).toBeInTheDocument();
});
```

### 2. API Tests (`api/analytics.test.ts`)

Tests for API calls including:
- Successful data fetching
- Error handling (network errors, API errors)
- Request validation
- Response parsing
- Retry logic

**Example Test:**
```typescript
it('should fetch dashboard data', async () => {
  const mockData = { summary: {}, sessions: [], sessionTrends: [] };
  (global.fetch as any).mockResolvedValueOnce({
    ok: true,
    json: async () => mockData,
  });

  const result = await getDashboardData();
  expect(result).toEqual(mockData);
});
```

### 3. Hook Tests (`hooks/analytics.test.ts`)

Tests for custom React hooks including:
- Data fetching and caching
- State management
- Side effects (timers, subscriptions)
- Loading/error states
- Refetching and refreshing

**Example Test:**
```typescript
it('should fetch analytics data on mount', async () => {
  const { result } = renderHook(() => useAnalyticsData());
  
  await waitFor(() => {
    expect(result.current.isLoading).toBe(false);
  });
  
  expect(result.current.data).toEqual(mockData);
});
```

### 4. Utility Tests (`utils/analytics.test.ts`)

Tests for pure utility functions including:
- Data transformation
- Formatting
- Calculations
- Data validation
- Data grouping and sorting

**Example Test:**
```typescript
it('should format duration to readable format', () => {
  expect(formatDuration(3661)).toBe('1 hour 1 minute 1 second');
});
```

## Testing Utilities

### Mock Generators

```typescript
import {
  generateMockSession,
  generateMockMetrics,
  generateMockDashboardData,
} from '@/test/test-utils';

// Generate realistic test data
const session = generateMockSession({ status: 'failed' });
const metrics = generateMockMetrics({ activeUsers: 100 });
const dashboard = generateMockDashboardData();
```

### Custom Render Function

```typescript
import { render } from '@/test/test-utils';

// Render component with all necessary providers
render(<MyComponent />);
```

### API Mocking

```typescript
import { setupFetchMock, mockApiResponse } from '@/test/test-utils';

setupFetchMock({
  '/api/analytics/dashboard': { summary: {}, sessions: [] },
  '/api/analytics/export': new Blob(['data']),
});
```

### Async Utilities

```typescript
import { waitForAsync, waitForElement } from '@/test/test-utils';

await waitForAsync(); // Wait for next tick

await waitForElement(
  () => screen.getByText('Loaded') !== null,
  3000 // timeout
);
```

### Error Simulation

```typescript
import {
  simulateNetworkError,
  simulateApiError,
  simulateTimeout,
} from '@/test/test-utils';

simulateNetworkError(); // Simulate network failure
simulateApiError(500);   // Simulate server error
simulateTimeout();       // Simulate timeout
```

## Best Practices

### 1. Test Organization

- One test file per component/module
- Descriptive test names that explain expected behavior
- Use `describe` blocks to group related tests

```typescript
describe('Dashboard', () => {
  describe('rendering', () => {
    it('should render dashboard with summary cards', () => {
      // test
    });
  });

  describe('data fetching', () => {
    it('should fetch data on mount', () => {
      // test
    });
  });
});
```

### 2. Mocking Strategy

- Mock external dependencies (API calls, localStorage)
- Don't mock the code being tested
- Use realistic test data

```typescript
// ✓ Good: Mock the API
vi.mock('@/api/analytics');

// ✗ Avoid: Mocking the component being tested
vi.mock('@/components/Dashboard');
```

### 3. Async Testing

- Use `waitFor` for async operations
- Set reasonable timeouts
- Avoid fixed delays (`setTimeout`)

```typescript
// ✓ Good: Wait for condition
await waitFor(() => {
  expect(screen.getByText('Loaded')).toBeInTheDocument();
});

// ✗ Avoid: Fixed delay
await new Promise(r => setTimeout(r, 1000));
```

### 4. User-Centric Testing

- Test user interactions, not implementation details
- Query by accessible elements (role, label, text)
- Avoid querying by CSS class or data-testid when possible

```typescript
// ✓ Good: Query by text visible to user
fireEvent.click(screen.getByText('Export'));

// ✗ Avoid: Query by implementation details
fireEvent.click(screen.getByTestId('export-btn'));
```

### 5. Error Handling

- Test both success and error scenarios
- Verify error messages are displayed
- Test graceful degradation

```typescript
describe('error handling', () => {
  it('should display error message on API failure', async () => {
    simulateApiError(500);
    render(<Dashboard />);
    
    await waitFor(() => {
      expect(screen.getByText(/Error|Failed/)).toBeInTheDocument();
    });
  });
});
```

### 6. Cleanup

- Automatic cleanup is handled in `setup.ts`
- Clear mocks between tests

```typescript
beforeEach(() => {
  vi.clearAllMocks();
});
```

## Common Testing Patterns

### Testing API Calls

```typescript
it('should call API with correct parameters', async () => {
  await getDashboardData({ startDate: '2024-01-01' });

  expect(global.fetch).toHaveBeenCalledWith(
    expect.stringContaining('startDate=2024-01-01'),
    expect.any(Object)
  );
});
```

### Testing Component State Changes

```typescript
it('should update filter when user selects option', () => {
  const { result } = renderHook(() => useFilters());

  act(() => {
    result.current.setInterval('weekly');
  });

  expect(result.current.interval).toBe('weekly');
});
```

### Testing Data Transformations

```typescript
it('should transform raw API response', () => {
  const raw = { sessions: [{ duration: 900 }] };
  const transformed = parseAnalyticsResponse(raw);

  expect(transformed.sessions[0].duration).toBe(900);
});
```

### Testing Conditional Rendering

```typescript
it('should show loading state initially', () => {
  render(<Dashboard />);
  
  expect(screen.getByText('Loading...')).toBeInTheDocument();
});

it('should show data after loading', async () => {
  render(<Dashboard />);
  
  await waitFor(() => {
    expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    expect(screen.getByText('1250')).toBeInTheDocument();
  });
});
```

## Coverage Goals

Aim for the following coverage targets:

- **Statements**: > 80%
- **Branches**: > 75%
- **Functions**: > 80%
- **Lines**: > 80%

Check coverage with:
```bash
npm run test:coverage
```

Coverage reports are generated in `coverage/` directory.

## Troubleshooting

### Tests timing out
- Increase timeout: `it('test', async () => {...}, 10000)`
- Check for unresolved promises
- Verify mocks are set up correctly

### Unexpected component updates
- Ensure cleanup is happening
- Check for memory leaks in effects
- Verify mock implementations

### Flaky tests
- Avoid time-dependent assertions
- Use `waitFor` instead of fixed delays
- Ensure deterministic test data

## CI/CD Integration

Tests should be run in CI with:
```bash
npm run test:ci
```

This runs tests once with coverage reporting and exits with proper status codes.

## Resources

- [Vitest Documentation](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [Testing Best Practices](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)
