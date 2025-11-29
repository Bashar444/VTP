import type { Metadata, Viewport } from 'next';
import './globals.css';
import { Providers } from './providers';
import React from 'react';

// Force dynamic rendering for all routes
export const dynamic = 'force-dynamic';
export const dynamicParams = true;
export const revalidate = 0;

export const metadata: Metadata = {
  title: 'VTP - Video Teaching Platform',
  description: 'Advanced educational video platform for Syrian students',
  icons: { icon: '/favicon.ico' },
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ar" dir="rtl">
      <body className="min-h-screen flex flex-col font-sans bg-white text-gray-900">
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  );
}
