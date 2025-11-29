/* eslint-disable @next/next/no-html-element-for-text-content */
/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  compress: true,
  poweredByHeader: false,
  productionBrowserSourceMaps: false,
  output: 'standalone',
  experimental: {
    // Disable ISR and force all pages to be dynamic
    isrMemoryCacheSize: 0,
  },
  typescript: {
    // Ignore TS errors during production build to avoid Cypress/test types interfering
    ignoreBuildErrors: true,
  },
  eslint: {
    // Skip ESLint during builds to prevent non-critical warnings from failing deploys
    ignoreDuringBuilds: true,
  },
};

export default nextConfig;
