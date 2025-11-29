import React, { useMemo } from 'react';
import { cn } from '@/utils/cn';

interface LineChartProps {
  data: Array<{ label: string; value: number }>;
  title?: string;
  height?: number;
  color?: string;
  showLegend?: boolean;
  showGrid?: boolean;
  className?: string;
}

export const LineChart: React.FC<LineChartProps> = ({
  data,
  title,
  height = 300,
  color = '#3B82F6',
  showLegend = true,
  showGrid = true,
  className,
}) => {
  const { maxValue, minValue } = useMemo(() => {
    const values = data.map(d => d.value);
    return {
      maxValue: Math.max(...values),
      minValue: Math.min(...values),
    };
  }, [data]);

  const range = maxValue - minValue || 1;
  const padding = 40;
  const width = 100;

  const points = data.map((d, i) => {
    const x = (i / Math.max(data.length - 1, 1)) * (width - padding * 2) + padding;
    const y = height - ((d.value - minValue) / range) * (height - padding * 2) - padding;
    return { x, y, ...d };
  });

  const pathD = points
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`)
    .join(' ');

  return (
    <div className={cn('bg-gray-800 rounded-lg p-6', className)}>
      {title && <h3 className="text-white font-semibold mb-4">{title}</h3>}

      <svg width="100%" height={height} viewBox={`0 0 ${width} ${height}`}>
        {/* Grid Lines */}
        {showGrid && (
          <>
            {[...Array(5)].map((_, i) => (
              <line
                key={`h-${i}`}
                x1="0"
                y1={(height / 5) * i}
                x2={width}
                y2={(height / 5) * i}
                stroke="#374151"
                strokeDasharray="4"
              />
            ))}
          </>
        )}

        {/* Path Line */}
        <path d={pathD} stroke={color} strokeWidth="2" fill="none" />

        {/* Points */}
        {points.map((p, i) => (
          <circle key={`point-${i}`} cx={p.x} cy={p.y} r="1.5" fill={color} />
        ))}
      </svg>

      {showLegend && (
        <div className="mt-4 flex items-center gap-2 text-sm">
          <div className="w-3 h-3 rounded-full" style={{ backgroundColor: color }} />
          <span className="text-gray-400">Data</span>
        </div>
      )}
    </div>
  );
};

interface BarChartProps {
  data: Array<{ label: string; value: number }>;
  title?: string;
  height?: number;
  color?: string;
  horizontal?: boolean;
  className?: string;
}

export const BarChart: React.FC<BarChartProps> = ({
  data,
  title,
  height = 300,
  color = '#3B82F6',
  horizontal = false,
  className,
}) => {
  const maxValue = Math.max(...data.map(d => d.value));
  const barWidth = 100 / data.length;

  return (
    <div className={cn('bg-gray-800 rounded-lg p-6', className)}>
      {title && <h3 className="text-white font-semibold mb-4">{title}</h3>}

      {horizontal ? (
        <div className="space-y-3">
          {data.map((d, i) => (
            <div key={i}>
              <div className="flex justify-between items-center mb-1">
                <span className="text-gray-400 text-xs truncate">{d.label}</span>
                <span className="text-white text-xs font-semibold">{d.value}</span>
              </div>
              <div className="w-full bg-gray-700 rounded-full h-2 overflow-hidden">
                <div
                  className="h-full transition-all duration-300"
                  style={{
                    width: `${(d.value / maxValue) * 100}%`,
                    backgroundColor: color,
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      ) : (
        <svg width="100%" height={height} viewBox={`0 0 100 ${height}`}>
          {data.map((d, i) => {
            const barHeight = (d.value / maxValue) * (height - 40);
            const x = (i + 0.5) * barWidth;
            const y = height - barHeight - 20;

            return (
              <g key={i}>
                <rect
                  x={x - barWidth / 2 + 2}
                  y={y}
                  width={barWidth - 4}
                  height={barHeight}
                  fill={color}
                  rx="2"
                  className="hover:opacity-80 transition-opacity"
                />
                <text
                  x={x}
                  y={height - 5}
                  textAnchor="middle"
                  fontSize="8"
                  fill="#9CA3AF"
                  className="pointer-events-none"
                >
                  {d.label}
                </text>
              </g>
            );
          })}
        </svg>
      )}
    </div>
  );
};

interface PieChartProps {
  data: Array<{ label: string; value: number }>;
  title?: string;
  size?: number;
  showLegend?: boolean;
  className?: string;
}

export const PieChart: React.FC<PieChartProps> = ({
  data,
  title,
  size = 200,
  showLegend = true,
  className,
}) => {
  const total = data.reduce((sum, d) => sum + d.value, 0);
  const colors = [
    '#3B82F6',
    '#10B981',
    '#F59E0B',
    '#EF4444',
    '#8B5CF6',
    '#EC4899',
    '#14B8A6',
    '#F97316',
  ];

  let currentAngle = -Math.PI / 2;
  const slices = data.map((d, i) => {
    const sliceAngle = (d.value / total) * Math.PI * 2;
    const startAngle = currentAngle;
    const endAngle = currentAngle + sliceAngle;

    const x1 = 100 + size * Math.cos(startAngle);
    const y1 = 100 + size * Math.sin(startAngle);
    const x2 = 100 + size * Math.cos(endAngle);
    const y2 = 100 + size * Math.sin(endAngle);

    const largeArc = sliceAngle > Math.PI ? 1 : 0;
    const pathD = `
      M 100 100
      L ${x1} ${y1}
      A ${size} ${size} 0 ${largeArc} 1 ${x2} ${y2}
      Z
    `;

    currentAngle = endAngle;

    return {
      pathD,
      color: colors[i % colors.length],
      label: d.label,
      value: d.value,
      percentage: ((d.value / total) * 100).toFixed(1),
    };
  });

  return (
    <div className={cn('bg-gray-800 rounded-lg p-6', className)}>
      {title && <h3 className="text-white font-semibold mb-4">{title}</h3>}

      <div className="flex flex-col lg:flex-row items-center justify-center gap-8">
        <svg width={size * 2.5} height={size * 2.5} viewBox="0 0 200 200">
          {slices.map((slice, i) => (
            <path
              key={i}
              d={slice.pathD}
              fill={slice.color}
              className="hover:opacity-80 transition-opacity"
            />
          ))}
        </svg>

        {showLegend && (
          <div className="space-y-2 text-sm">
            {slices.map((slice, i) => (
              <div key={i} className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: slice.color }} />
                <span className="text-gray-400">{slice.label}</span>
                <span className="text-white font-semibold ml-auto">{slice.percentage}%</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

interface AreaChartProps {
  data: Array<{ label: string; value: number }>;
  title?: string;
  height?: number;
  color?: string;
  className?: string;
}

export const AreaChart: React.FC<AreaChartProps> = ({
  data,
  title,
  height = 300,
  color = '#3B82F6',
  className,
}) => {
  const maxValue = Math.max(...data.map(d => d.value));
  const padding = 40;
  const width = 100;

  const points = data.map((d, i) => {
    const x = (i / Math.max(data.length - 1, 1)) * (width - padding * 2) + padding;
    const y = height - (d.value / maxValue) * (height - padding * 2) - padding;
    return { x, y };
  });

  const pathD = points
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`)
    .join(' ');

  const areaD = `${pathD} L ${points[points.length - 1].x} ${height - padding} L ${points[0].x} ${height - padding} Z`;

  return (
    <div className={cn('bg-gray-800 rounded-lg p-6', className)}>
      {title && <h3 className="text-white font-semibold mb-4">{title}</h3>}

      <svg width="100%" height={height} viewBox={`0 0 ${width} ${height}`}>
        {/* Area */}
        <path d={areaD} fill={color} opacity="0.1" />

        {/* Line */}
        <path d={pathD} stroke={color} strokeWidth="2" fill="none" />

        {/* Points */}
        {points.map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r="1.5" fill={color} />
        ))}
      </svg>
    </div>
  );
};

