import React from 'react';
import { TrendingUp, TrendingDown, Activity } from 'lucide-react';
import { cn } from '@/utils/cn';

interface AnalyticsCardProps {
  title: string;
  value: string | number;
  suffix?: string;
  icon?: React.ReactNode;
  trend?: {
    value: number;
    isPositive: boolean;
    period: string;
  };
  backgroundColor?: string;
  textColor?: string;
  trendColor?: string;
  onClick?: () => void;
  isLoading?: boolean;
  className?: string;
}

export const AnalyticsCard: React.FC<AnalyticsCardProps> = ({
  title,
  value,
  suffix,
  icon,
  trend,
  backgroundColor = 'bg-gray-800',
  textColor = 'text-white',
  trendColor,
  onClick,
  isLoading = false,
  className,
}) => {
  return (
    <div
      onClick={onClick}
      className={cn(
        'rounded-lg p-6 cursor-pointer transition-all hover:shadow-lg hover:scale-105',
        backgroundColor,
        onClick && 'hover:ring-2 hover:ring-blue-500',
        className
      )}
    >
      <div className="flex items-start justify-between mb-4">
        <div>
          <p className="text-gray-400 text-sm font-medium">{title}</p>
          {isLoading ? (
            <div className="h-8 bg-gray-700 rounded animate-pulse mt-2" />
          ) : (
            <p className={cn('text-3xl font-bold mt-1', textColor)}>
              {value}
              {suffix && <span className="text-xl ml-1">{suffix}</span>}
            </p>
          )}
        </div>
        {icon && <div className="text-gray-500">{icon}</div>}
      </div>

      {trend && (
        <div className="flex items-center gap-2 pt-4 border-t border-gray-700">
          {trend.isPositive ? (
            <TrendingUp className={cn('w-4 h-4', trendColor || 'text-green-500')} />
          ) : (
            <TrendingDown className={cn('w-4 h-4', trendColor || 'text-red-500')} />
          )}
          <span
            className={cn(
              'text-sm font-semibold',
              trendColor || (trend.isPositive ? 'text-green-500' : 'text-red-500')
            )}
          >
            {trend.isPositive ? '+' : ''}{trend.value}%
          </span>
          <span className="text-gray-400 text-xs">{trend.period}</span>
        </div>
      )}
    </div>
  );
};

interface StatGridProps {
  stats: Array<{
    title: string;
    value: string | number;
    suffix?: string;
    icon?: React.ReactNode;
    trend?: {
      value: number;
      isPositive: boolean;
      period: string;
    };
    backgroundColor?: string;
  }>;
  columns?: number;
  isLoading?: boolean;
  onCardClick?: (index: number) => void;
  className?: string;
}

export const StatGrid: React.FC<StatGridProps> = ({
  stats,
  columns = 4,
  isLoading = false,
  onCardClick,
  className,
}) => {
  const gridClass = {
    1: 'grid-cols-1',
    2: 'grid-cols-1 md:grid-cols-2',
    3: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4',
  }[columns] || 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4';

  return (
    <div className={cn('grid gap-6', gridClass, className)}>
      {stats.map((stat, index) => (
        <AnalyticsCard
          key={index}
          title={stat.title}
          value={stat.value}
          suffix={stat.suffix}
          icon={stat.icon}
          trend={stat.trend}
          backgroundColor={stat.backgroundColor}
          isLoading={isLoading}
          onClick={() => onCardClick?.(index)}
        />
      ))}
    </div>
  );
};

