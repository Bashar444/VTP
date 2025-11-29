# Analytics Module Test Suite

Comprehensive test suite for the VTP frontend analytics module.

## Quick Start

```bash
# Install dependencies
npm install

# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with UI dashboard
npm run test:ui

# Generate coverage report
npm run test:coverage

# Run tests in CI mode
npm run test:ci
```

## Test Files

### Components
- **`src/components/analytics/AnalyticsFilters.test.tsx`** - Filter UI component tests
- **`src/components/analytics/Dashboard.test.tsx`** - Dashboard component tests

### API
- **`src/api/analytics.test.ts`** - API endpoint tests

### Hooks
- **`src/hooks/analytics.test.ts`** - Custom React hooks tests

### Utilities
- **`src/utils/analytics.test.ts`** - Utility function tests

### Integration
- **`src/test/integration.test.ts`** - End-to-end integration tests

## Test Coverage

Current coverage targets:
- **Statements**: > 80%
- **Branches**: > 75%
- **Functions**: > 80%
- **Lines**: > 80%

## Key Features

✅ **Component Testing**
- React component rendering
- User interactions
- Props and state management
- Loading/error states

✅ **API Testing**
- Fetch operations
- Error handling
- Request/response validation
- Mocking strategies

✅ **Hook Testing**
- Custom React hooks
- Data fetching
- State management
- Side effects

✅ **Utility Testing**
- Data transformation
- Formatting functions
- Calculations
- Data validation

✅ **Integration Testing**
- Cross-component workflows
- Filter to export flow
- Error recovery
- Performance optimization

## Test Utilities

Helper functions and mock data generators:

```typescript
import {
  render,                      // Custom render with providers
  generateMockSession,         // Generate mock session data
  generateMockMetrics,         // Generate mock metrics
  generateMockDashboardData,   // Generate full dashboard data
  setupFetchMock,              // Mock fetch globally
  simulateNetworkError,        // Simulate network failure
  simulateApiError,            // Simulate API error
  waitForElement,              // Wait for elements
} from '@/test/test-utils';
```

## Configuration

- **`vitest.config.ts`** - Vitest configuration
- **`src/test/setup.ts`** - Test environment setup
- **`src/test/test-utils.tsx`** - Testing utilities

## Common Commands

```bash
# Run specific test file
npm test -- AnalyticsFilters.test.tsx

# Run tests matching pattern
npm test -- --grep "should render"

# Run with coverage for specific file
npm test -- --coverage src/components/analytics/Dashboard.test.tsx

# Run test and keep watching
npm test -- --watch

# Generate detailed coverage report
npm run test:coverage

# Run tests with UI (interactive dashboard)
npm run test:ui
```

## Writing Tests

### Basic Component Test
```typescript
it('should render component', () => {
  render(<MyComponent />);
  expect(screen.getByText('Expected text')).toBeInTheDocument();
});
```

### Async Component Test
```typescript
it('should load data', async () => {
  render(<MyComponent />);
  await waitFor(() => {
    expect(screen.getByText('Loaded')).toBeInTheDocument();
  });
});
```

### API Test
```typescript
it('should fetch data', async () => {
  (global.fetch as any).mockResolvedValueOnce({
    ok: true,
    json: async () => mockData,
  });
  
  const result = await apiCall();
  expect(result).toEqual(mockData);
});
```

### Hook Test
```typescript
it('should use hook', async () => {
  const { result } = renderHook(() => useMyHook());
  await waitFor(() => {
    expect(result.current.data).toBeDefined();
  });
});
```

## Debugging Tests

### Visual Debugging
```bash
# Run tests with UI
npm run test:ui
```

### Debug Single Test
```bash
npm test -- --grep "test name" --inspect-brk
```

### Print DOM
```typescript
import { render, screen } from '@testing-library/react';

render(<Component />);
screen.debug(); // Print full DOM
```

## Best Practices

1. **Use Descriptive Names** - Test names should explain expected behavior
2. **One Assertion Per Concept** - Group related assertions
3. **Mock External Dependencies** - Don't mock code being tested
4. **Use User Events** - Test from user perspective
5. **Avoid Implementation Details** - Test behavior, not internals
6. **Clean Up** - Mocks are automatically cleared

## CI/CD Integration

Tests run automatically in CI pipeline:

```yaml
test:
  script:
    - npm ci
    - npm run test:ci
```

## Resources

- [Vitest Docs](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [Testing Best Practices](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)
- [TESTING_GUIDE.md](../TESTING_GUIDE.md) - Comprehensive testing guide

## Support

For issues or questions:
1. Check [TESTING_GUIDE.md](../TESTING_GUIDE.md) for detailed documentation
2. Review existing test examples
3. Check test output and error messages
4. Consult framework documentation
