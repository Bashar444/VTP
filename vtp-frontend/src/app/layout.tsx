import type { Metadata, Viewport } from 'next';
import './globals.css';
import { NavBar } from '@/components/NavBar';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

export const metadata: Metadata = {
  title: 'VTP - Video Teaching Platform',
  description: 'Advanced educational video platform for Syrian students',
  icons: { icon: '/favicon.ico' },
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
};

const queryClient = new QueryClient();

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen flex flex-col">
        <QueryClientProvider client={queryClient}>
          <NavBar />
          <div className="flex-1">{children}</div>
        </QueryClientProvider>
      </body>
    </html>
  );
}
