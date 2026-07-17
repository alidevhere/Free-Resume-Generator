/** @type {import('next').NextConfig} */
const backendApiUrl = process.env.BACKEND_API_URL || "http://localhost:8080";
const isTauriExport = process.env.TAURI_EXPORT === "true";

const nextConfig = {
  reactStrictMode: true,
  ...(isTauriExport ? { output: "export" } : {}),
  ...(!isTauriExport
    ? {
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
      }
    : {})
};

export default nextConfig;
