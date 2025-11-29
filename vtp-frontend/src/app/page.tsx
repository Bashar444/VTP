'use client';

export default function Home() {
  return (
    <main className="flex items-center justify-center min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          VTP Frontend
        </h1>
        <p className="text-xl text-gray-600 mb-8">
          Video Teaching Platform - Phase 5A Setup Complete
        </p>
        <div className="space-y-4">
          <p className="text-lg text-gray-700">
            ✅ Next.js 14 project initialized
          </p>
          <p className="text-lg text-gray-700">
            ✅ TypeScript configured
          </p>
          <p className="text-lg text-gray-700">
            ✅ Tailwind CSS ready
          </p>
          <p className="text-lg text-gray-700">
            ✅ Ready for component development
          </p>
        </div>
      </div>
    </main>
  );
}
