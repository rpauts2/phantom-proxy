/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  experimental: {
    serverActions: {
      bodySizeLimit: '2mb',
    },
  },
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: 'http://localhost:8080/api/v1/:path*',
      },
      {
        source: '/health',
        destination: 'http://localhost:8080/health',
      },
      {
        source: '/metrics',
        destination: 'http://localhost:8080/metrics',
      },
      {
        source: '/login',
        destination: 'http://localhost:8080/login',
      },
    ]
  },
}

module.exports = nextConfig
------- REPLACE

