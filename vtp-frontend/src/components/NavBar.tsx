"use client";
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';

export function NavBar() {
  const { user, logout } = useAuth();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <header className="bg-white shadow-sm">
      <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
        <div className="flex items-center gap-8">
          <Link href="/" className="text-2xl font-bold text-indigo-600">
            VTP
          </Link>
          {user && (
            <nav className="flex items-center gap-6">
              {/* Student/Instructor Nav */}
              {(user.role === 'student' || user.role === 'instructor') && (
                <>
                  <Link href="/courses" className="text-gray-700 hover:text-indigo-600">
                    Courses
                  </Link>
                  <Link href="/my-courses" className="text-gray-700 hover:text-indigo-600">
                    My Courses
                  </Link>
                </>
              )}
              
              {/* Instructor Only */}
              {user.role === 'instructor' && (
                <Link href="/instructor/courses" className="text-gray-700 hover:text-indigo-600">
                  Manage Courses
                </Link>
              )}
              
              {/* Admin Only */}
              {user.role === 'admin' && (
                <>
                  <Link href="/admin/users" className="text-gray-700 hover:text-indigo-600">
                    Users
                  </Link>
                  <Link href="/admin/courses" className="text-gray-700 hover:text-indigo-600">
                    All Courses
                  </Link>
                  <Link href="/admin/dashboard" className="text-gray-700 hover:text-indigo-600">
                    Statistics
                  </Link>
                </>
              )}
            </nav>
          )}
        </div>

        <div className="flex items-center gap-4">
          {user ? (
            <>
              <Link href="/profile" className="text-gray-700 hover:text-indigo-600">
                {user.full_name}
              </Link>
              <button
                onClick={handleLogout}
                className="text-gray-600 hover:text-red-600"
              >
                Logout
              </button>
            </>
          ) : (
            <>
              <Link href="/login" className="text-indigo-600 hover:underline">
                Login
              </Link>
              <Link
                href="/register"
                className="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700"
              >
                Sign Up
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
