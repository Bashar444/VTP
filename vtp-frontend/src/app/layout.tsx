import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'VTP - Video Teaching Platform',
  description: 'Advanced educational video platform for Syrian students',
  viewport: 'width=device-width, initial-scale=1',
  icons: {
    icon: '/favicon.ico',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