interface HeatmapProps {
  data: Array<Array<number>>;
  xLabels?: string[];
  yLabels?: string[];
  title?: string;
  className?: string;
}

export const Heatmap: React.FC<HeatmapProps> = ({
  data,
  xLabels,
  yLabels,
  title,
  className,
}) => {
  const flat = data.flat();
  const maxValue = Math.max(...flat);
  const minValue = Math.min(...flat);

  const getColor = (value: number) => {
    const ratio = (value - minValue) / (maxValue - minValue || 1);
    const hue = (1 - ratio) * 240;
    return `hsl(${hue}, 100%, 50%)`;
  };

  return (
    <div className={cn('bg-gray-800 rounded-lg p-6', className)}>
      {title && <h3 className="text-white font-semibold mb-4">{title}</h3>}

      <div className="overflow-x-auto">
        <table className="w-full text-xs">
          <tbody>
            {data.map((row, i) => (
              <tr key={i}>
                {yLabels && (
                  <td className="text-gray-400 pr-2 pb-1 text-right">{yLabels[i]}</td>
                )}
                {row.map((value, j) => (
                  <td
                    key={j}
                    className="w-8 h-8 border border-gray-700"
                    style={{ backgroundColor: getColor(value) }}
                    title={`${value}`}
                  />
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {xLabels && (
        <div className="flex gap-1 mt-2 ml-12 text-xs text-gray-400">
          {xLabels.map((label, i) => (
            <div key={i} className="w-8 text-center">
              {label}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

// Loading Skeleton for Charts
export const ChartSkeleton: React.FC<{ height?: number }> = ({ height = 300 }) => (
  <div className="bg-gray-800 rounded-lg p-6 animate-pulse">
    <div className="h-4 bg-gray-700 rounded w-1/4 mb-4" />
    <div className="w-full rounded" style={{ height: `${height}px`, background: '#374151' }} />
  </div>
);