// Metric Summary Component
interface MetricSummaryProps {
  label: string;
  value: string | number;
  change?: number;
  isPositive?: boolean;
  unit?: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export const MetricSummary: React.FC<MetricSummaryProps> = ({
  label,
  value,
  change,
  isPositive = true,
  unit,
  size = 'md',
  className,
}) => {
  const sizeClass = {
    sm: 'text-sm',
    md: 'text-base',
    lg: 'text-lg',
  }[size];

  const valueClass = {
    sm: 'text-2xl',
    md: 'text-3xl',
    lg: 'text-4xl',
  }[size];

  return (
    <div className={cn('space-y-1', className)}>
      <p className="text-gray-400 text-xs uppercase tracking-wider">{label}</p>
      <div className="flex items-baseline gap-2">
        <p className={cn('font-bold text-white', valueClass)}>
          {value}
          {unit && <span className="text-sm ml-1">{unit}</span>}
        </p>
        {change !== undefined && (
          <span
            className={cn(
              'text-sm font-semibold flex items-center gap-1',
              isPositive ? 'text-green-500' : 'text-red-500'
            )}
          >
            {isPositive ? '↑' : '↓'} {Math.abs(change)}%
          </span>
        )}
      </div>
    </div>
  );
};

// Key Insight Card
interface InsightCardProps {
  title: string;
  description: string;
  icon?: React.ReactNode;
  actionLabel?: string;
  onAction?: () => void;
  variant?: 'info' | 'success' | 'warning' | 'danger';
  className?: string;
}

export const InsightCard: React.FC<InsightCardProps> = ({
  title,
  description,
  icon,
  actionLabel,
  onAction,
  variant = 'info',
  className,
}) => {
  const variantClasses = {
    info: 'bg-blue-900/30 border-blue-700 text-blue-100',
    success: 'bg-green-900/30 border-green-700 text-green-100',
    warning: 'bg-amber-900/30 border-amber-700 text-amber-100',
    danger: 'bg-red-900/30 border-red-700 text-red-100',
  };

  const iconColors = {
    info: 'text-blue-400',
    success: 'text-green-400',
    warning: 'text-amber-400',
    danger: 'text-red-400',
  };

  return (
    <div
      className={cn(
        'border rounded-lg p-4 space-y-3',
        variantClasses[variant],
        className
      )}
    >
      <div className="flex items-start gap-3">
        {icon && <div className={cn('flex-shrink-0 mt-1', iconColors[variant])}>{icon}</div>}
        <div className="flex-1">
          <h3 className="font-semibold text-sm">{title}</h3>
          <p className="text-xs opacity-90 mt-1">{description}</p>
        </div>
      </div>

      {actionLabel && onAction && (
        <button
          onClick={onAction}
          className={cn(
            'text-xs font-semibold py-2 px-3 rounded transition-colors w-full',
            variant === 'info' && 'bg-blue-700 hover:bg-blue-600 text-white',
            variant === 'success' && 'bg-green-700 hover:bg-green-600 text-white',
            variant === 'warning' && 'bg-amber-700 hover:bg-amber-600 text-white',
            variant === 'danger' && 'bg-red-700 hover:bg-red-600 text-white'
          )}
        >
          {actionLabel}
        </button>
      )}
    </div>
  );
};

// Loading Skeleton for Analytics Card
export const AnalyticsCardSkeleton: React.FC<{ count?: number }> = ({ count = 4 }) => {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      {[...Array(count)].map((_, i) => (
        <div key={i} className="bg-gray-800 rounded-lg p-6 animate-pulse">
          <div className="h-4 bg-gray-700 rounded w-1/3 mb-4" />
          <div className="h-8 bg-gray-700 rounded w-2/3 mb-4" />
          <div className="h-3 bg-gray-700 rounded w-1/2" />
        </div>
      ))}
    </div>
  );
};

// Performance Indicator
interface PerformanceIndicatorProps {
  label: string;
  value: number;
  target?: number;
  unit?: string;
  color?: 'green' | 'blue' | 'yellow' | 'red';
  className?: string;
}

export const PerformanceIndicator: React.FC<PerformanceIndicatorProps> = ({
  label,
  value,
  target,
  unit = '',
  color = 'blue',
  className,
}) => {
  const percentage = target ? (value / target) * 100 : value;
  const isGood = percentage >= 80;
  const isOkay = percentage >= 50;

  const colorClass = {
    green: 'bg-green-600',
    blue: 'bg-blue-600',
    yellow: 'bg-yellow-600',
    red: 'bg-red-600',
  }[isGood ? 'green' : isOkay ? 'yellow' : 'red'];

  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex justify-between items-center">
        <span className="text-gray-400 text-sm">{label}</span>
        <span className="text-white font-semibold">
          {value.toFixed(1)}{unit}
          {target && <span className="text-gray-400 text-sm"> / {target}{unit}</span>}
        </span>
      </div>
      <div className="w-full bg-gray-700 rounded-full h-2 overflow-hidden">
        <div
          className={cn('h-full transition-all duration-500', colorClass)}
          style={{ width: `${Math.min(percentage, 100)}%` }}
        />
      </div>
    </div>
  );
};
