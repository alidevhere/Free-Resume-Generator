/** @type {import('next').NextConfig} */
const backendApiUrl = process.env.BACKEND_API_URL || "http://localhost:8080";

const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${backendApiUrl}/api/:path*`
      },
      {
        source: "/health",
        destination: `${backendApiUrl}/health`
      }
    ];
  }
};

export default nextConfig;
