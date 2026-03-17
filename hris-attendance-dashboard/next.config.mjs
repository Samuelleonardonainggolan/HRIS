/** @type {import('next').NextConfig} */
const nextConfig = {
  // Allow dev requests from additional origins (silences cross-origin dev warnings)
  allowedDevOrigins: [
    '192.168.56.1',
    '127.0.0.1',
    'localhost',
  ],
  async rewrites() {
    const backendBaseUrl = process.env.BACKEND_BASE_URL || 'http://localhost:8080';
    return [
      {
        source: '/api/v1/auth/:path*',
        destination: `${backendBaseUrl}/api/v1/auth/:path*`,
      },
      {
        source: '/api/v1/logout',
        destination: `${backendBaseUrl}/api/v1/logout`,
      },
    ];
  },
};

export default nextConfig;
