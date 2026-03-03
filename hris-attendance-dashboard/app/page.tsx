// app/page.tsx
import Link from 'next/link';
import { ArrowRight, CheckCircle } from 'lucide-react';

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Hero Section */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-16">
        <div className="text-center">
          <h1 className="text-5xl font-extrabold text-gray-900 sm:text-6xl">
            HRIS System
          </h1>
          <p className="mt-4 text-xl text-gray-600 max-w-2xl mx-auto">
            Face Recognition Attendance Management System
          </p>
          
          <div className="mt-10 flex justify-center space-x-4">
            <Link
              href="/login"
              className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-lg text-white bg-indigo-600 hover:bg-indigo-700 transition-colors"
            >
              Login
              <ArrowRight className="ml-2 h-5 w-5" />
            </Link>
            <Link
              href="/dashboard"
              className="inline-flex items-center px-6 py-3 border border-gray-300 text-base font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 transition-colors"
            >
              Dashboard
            </Link>
          </div>
        </div>

        {/* Features */}
        <div className="mt-20 grid grid-cols-1 md:grid-cols-3 gap-8">
          {[
            'Face Recognition Attendance',
            'Real-time Tracking',
            'Automated Reports',
            'Multi-role Access',
            'Department Management',
            'Secure Authentication',
          ].map((feature, index) => (
            <div
              key={index}
              className="bg-white rounded-lg p-6 shadow-md hover:shadow-lg transition-shadow"
            >
              <CheckCircle className="h-8 w-8 text-green-500 mb-3" />
              <h3 className="text-lg font-semibold text-gray-900">{feature}</h3>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}