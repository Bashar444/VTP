"use client";
export const dynamic = 'force-dynamic';
import Link from 'next/link';
import { useFeaturedCourses } from '@/hooks/useFeaturedCourses';
import { useTranslations } from 'next-intl';

export default function Home() {
  const { data: courses = [], isLoading: loading, error } = useFeaturedCourses(4);
  const t = useTranslations();
  return (
    <main className="max-w-7xl mx-auto px-4 py-10">
      <section className="mb-10 text-center">
        <h1 className="text-4xl font-bold text-gray-900 mb-3">{t('home.title')}</h1>
        <p className="text-lg text-gray-600 mb-6">{t('home.subtitle')}</p>
        <div className="flex flex-wrap gap-3 justify-center">
          <Link href="/courses" className="px-5 py-2 rounded-md bg-indigo-600 text-white hover:bg-indigo-500 text-sm font-medium">{t('home.browseCourses')}</Link>
          <Link href="/dashboard" className="px-5 py-2 rounded-md bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 text-sm font-medium">{t('home.dashboard')}</Link>
          <Link href="/stream/demo" className="px-5 py-2 rounded-md bg-green-600 text-white hover:bg-green-500 text-sm font-medium">{t('home.joinStream')}</Link>
          <Link href="/login" className="px-5 py-2 rounded-md bg-gray-800 text-white hover:bg-gray-700 text-sm font-medium">{t('home.login')}</Link>
        </div>
      </section>

      <section>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-2xl font-semibold text-gray-900">{t('home.featured')}</h2>
          <Link href="/courses" className="text-sm text-indigo-600 hover:underline">{t('home.viewAll')}</Link>
        </div>
        {loading && (
          <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 md:grid-cols-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="h-40 bg-gray-100 animate-pulse rounded" />
            ))}
          </div>
        )}
        {!loading && error && (
          <div className="p-4 bg-red-50 border border-red-200 text-red-700 rounded text-sm">{t('errors.network')}</div>
        )}
        {!loading && !error && courses.length === 0 && (
          <p className="text-sm text-gray-600">{t('home.noFeatured')}</p>
        )}
        {!loading && !error && courses.length > 0 && (
          <div className="grid gap-6 grid-cols-1 sm:grid-cols-2 md:grid-cols-4">
            {courses.map((c) => (
              <Link
                href={`/courses/${c.id}`}
                key={c.id}
                className="group border border-gray-200 rounded-lg p-4 bg-white hover:shadow-sm transition"
              >
                <div className="h-28 mb-3 rounded bg-gradient-to-br from-indigo-100 to-indigo-200 flex items-center justify-center text-indigo-600 text-sm font-medium">
                  {c.thumbnail ? (
                    <img src={c.thumbnail} alt={c.title} className="h-full w-full object-cover rounded" />
                  ) : (
                    <span>{c.title.slice(0, 20)}</span>
                  )}
                </div>
                <h3 className="text-sm font-semibold text-gray-900 mb-1 group-hover:text-indigo-600 line-clamp-2">{c.title}</h3>
                <p className="text-xs text-gray-600 line-clamp-3 mb-2">{c.description}</p>
                <div className="flex items-center justify-between text-xs text-gray-500">
                  <span>{c.rating ? `★ ${c.rating.toFixed(1)}` : t('course.new')}</span>
                  <span>{c.studentCount ? `${c.studentCount} ${t('course.learners')}` : ''}</span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}
