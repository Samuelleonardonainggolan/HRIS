import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const __workspaceRoot = path.resolve(__dirname, "..");

/** @type {import('next').NextConfig} */
const nextConfig = {
  turbopack: {
    root: __workspaceRoot,
  },
  // Allow dev requests from additional origins (silences cross-origin dev warnings)
  allowedDevOrigins: [
    '192.168.56.1',
    '127.0.0.1',
    'localhost',
  ],
  async rewrites() {
    const backendBaseUrl = process.env.BACKEND_BASE_URL || 'http://localhost:8080';
    const backendOrigin = new URL(backendBaseUrl).origin;
    return [
      {
        source: '/api/v1/auth/:path*',
        destination: `${backendOrigin}/api/v1/auth/:path*`,
      },
      {
        source: '/api/v1/logout',
        destination: `${backendOrigin}/api/v1/logout`,
      },
      {
        source: '/api/v1/leave-requests/:path*',
        destination: `${backendOrigin}/api/v1/leave-requests/:path*`,
      },
      {
        source: '/api/v1/:path*',
        destination: `${backendOrigin}/api/v1/:path*`,
      },
    ];
  },
};

export default nextConfig ;
