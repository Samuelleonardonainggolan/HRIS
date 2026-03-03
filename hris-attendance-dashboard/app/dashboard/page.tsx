// app/dashboard/page.tsx
'use client';

import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function DashboardPage() {
  const { user } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (user) {
      // Redirect based on role
      switch (user.role) {
        case 'manager_hr':
          router.push('/dashboard/manager-hr');
          break;
        case 'manager_departemen':
          router.push('/dashboard/manager-dept');
          break;
        case 'admin_departemen':
          router.push('/dashboard/admin-dept');
          break;
        case 'staf':
          router.push('/dashboard/staf');
          break;
        default:
          // Stay on this page
          break;
      }
    }
  }, [user, router]);

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
      <p className="mt-2 text-gray-600">Redirecting to your dashboard...</p>
    </div>
  );
}