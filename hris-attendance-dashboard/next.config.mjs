/** @type {import('next').NextConfig} */
const nextConfig = {
  // Allow dev requests from additional origins (silences cross-origin dev warnings)
  allowedDevOrigins: [
    '192.168.56.1',
    '127.0.0.1',
    'localhost',
  ],
};

export default nextConfig;
