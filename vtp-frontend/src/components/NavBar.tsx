"use client";
import Link from 'next/link';
import { useEffect, useState } from 'react';
import { useAuth } from '@/store';
import { useTranslations } from 'next-intl';

interface HealthStatus {
  status: string;
  healthy?: boolean;
}

export function NavBar() {
  const { user } = useAuth();
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        setLoading(true);
        setError(null);
        const base = process.env.NEXT_PUBLIC_API_URL;
        if (!base) {
          setError('API URL not set');
          setLoading(false);
          return;
        }
        const res = await fetch(base + '/health');
        if (!res.ok) throw new Error('Health request failed');
        const data = await res.json();
        setHealth(data);
      } catch (e: any) {
        setError(e.message);
      } finally {
        setLoading(false);
      }
    };
    fetchHealth();
    const interval = setInterval(fetchHealth, 30000); // refresh every 30s
    return () => clearInterval(interval);
  }, []);

  const t = useTranslations();
  return (
    <header className="sticky top-0 z-50 backdrop-blur bg-white/70 border-b border-gray-200">
      <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link href="/" className="text-xl font-semibold text-indigo-600">{t('brand.vtp')}</Link>
          <nav className="hidden md:flex items-center gap-4 text-sm font-medium text-gray-700">
            <Link href="/courses" className="hover:text-indigo-600">{t('nav.courses')}</Link>
                        <Link href="/instructors" className="hover:text-indigo-600">{t('nav.instructors')}</Link>
            <Link href="/materials" className="hover:text-indigo-600">{t('nav.materials')}</Link>
            <Link href="/assignments" className="hover:text-indigo-600">{t('assignments.title')}</Link>
            <Link href="/my-courses" className="hover:text-indigo-600">{t('nav.myCourses')}</Link>
            <Link href="/dashboard" className="hover:text-indigo-600">{t('nav.dashboard')}</Link>
            <Link href="/stream/demo" className="hover:text-indigo-600">{t('nav.stream')}</Link>
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <div className="text-xs">
            {loading && <span className="text-gray-500">{t('status.health.loading')}</span>}
            {!loading && error && <span className="text-red-600">{t('status.health.down')}</span>}
            {!loading && !error && health && (
              <span className={health.healthy ? 'text-green-600' : 'text-yellow-600'}>
                {t('status.api')}: {health.status || (health.healthy ? 'OK' : 'WARN')}
              </span>
            )}
          </div>
          {user ? (
            <div className="flex items-center gap-2 text-sm">
              <span className="text-gray-700">{user.firstName ? `${user.firstName} ${user.lastName}` : user.email}</span>
              <Link href="/logout" className="text-indigo-600 hover:underline">{t('nav.logout')}</Link>
            </div>
          ) : (
            <div className="flex items-center gap-2 text-sm">
              <Link href="/login" className="text-indigo-600 hover:underline">{t('nav.login')}</Link>
              <Link href="/register" className="text-gray-600 hover:text-indigo-600">{t('nav.register')}</Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
