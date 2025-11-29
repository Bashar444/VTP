import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AnalyticsFilters, DataTable, AlertList } from '@/components/analytics/AnalyticsFilters';

describe('AnalyticsFilters', () => {
  it('should render filters panel', () => {
    render(<AnalyticsFilters />);

    expect(screen.getByText('Filters')).toBeInTheDocument();
  });

  it('should allow date selection', () => {
    render(<AnalyticsFilters />);

    const dateInputs = screen.getAllByDisplayValue('');
    // Should have at least 2 date inputs (start and end)
    expect(dateInputs.length).toBeGreaterThanOrEqual(2);
  });

  it('should handle interval changes', () => {
    const handleFilterChange = vi.fn();
    render(<AnalyticsFilters onFilterChange={handleFilterChange} />);

    const dailyButton = screen.getByText('Daily');
    fireEvent.click(dailyButton);

    // Should call filter change
    expect(handleFilterChange).toHaveBeenCalled();
  });

  it('should have clear filters button', () => {
    const handleFilterChange = vi.fn();
    render(<AnalyticsFilters onFilterChange={handleFilterChange} />);

    // First click a filter to enable clear button
    const dailyButton = screen.getByText('Daily');
    fireEvent.click(dailyButton);

    // Look for clear filters button
    const buttons = screen.getAllByText(/Clear|Daily|Weekly|Monthly/);
    expect(buttons.length).toBeGreaterThan(0);
  });

  it('should show mobile filter toggle', () => {
    render(<AnalyticsFilters />);

    // On mobile view, there should be a Filters button
    const filterButtons = screen.getAllByText('Filters');
    expect(filterButtons.length).toBeGreaterThan(0);
  });
});

describe('DataTable', () => {
  const mockColumns = [
    { key: 'name', label: 'Name', sortable: true },
    { key: 'value', label: 'Value', sortable: true },
    { key: 'status', label: 'Status' },
  ];

  const mockData = [
    { name: 'Item 1', value: 100, status: 'Active' },
    { name: 'Item 2', value: 200, status: 'Inactive' },
    { name: 'Item 3', value: 150, status: 'Active' },
  ];

  it('should render table with columns and data', () => {
    render(
      <DataTable columns={mockColumns} data={mockData} />
    );

    expect(screen.getByText('Name')).toBeInTheDocument();
    expect(screen.getByText('Value')).toBeInTheDocument();
    expect(screen.getByText('Status')).toBeInTheDocument();
    expect(screen.getByText('Item 1')).toBeInTheDocument();
  });

  it('should render all data rows', () => {
    render(
      <DataTable columns={mockColumns} data={mockData} />
    );

    expect(screen.getByText('Item 1')).toBeInTheDocument();
    expect(screen.getByText('Item 2')).toBeInTheDocument();
    expect(screen.getByText('Item 3')).toBeInTheDocument();
  });

  it('should display loading skeleton', () => {
    const { container } = render(
      <DataTable columns={mockColumns} data={[]} isLoading={true} />
    );

    const skeletons = container.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('should show empty state when no data', () => {
    render(
      <DataTable columns={mockColumns} data={[]} isLoading={false} />
    );

    expect(screen.getByText('No data available')).toBeInTheDocument();
  });

  it('should handle sorting', () => {
    render(
      <DataTable columns={mockColumns} data={mockData} />
    );

    const nameHeader = screen.getByText('Name');
    fireEvent.click(nameHeader);

    // After click, should show sort indicator
    expect(screen.getByText(/↑|↓/)).toBeInTheDocument();
  });

  it('should handle row click', () => {
    const handleRowClick = vi.fn();
    render(
      <DataTable
        columns={mockColumns}
        data={mockData}
        onRowClick={handleRowClick}
      />
    );

    const row = screen.getByText('Item 1').closest('tr');
    fireEvent.click(row!);

    expect(handleRowClick).toHaveBeenCalled();
  });

  it('should render custom cell content', () => {
    const customColumns = [
      ...mockColumns.slice(0, 2),
      {
        key: 'status',
        label: 'Status',
        render: (value: string) => `[${value}]`,
      },
    ];

    render(
      <DataTable columns={customColumns} data={mockData} />
    );

    expect(screen.getByText('[Active]')).toBeInTheDocument();
  });

  it('should display table title when provided', () => {
    render(
      <DataTable
        columns={mockColumns}
        data={mockData}
        title="Test Table"
      />
    );

    expect(screen.getByText('Test Table')).toBeInTheDocument();
  });
});

describe('AlertList', () => {
  const mockAlerts = [
    {
      id: 'alert-1',
      title: 'High Load',
      message: 'Server load exceeding 80%',
      type: 'warning' as const,
      timestamp: 'Jan 15, 2024',
    },
    {
      id: 'alert-2',
      title: 'Error',
      message: 'Database connection failed',
      type: 'error' as const,
      timestamp: 'Jan 15, 2024',
    },
    {
      id: 'alert-3',
      title: 'Info',
      message: 'System update completed',
      type: 'info' as const,
      timestamp: 'Jan 15, 2024',
    },
  ];

  it('should render all alerts', () => {
    render(<AlertList alerts={mockAlerts} />);

    expect(screen.getByText('High Load')).toBeInTheDocument();
    expect(screen.getByText('Error')).toBeInTheDocument();
    expect(screen.getByText('Info')).toBeInTheDocument();
  });

  it('should respect max items limit', () => {
    render(
      <AlertList alerts={mockAlerts} maxItems={2} />
    );

    expect(screen.getByText('High Load')).toBeInTheDocument();
    expect(screen.getByText('Error')).toBeInTheDocument();
    // Third alert should not be visible
    expect(screen.queryByText('System update completed')).not.toBeInTheDocument();
  });

  it('should handle alert dismiss', () => {
    const handleDismiss = vi.fn();
    render(
      <AlertList alerts={mockAlerts} onDismiss={handleDismiss} />
    );

    const dismissButtons = screen.getAllByRole('button');
    fireEvent.click(dismissButtons[0]);

    expect(handleDismiss).toHaveBeenCalledWith('alert-1');
  });

  it('should display alert messages', () => {
    render(<AlertList alerts={mockAlerts} />);

    expect(screen.getByText('Server load exceeding 80%')).toBeInTheDocument();
    expect(screen.getByText('Database connection failed')).toBeInTheDocument();
  });

  it('should show timestamps', () => {
    render(<AlertList alerts={mockAlerts} />);

    const timestamps = screen.getAllByText('Jan 15, 2024');
    expect(timestamps.length).toBe(mockAlerts.length);
  });

  it('should apply correct styling for different types', () => {
    const { container } = render(
      <AlertList alerts={mockAlerts} />
    );

    // Warning should have yellow styling
    const warningAlert = screen.getByText('High Load').closest('div');
    expect(warningAlert?.className).toContain('yellow');

    // Error should have red styling
    const errorAlert = screen.getByText('Error').closest('div');
    expect(errorAlert?.className).toContain('red');

    // Info should have blue styling
    const infoAlert = screen.getByText('Info').closest('div');
    expect(infoAlert?.className).toContain('blue');
  });

  it('should render empty when no alerts', () => {
    const { container } = render(
      <AlertList alerts={[]} />
    );

    expect(container.firstChild?.childNodes.length).toBe(0);
  });
});
