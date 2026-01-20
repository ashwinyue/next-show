import { Outlet, Navigate } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { useAuthStore } from '@/stores/useAuthStore';
import { Toaster } from '@/components/ui/toaster';

export function Layout() {
  const { token } = useAuthStore();

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="flex h-screen bg-background text-foreground overflow-hidden">
      <Sidebar />
      <main className="flex-1 overflow-auto relative flex flex-col">
        <Outlet />
      </main>
      <Toaster />
    </div>
  );
}
