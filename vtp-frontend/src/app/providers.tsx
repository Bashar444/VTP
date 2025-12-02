'use client';

import React, { useState, useEffect } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { NavBar } from '@/components/NavBar';
import { NextIntlClientProvider } from 'next-intl';
import { getMessages, defaultLocale } from '@/i18n/config';
import { useAuthStore } from '@/store/auth.store';

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 60 * 1000, // 1 minute
        refetchOnWindowFocus: false,
      },
    },
  }));

  const initFromStorage = useAuthStore((state) => state.initFromStorage);

  useEffect(() => {
    // Initialize auth state from localStorage on mount
    initFromStorage();
  }, [initFromStorage]);

  const messages = getMessages(defaultLocale);
  return (
    <QueryClientProvider client={queryClient}>
      <NextIntlClientProvider locale={defaultLocale} messages={messages} timeZone="Asia/Damascus">
        <NavBar />
        <div className="flex-1" dir="rtl">{children}</div>
      </NextIntlClientProvider>
    </QueryClientProvider>
  );
}
