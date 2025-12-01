"use client";

import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import Link from 'next/link';

export default function DashboardPage() {
  const { user } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!user) {
      router.push('/login');
      return;
    }

    // Redirect based on role
    if (user.role === 'student') {
      router.push('/my-courses');
    } else if (user.role === 'instructor') {
      router.push('/instructor/courses');
    } else if (user.role === 'admin') {
      router.push('/admin/dashboard');
    }
  }, [user, router]);

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">Redirecting...</p>
      </div>
    </div>
  );
}
          <p className="text-gray-400">{t('dashboard.subtitle')}</p>
        </div>

        {/* Filters */}
        <div className="mb-8">
          <AnalyticsFilters />
        </div>

        {/* Key Metrics Grid */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-white mb-4">{t('dashboard.keyMetrics')}</h2>
          <StatGrid
            stats={[
              {
                title: t('dashboard.stat.totalStudents'),
                value: metrics?.totalStudents || 0,
                icon: <Users className="w-6 h-6" />,
                trend: {
                  value: 12,
                  isPositive: true,
                  period: t('dashboard.trend.vsLastMonth'),
                },
              },
              {
                title: t('dashboard.stat.activeCourses'),
                value: metrics?.totalCourses || 0,
                icon: <BookOpen className="w-6 h-6" />,
                trend: {
                  value: 5,
                  isPositive: true,
                  period: t('dashboard.trend.addedThisMonth'),
                },
              },
              {
                title: t('dashboard.stat.activeUsers'),
                value: metrics?.activeUsers || 0,
                icon: <Activity className="w-6 h-6" />,
                trend: {
                  value: 8,
                  isPositive: true,
                  period: t('dashboard.trend.growthThisWeek'),
                },
              },
              {
                title: t('dashboard.stat.totalRevenue'),
                value: `$${(metrics?.totalRevenue || 0).toLocaleString()}`,
                icon: <DollarSign className="w-6 h-6" />,
                trend: {
                  value: 23,
                  isPositive: true,
                  period: t('dashboard.trend.vsLastMonth'),
                },
              },
              {
                title: t('dashboard.stat.networkQuality'),
                value: networkQualityLoading ? '…' : (networkQuality !== undefined ? `${networkQuality}%` : 'N/A'),
                icon: <Activity className="w-6 h-6" />,
                trend: {
                  value: 0,
                  isPositive: true,
                  period: t('dashboard.trend.realTime'),
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
            <h2 className="text-2xl font-bold text-white mb-4">{t('dashboard.systemAlerts')}</h2>
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
            <h2 className="text-2xl font-bold text-white mb-4">{t('dashboard.engagementTrend')}</h2>
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
            <h2 className="text-2xl font-bold text-white mb-4">{t('dashboard.watchTimeByCourse')}</h2>
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
            <h2 className="text-2xl font-bold text-white mb-4">{t('dashboard.completionRates')}</h2>
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
            <h2 className="text-2xl font-bold text-white mb-4">{t('dashboard.topCourses')}</h2>
            <DataTable
              columns={[
                {
                  key: 'courseName',
                  label: t('dashboard.table.course'),
                  sortable: true,
                },
                {
                  key: 'enrollmentCount',
                  label: t('dashboard.table.enrollments'),
                  sortable: true,
                },
                {
                  key: 'completionRate',
                  label: t('dashboard.table.completion'),
                  render: (value: number) => `${value.toFixed(1)}%`,
                  sortable: true,
                },
                {
                  key: 'avgRating',
                  label: t('dashboard.table.rating'),
                  render: (value: number) => `${value.toFixed(1)}/5`,
                  sortable: true,
                },
                {
                  key: 'activeStudents',
                  label: t('dashboard.table.active'),
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
          <h2 className="text-2xl font-bold text-white mb-4">{t('dashboard.keyInsights')}</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <InsightCard
              title={t('dashboard.insight.highEngagement.title')}
              description={t('dashboard.insight.highEngagement.desc', { rate: metrics?.avgEngagementRate || 0 })}
              variant="success"
              icon={<TrendingUp className="w-5 h-5" />}
            />

            <InsightCard
              title={t('dashboard.insight.courseCompletion.title')}
              description={t('dashboard.insight.courseCompletion.desc', { rate: metrics?.avgCourseCompletion || 0 })}
              variant="info"
              icon={<BookOpen className="w-5 h-5" />}
            />

            <InsightCard
              title={t('dashboard.insight.growth.title')}
              description={t('dashboard.insight.growth.desc', { signups: metrics?.newSignupsToday || 0, enrollments: metrics?.newEnrollmentsToday || 0 })}
              variant="success"
              icon={<Users className="w-5 h-5" />}
            />
          </div>
        </div>

        {/* Quick Stats */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-white mb-4">{t('dashboard.quickStats')}</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="bg-gray-800 rounded-lg p-4">
              <p className="text-gray-400 text-xs uppercase mb-1">{t('dashboard.stat.avgEngagement')}</p>
              <p className="text-3xl font-bold text-white">{metrics?.avgEngagementRate || 0}%</p>
            </div>

            <div className="bg-gray-800 rounded-lg p-4">
              <p className="text-gray-400 text-xs uppercase mb-1">{t('dashboard.stat.completionRate')}</p>
              <p className="text-3xl font-bold text-white">{metrics?.avgCourseCompletion || 0}%</p>
            </div>

            <div className="bg-gray-800 rounded-lg p-4">
              <p className="text-gray-400 text-xs uppercase mb-1">{t('dashboard.stat.newSignupsToday')}</p>
              <p className="text-3xl font-bold text-white">{metrics?.newSignupsToday || 0}</p>
            </div>

            <div className="bg-gray-800 rounded-lg p-4">
              <p className="text-gray-400 text-xs uppercase mb-1">{t('dashboard.stat.newEnrollments')}</p>
              <p className="text-3xl font-bold text-white">{metrics?.newEnrollmentsToday || 0}</p>
            </div>
          </div>
        </div>

        {/* 5G Network Status & Metrics Section */}
        <div className="mb-8 border-t border-gray-700 pt-12">
          <h2 className="text-2xl font-bold text-white mb-6">{t('dashboard.networkStatus5g')}</h2>
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
