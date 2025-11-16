'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';
import { useEffect } from 'react';

const navigation = [
  { name: 'Dashboard', href: '/dashboard' },
  { name: 'Analytics', href: '/analytics' },
];

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout, checkAuth, setMockUser } = useAuth();

  useEffect(() => {
    // Development mode: Auto-login with mock user
    if (process.env.NODE_ENV === 'development' && typeof window !== 'undefined') {
      const token = localStorage.getItem('token');
      if (!token) {
        // Create mock user for development
        const mockUser = {
          id: 'dev-user-123',
          email: 'admin@taskflow.dev',
          name: 'Admin User',
        };
        localStorage.setItem('token', 'dev-mock-token');
        setMockUser(mockUser);
        return;
      }
    }
    checkAuth();
  }, [checkAuth, setMockUser]);

  useEffect(() => {
    if (!user && typeof window !== 'undefined') {
      const token = localStorage.getItem('token');
      if (!token && process.env.NODE_ENV !== 'development') {
        router.push('/login');
      }
    }
  }, [user, router]);

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  if (!user) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-gray-50 overflow-hidden">
      {/* Sidebar */}
      <div className="w-64 bg-white border-r border-gray-200 flex flex-col flex-shrink-0">
        <div className="p-6 border-b border-gray-200 flex-shrink-0">
          <h1 className="text-2xl font-bold text-primary">TaskFlow</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Intelligent Prioritization
          </p>
        </div>

        <nav className="flex-1 p-4 space-y-2 overflow-y-auto">
          {navigation.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.name}
                href={item.href}
                className={`
                  block px-4 py-2 rounded-lg transition-colors
                  ${
                    isActive
                      ? 'bg-primary text-primary-foreground'
                      : 'text-gray-700 hover:bg-gray-100'
                  }
                `}
              >
                {item.name}
              </Link>
            );
          })}
        </nav>

        <div className="p-4 border-t border-gray-200 flex-shrink-0">
          <div className="mb-3">
            <p className="text-sm font-medium">{user.name}</p>
            <p className="text-xs text-muted-foreground">{user.email}</p>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={handleLogout}
            className="w-full"
          >
            Sign Out
          </Button>
        </div>
      </div>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto">
        <div className="p-8">
          {children}
        </div>
      </main>
    </div>
  );
}
