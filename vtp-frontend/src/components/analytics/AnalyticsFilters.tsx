import React, { useState } from 'react';
import { Calendar, Filter, X } from 'lucide-react';
import { cn } from '@/utils/cn';

interface AnalyticsFilterProps {
  onFilterChange?: (filters: AnalyticsFilterState) => void;
  className?: string;
}

export interface AnalyticsFilterState {
  startDate?: string;
  endDate?: string;
  interval?: 'daily' | 'weekly' | 'monthly';
  courseId?: string;
  studentId?: string;
  category?: string;
}

export const AnalyticsFilters: React.FC<AnalyticsFilterProps> = ({
  onFilterChange,
  className,
}) => {
  const [filters, setFilters] = useState<AnalyticsFilterState>({});
  const [showMobileFilters, setShowMobileFilters] = useState(false);

  const handleFilterChange = (newFilters: AnalyticsFilterState) => {
    setFilters(newFilters);
    onFilterChange?.(newFilters);
  };

  const handleDateChange = (type: 'start' | 'end', value: string) => {
    const newFilters = {
      ...filters,
      [type === 'start' ? 'startDate' : 'endDate']: value,
    };
    handleFilterChange(newFilters);
  };

  const handleIntervalChange = (interval: 'daily' | 'weekly' | 'monthly') => {
    const newFilters = { ...filters, interval };
    handleFilterChange(newFilters);
  };

  const clearFilters = () => {
    setFilters({});
    onFilterChange?.({});
  };

  const intervals: Array<'daily' | 'weekly' | 'monthly'> = ['daily', 'weekly', 'monthly'];

  return (
    <div className={cn('space-y-4', className)}>
      {/* Desktop Filters */}
      <div className="hidden lg:block space-y-4 bg-gray-800 rounded-lg p-6">
        <h3 className="font-semibold text-white mb-4">Filters</h3>

        {/* Date Range */}
        <div className="space-y-2">
          <label className="text-sm text-gray-400">Date Range</label>
          <div className="flex gap-2">
            <input
              type="date"
              value={filters.startDate || ''}
              onChange={e => handleDateChange('start', e.target.value)}
              className="flex-1 px-3 py-2 bg-gray-700 text-white border border-gray-600 rounded text-sm focus:outline-none focus:border-blue-500"
            />
            <input
              type="date"
              value={filters.endDate || ''}
              onChange={e => handleDateChange('end', e.target.value)}
              className="flex-1 px-3 py-2 bg-gray-700 text-white border border-gray-600 rounded text-sm focus:outline-none focus:border-blue-500"
            />
          </div>
        </div>

        {/* Interval */}
        <div className="space-y-2">
          <label className="text-sm text-gray-400">Interval</label>
          <div className="flex gap-2">
            {intervals.map(interval => (
              <button
                key={interval}
                onClick={() => handleIntervalChange(interval)}
                className={cn(
                  'flex-1 py-2 px-3 rounded text-sm font-medium transition-colors',
                  filters.interval === interval
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                )}
              >
                {interval.charAt(0).toUpperCase() + interval.slice(1)}
              </button>
            ))}
          </div>
        </div>

        {/* Clear Filters */}
        {Object.keys(filters).length > 0 && (
          <button
            onClick={clearFilters}
            className="w-full py-2 px-4 bg-gray-700 hover:bg-gray-600 text-white text-sm rounded transition-colors"
          >
            Clear Filters
          </button>
        )}
      </div>

      {/* Mobile Filters Toggle */}
      <div className="lg:hidden">
        <button
          onClick={() => setShowMobileFilters(!showMobileFilters)}
          className="w-full py-2.5 px-4 bg-gray-800 hover:bg-gray-700 text-white font-semibold rounded-lg transition-colors flex items-center justify-center gap-2"
        >
          <Filter className="w-5 h-5" />
          Filters
        </button>

        {showMobileFilters && (
          <div className="mt-3 bg-gray-800 rounded-lg p-4 space-y-4">
            {/* Date Range */}
            <div className="space-y-2">
              <label className="text-sm text-gray-400">Start Date</label>
              <input
                type="date"
                value={filters.startDate || ''}
                onChange={e => handleDateChange('start', e.target.value)}
                className="w-full px-3 py-2 bg-gray-700 text-white border border-gray-600 rounded text-sm focus:outline-none focus:border-blue-500"
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm text-gray-400">End Date</label>
              <input
                type="date"
                value={filters.endDate || ''}
                onChange={e => handleDateChange('end', e.target.value)}
                className="w-full px-3 py-2 bg-gray-700 text-white border border-gray-600 rounded text-sm focus:outline-none focus:border-blue-500"
              />
            </div>

            {/* Interval */}
            <div className="space-y-2">
              <label className="text-sm text-gray-400">Interval</label>
              <div className="grid grid-cols-3 gap-2">
                {intervals.map(interval => (
                  <button
                    key={interval}
                    onClick={() => handleIntervalChange(interval)}
                    className={cn(
                      'py-2 px-3 rounded text-sm font-medium transition-colors',
                      filters.interval === interval
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                    )}
                  >
                    {interval.charAt(0).toUpperCase() + interval.slice(1)}
                  </button>
                ))}
              </div>
            </div>

            {/* Clear Filters */}
            {Object.keys(filters).length > 0 && (
              <button
                onClick={clearFilters}
                className="w-full py-2 px-4 bg-gray-700 hover:bg-gray-600 text-white text-sm rounded transition-colors"
              >
                Clear Filters
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

// Data Table Component
interface DataTableColumn<T> {
  key: keyof T | string;
  label: string;
  render?: (value: any, row: T) => React.ReactNode;
  sortable?: boolean;
  width?: string;
}

interface DataTableProps<T> {
  columns: DataTableColumn<T>[];
  data: T[];
  title?: string;
  isLoading?: boolean;
  onRowClick?: (row: T) => void;
  className?: string;
}

export const DataTable = React.forwardRef<HTMLDivElement, DataTableProps<any>>(
  ({ columns, data, title, isLoading = false, onRowClick, className }, ref) => {
    const [sortKey, setSortKey] = useState<string | null>(null);
    const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');

    const handleSort = (key: string) => {
      if (sortKey === key) {
        setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
      } else {
        setSortKey(key);
        setSortDir('asc');
      }
    };

    const sortedData = [...data].sort((a, b) => {
      if (!sortKey) return 0;

      const aVal = a[sortKey as keyof typeof a];
      const bVal = b[sortKey as keyof typeof b];

      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDir === 'asc' ? aVal - bVal : bVal - aVal;
      }

      const aStr = String(aVal).toLowerCase();
      const bStr = String(bVal).toLowerCase();
      return sortDir === 'asc' ? aStr.localeCompare(bStr) : bStr.localeCompare(aStr);
    });

    return (
      <div ref={ref} className={cn('bg-gray-800 rounded-lg overflow-hidden', className)}>
        {title && <div className="px-6 py-4 border-b border-gray-700">
          <h3 className="text-white font-semibold">{title}</h3>
        </div>}

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-700">
                {columns.map(col => (
                  <th
                    key={String(col.key)}
                    onClick={() => col.sortable && handleSort(String(col.key))}
                    className={cn(
                      'px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider',
                      col.sortable && 'cursor-pointer hover:text-white'
                    )}
                  >
                    <div className="flex items-center gap-2">
                      {col.label}
                      {col.sortable && sortKey === String(col.key) && (
                        <span className="text-white text-xs">{sortDir === 'asc' ? '↑' : '↓'}</span>
                      )}
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                [...Array(5)].map((_, i) => (
                  <tr key={i} className="border-b border-gray-700">
                    {columns.map(col => (
                      <td key={String(col.key)} className="px-6 py-4">
                        <div className="h-4 bg-gray-700 rounded animate-pulse" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : data.length === 0 ? (
                <tr>
                  <td colSpan={columns.length} className="px-6 py-8 text-center text-gray-400">
                    No data available
                  </td>
                </tr>
              ) : (
                sortedData.map((row, i) => (
                  <tr
                    key={i}
                    onClick={() => onRowClick?.(row)}
                    className={cn(
                      'border-b border-gray-700 hover:bg-gray-700/50 transition-colors',
                      onRowClick && 'cursor-pointer'
                    )}
                  >
                    {columns.map(col => (
                      <td key={String(col.key)} className="px-6 py-4 text-sm text-white">
                        {col.render
                          ? col.render(row[col.key as keyof typeof row], row)
                          : String(row[col.key as keyof typeof row])}
                      </td>
                    ))}
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    );
  }
);

DataTable.displayName = 'DataTable';

// Alert List Component
interface AlertItem {
  id: string;
  title: string;
  message: string;
  type: 'warning' | 'error' | 'info' | 'success';
  severity?: 'low' | 'medium' | 'high' | 'critical';
  timestamp: string;
}

interface AlertListProps {
  alerts: AlertItem[];
  onDismiss?: (id: string) => void;
  maxItems?: number;
  className?: string;
}

export const AlertList: React.FC<AlertListProps> = ({
  alerts,
  onDismiss,
  maxItems = 5,
  className,
}) => {
  const displayAlerts = alerts.slice(0, maxItems);

  const typeConfig = {
    warning: { bg: 'bg-yellow-900/20', border: 'border-yellow-700', icon: '⚠️' },
    error: { bg: 'bg-red-900/20', border: 'border-red-700', icon: '❌' },
    info: { bg: 'bg-blue-900/20', border: 'border-blue-700', icon: 'ℹ️' },
    success: { bg: 'bg-green-900/20', border: 'border-green-700', icon: '✓' },
  };

  return (
    <div className={cn('space-y-3', className)}>
      {displayAlerts.map(alert => {
        const config = typeConfig[alert.type];
        return (
          <div
            key={alert.id}
            className={cn(
              'border rounded-lg p-3 flex items-start gap-3',
              config.bg,
              config.border
            )}
          >
            <span className="text-lg flex-shrink-0">{config.icon}</span>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-white">{alert.title}</p>
              <p className="text-xs text-gray-400 mt-1">{alert.message}</p>
              <p className="text-xs text-gray-500 mt-1">{alert.timestamp}</p>
            </div>
            {onDismiss && (
              <button
                onClick={() => onDismiss(alert.id)}
                className="text-gray-400 hover:text-white transition-colors flex-shrink-0"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
        );
      })}
    </div>
  );
};
