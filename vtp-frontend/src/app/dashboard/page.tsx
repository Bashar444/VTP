'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/store';
import { AnalyticsService } from '@/services/analytics.service';
import { AnalyticsCard, StatGrid, InsightCard } from '@/components/analytics/AnalyticsCard';
import { LineChart, BarChart, PieChart } from '@/components/analytics/Charts';
import { AnalyticsFilters, DataTable, AlertList } from '@/components/analytics/AnalyticsFilters';
import type {
  DashboardMetrics,
  EngagementMetrics,
  CoursePerformance,
  SystemAlert,
} from '@/services/analytics.service';
import { TrendingUp, Users, BookOpen, AlertCircle, DollarSign, Activity } from 'lucide-react';
import NetworkStatus from '@/components/NetworkStatus';
import QualitySelector from '@/components/QualitySelector';
import EdgeNodeViewer from '@/components/EdgeNodeViewer';
import MetricsDisplay from '@/components/MetricsDisplay';
import { g5Service } from '@/services/g5Service';

export default function DashboardPage() {
  const router = useRouter();
  const { user } = useAuth();
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [engagement, setEngagement] = useState<EngagementMetrics[]>([]);
  const [courses, setCourses] = useState<CoursePerformance[]>([]);
  const [alerts, setAlerts] = useState<SystemAlert[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [globalQuality, setGlobalQuality] = useState<number | null>(null);
  const [networkQualityLoading, setNetworkQualityLoading] = useState(false);

  // Redirect if not authenticated
  useEffect(() => {
    if (!user) {
      router.push('/auth/login');
    }
  }, [user, router]);

  // Fetch analytics + network quality data
  useEffect(() => {
    const fetchData = async () => {
      if (!user) return;

      try {
        setIsLoading(true);
        const [dashMetrics, engagementData, courseData, systemAlerts] = await Promise.all([
          AnalyticsService.getDashboardMetrics(),
          AnalyticsService.getEngagementMetrics(undefined, undefined, 'daily'),
          AnalyticsService.getTopPerformingCourses(10),
          AnalyticsService.getSystemAlerts(true),
        ]);

        setMetrics(dashMetrics);
        setEngagement(engagementData);
        setCourses(courseData);
        setAlerts(systemAlerts);
        // Also fetch network quality
        try {
          setNetworkQualityLoading(true);
          const quality = await g5Service.getNetworkQuality();
          setGlobalQuality(quality);
        } catch (e) {
          // Non-blocking
          console.warn('Failed to fetch network quality');
        } finally {
          setNetworkQualityLoading(false);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load analytics');
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [user]);

  if (error) {
    return (
      <div className="min-h-screen bg-gray-900 pt-24 pb-12">
        <div className="container mx-auto px-4">
          <div className="bg-red-900/20 border border-red-700 rounded-lg p-6">
            <p className="text-red-400">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        {/* Page Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-white mb-2">Analytics Dashboard</h1>
          <p className="text-gray-400">Real-time insights and performance metrics</p>
        </div>

        {/* Filters */}
        <div className="mb-8">
          <AnalyticsFilters />
        </div>

        {/* Key Metrics Grid */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-white mb-4">Key Metrics</h2>
          <StatGrid
            stats={[
              {
                title: 'Total Students',
                value: metrics?.totalStudents || 0,
                icon: <Users className="w-6 h-6" />,
                trend: {
                  value: 12,
                  isPositive: true,
                  period: 'vs last month',
                },
              },
              {
                title: 'Active Courses',
                value: metrics?.totalCourses || 0,
                icon: <BookOpen className="w-6 h-6" />,
                trend: {
                  value: 5,
                  isPositive: true,
                  period: 'added this month',
                },
              },
              {
                title: 'Active Users',
                value: metrics?.activeUsers || 0,
                icon: <Activity className="w-6 h-6" />,
                trend: {
                  value: 8,
                  isPositive: true,
                  period: 'growth this week',
                },
              },
              {
                title: 'Total Revenue',
                value: `$${(metrics?.totalRevenue || 0).toLocaleString()}`,
                icon: <DollarSign className="w-6 h-6" />,
                trend: {
                  value: 23,
                  isPositive: true,
                  period: 'vs last month',
                },
              },
              {
                title: 'Network Quality',
                value: networkQualityLoading ? '…' : (globalQuality !== null ? `${globalQuality}%` : 'N/A'),
                icon: <Activity className="w-6 h-6" />,
                trend: {
                  value: 0,
                  isPositive: true,
                  period: 'real-time',
                },
              },
            ]}
            isLoading={isLoading}
            columns={5}
          />
        </div>

        {/* System Alerts */}
        {alerts.length > 0 && (
          <div className="mb-8">
            <h2 className="text-2xl font-bold text-white mb-4">System Alerts</h2>
            <AlertList
              alerts={alerts.map(a => ({
                id: a.id,
                title: a.title,
                message: a.message,
                type: a.type,
                severity: a.severity,
                timestamp: new Date(a.createdAt).toLocaleDateString(),
              }))}
              maxItems={5}
            />
          </div>
        )}

        {/* Charts Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
          {/* Engagement Trend */}
          <div>
            <h2 className="text-2xl font-bold text-white mb-4">Engagement Trend</h2>
            <LineChart
              data={
                engagement.length > 0
                  ? engagement.map(e => ({
                      label: new Date(e.date).toLocaleDateString(),
                      value: e.activeUsers,
                    }))
                  : []
              }
              color="#3B82F6"
              height={300}
            />
          </div>

          {/* Watch Time Distribution */}
          <div>
            <h2 className="text-2xl font-bold text-white mb-4">Watch Time by Course</h2>
            <BarChart
              data={
                courses.length > 0
                  ? courses.slice(0, 5).map(c => ({
                      label: c.courseName.slice(0, 15),
                      value: c.totalWatchTime,
                    }))
                  : []
              }
              color="#10B981"
              height={300}
            />
          </div>
        </div>

        {/* Performance Metrics */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
          {/* Completion Rates */}
          <div>
            <h2 className="text-2xl font-bold text-white mb-4">Completion Rates</h2>
            <PieChart
              data={
                courses.length > 0
                  ? courses.slice(0, 4).map(c => ({
                      label: c.courseName.slice(0, 12),
                      value: c.completionRate,
                    }))
                  : []
              }
              size={100}
              showLegend={true}
            />
          </div>

          {/* Top Performing Courses */}
          <div className="lg:col-span-2">
            <h2 className="text-2xl font-bold text-white mb-4">Top Performing Courses</h2>
            <DataTable
              columns={[
                {
                  key: 'courseName',
                  label: 'Course',
                  sortable: true,
                },
                {
                  key: 'enrollmentCount',
                  label: 'Enrollments',
                  sortable: true,
                },
                {
                  key: 'completionRate',
                  label: 'Completion %',
                  render: (value: number) => `${value.toFixed(1)}%`,
                  sortable: true,
                },
                {
                  key: 'avgRating',
                  label: 'Rating',
                  render: (value: number) => `${value.toFixed(1)}/5`,
                  sortable: true,
                },
                {
                  key: 'activeStudents',
                  label: 'Active',
                  sortable: true,
                },
              ]}
              data={courses}
              isLoading={isLoading}
            />
          </div>
        </div>

        {/* Key Insights */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-white mb-4">Key Insights</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <InsightCard
              title="High Engagement"
              description={`${metrics?.avgEngagementRate || 0}% of users are actively engaged this week`}
              variant="success"
              icon={<TrendingUp className="w-5 h-5" />}
            />

            <InsightCard
              title="Course Completion"
              description={`Average completion rate is ${metrics?.avgCourseCompletion || 0}% across all courses`}
              variant="info"
              icon={<BookOpen className="w-5 h-5" />}
            />

            <InsightCard
              title="Growth Trending"
              description={`${metrics?.newSignupsToday || 0} new signups and ${metrics?.newEnrollmentsToday || 0} new enrollments today`}
              variant="success"
              icon={<Users className="w-5 h-5" />}
            />
          </div>
        </div>

        {/* Quick Stats */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-white mb-4">Quick Stats</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="bg-gray-800 rounded-lg p-4">
              <p className="text-gray-400 text-xs uppercase mb-1">Avg Engagement</p>
              <p className="text-3xl font-bold text-white">{metrics?.avgEngagementRate || 0}%</p>
            </div>

            <div className="bg-gray-800 rounded-lg p-4">
              <p className="text-gray-400 text-xs uppercase mb-1">Completion Rate</p>
              <p className="text-3xl font-bold text-white">{metrics?.avgCourseCompletion || 0}%</p>
            </div>

            <div className="bg-gray-800 rounded-lg p-4">
              <p className="text-gray-400 text-xs uppercase mb-1">New Signups Today</p>
              <p className="text-3xl font-bold text-white">{metrics?.newSignupsToday || 0}</p>
            </div>

            <div className="bg-gray-800 rounded-lg p-4">
              <p className="text-gray-400 text-xs uppercase mb-1">New Enrollments</p>
              <p className="text-3xl font-bold text-white">{metrics?.newEnrollmentsToday || 0}</p>
            </div>
          </div>
        </div>

        {/* 5G Network Status & Metrics Section */}
        <div className="mb-8 border-t border-gray-700 pt-12">
          <h2 className="text-2xl font-bold text-white mb-6">5G Network Status</h2>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
            <NetworkStatus refreshInterval={5000} />
            <MetricsDisplay refreshInterval={5000} />
          </div>
        </div>

        {/* Quality & Edge Nodes Section */}
        <div className="mb-8">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <QualitySelector onProfileChanged={(profile) => console.log('Quality profile changed:', profile)} />
            <EdgeNodeViewer refreshInterval={10000} />
          </div>
        </div>
      </div>
    </div>
  );
}
