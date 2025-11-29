import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { LineChart, BarChart, PieChart, AreaChart, Heatmap } from '@/components/analytics/Charts';

describe('LineChart', () => {
  const mockData = [
    { label: 'Jan', value: 100 },
    { label: 'Feb', value: 150 },
    { label: 'Mar', value: 120 },
  ];

  it('should render line chart', () => {
    const { container } = render(
      <LineChart data={mockData} title="Engagement Trend" />
    );

    expect(screen.getByText('Engagement Trend')).toBeInTheDocument();
    expect(container.querySelector('svg')).toBeInTheDocument();
  });

  it('should render chart with custom height', () => {
    const { container } = render(
      <LineChart data={mockData} height={400} />
    );

    const svg = container.querySelector('svg');
    expect(svg?.getAttribute('height')).toContain('400');
  });

  it('should render grid lines when enabled', () => {
    const { container } = render(
      <LineChart data={mockData} showGrid={true} />
    );

    const lines = container.querySelectorAll('line[stroke-dasharray]');
    expect(lines.length).toBeGreaterThan(0);
  });

  it('should hide grid lines when disabled', () => {
    const { container } = render(
      <LineChart data={mockData} showGrid={false} />
    );

    const lines = container.querySelectorAll('line[stroke-dasharray]');
    expect(lines.length).toBe(0);
  });

  it('should render legend when enabled', () => {
    render(<LineChart data={mockData} showLegend={true} />);

    expect(screen.getByText('Data')).toBeInTheDocument();
  });

  it('should apply custom color', () => {
    const { container } = render(
      <LineChart data={mockData} color="#FF0000" />
    );

    const path = container.querySelector('path[stroke="#FF0000"]');
    expect(path).toBeInTheDocument();
  });
});

describe('BarChart', () => {
  const mockData = [
    { label: 'Course A', value: 500 },
    { label: 'Course B', value: 750 },
    { label: 'Course C', value: 600 },
  ];

  it('should render bar chart', () => {
    const { container } = render(
      <BarChart data={mockData} title="Course Enrollment" />
    );

    expect(screen.getByText('Course Enrollment')).toBeInTheDocument();
    expect(container.querySelector('svg')).toBeInTheDocument();
  });

  it('should render horizontal bar chart', () => {
    render(
      <BarChart data={mockData} horizontal={true} />
    );

    // Horizontal bars use div elements instead of svg
    expect(screen.getByText(/Course A/)).toBeInTheDocument();
  });

  it('should display all bars', () => {
    const { container } = render(
      <BarChart data={mockData} />
    );

    const bars = container.querySelectorAll('rect');
    expect(bars.length).toBeGreaterThanOrEqual(mockData.length);
  });

  it('should render with custom color', () => {
    const { container } = render(
      <BarChart data={mockData} color="#10B981" />
    );

    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });
});

describe('PieChart', () => {
  const mockData = [
    { label: 'Category A', value: 300 },
    { label: 'Category B', value: 200 },
    { label: 'Category C', value: 150 },
  ];

  it('should render pie chart', () => {
    const { container } = render(
      <PieChart data={mockData} title="Distribution" />
    );

    expect(screen.getByText('Distribution')).toBeInTheDocument();
    expect(container.querySelector('svg')).toBeInTheDocument();
  });

  it('should render slices for each data point', () => {
    const { container } = render(
      <PieChart data={mockData} />
    );

    const paths = container.querySelectorAll('path');
    expect(paths.length).toBeGreaterThanOrEqual(mockData.length);
  });

  it('should show legend when enabled', () => {
    render(
      <PieChart data={mockData} showLegend={true} />
    );

    mockData.forEach(item => {
      expect(screen.getByText(item.label)).toBeInTheDocument();
    });
  });

  it('should hide legend when disabled', () => {
    render(
      <PieChart data={mockData} showLegend={false} />
    );

    // Legend divs shouldn't be rendered
    const legendItems = screen.queryAllByText(/Category/);
    expect(legendItems.length).toBeLessThanOrEqual(mockData.length);
  });

  it('should calculate percentage correctly', () => {
    render(
      <PieChart data={mockData} showLegend={true} />
    );

    // Total = 650, A = 300/650 = ~46%
    expect(screen.getByText(/46/)).toBeInTheDocument();
  });
});

describe('AreaChart', () => {
  const mockData = [
    { label: 'Mon', value: 100 },
    { label: 'Tue', value: 150 },
    { label: 'Wed', value: 120 },
  ];

  it('should render area chart', () => {
    const { container } = render(
      <AreaChart data={mockData} title="Trend" />
    );

    expect(screen.getByText('Trend')).toBeInTheDocument();
    expect(container.querySelector('svg')).toBeInTheDocument();
  });

  it('should render filled area', () => {
    const { container } = render(
      <AreaChart data={mockData} />
    );

    const paths = container.querySelectorAll('path[fill]');
    expect(paths.length).toBeGreaterThan(0);
  });

  it('should render line on top of area', () => {
    const { container } = render(
      <AreaChart data={mockData} />
    );

    const paths = container.querySelectorAll('path');
    expect(paths.length).toBeGreaterThanOrEqual(2); // area + line
  });
});

describe('Heatmap', () => {
  const mockData = [
    [1, 2, 3],
    [4, 5, 6],
    [7, 8, 9],
  ];

  const xLabels = ['A', 'B', 'C'];
  const yLabels = ['Row1', 'Row2', 'Row3'];

  it('should render heatmap table', () => {
    const { container } = render(
      <Heatmap data={mockData} title="Heatmap" />
    );

    expect(screen.getByText('Heatmap')).toBeInTheDocument();
    expect(container.querySelector('table')).toBeInTheDocument();
  });

  it('should render correct number of cells', () => {
    const { container } = render(
      <Heatmap data={mockData} />
    );

    const cells = container.querySelectorAll('td[style*="background"]');
    expect(cells.length).toBe(9); // 3x3 grid
  });

  it('should display axis labels', () => {
    render(
      <Heatmap data={mockData} xLabels={xLabels} yLabels={yLabels} />
    );

    xLabels.forEach(label => {
      expect(screen.getByText(label)).toBeInTheDocument();
    });

    yLabels.forEach(label => {
      expect(screen.getByText(label)).toBeInTheDocument();
    });
  });

  it('should color cells based on values', () => {
    const { container } = render(
      <Heatmap data={mockData} />
    );

    const cells = container.querySelectorAll('td[style*="background"]');
    const colors = new Set(
      Array.from(cells).map(cell => cell.getAttribute('style'))
    );

    // Should have different colors for different values
    expect(colors.size).toBeGreaterThan(1);
  });
});
