/** @type {import('next').NextConfig} */

// Determine deployment mode from environment variable
const deploymentMode = process.env.NEXT_PUBLIC_DEPLOYMENT_MODE || 'full'
const isLiteMode = deploymentMode === 'lite'
const basePath = process.env.NEXT_PUBLIC_BASE_PATH || ''

console.log(`[Next.js Config] Building for deployment mode: ${deploymentMode}`)
console.log(`[Next.js Config] Base path: ${basePath || '(none)'}`)
console.log(`[Next.js Config] Output: ${isLiteMode ? 'export (static)' : 'standalone (SSR)'}`)

const nextConfig = {
  reactStrictMode: true,

  // Base path for GitHub Pages (if using repo-name subdirectory)
  basePath: basePath,

  // Image configuration
  images: {
    domains: ['localhost', 'cgc-lb-and-cdn-content.nyc3.cdn.digitaloceanspaces.com'],
    unoptimized: true  // Required for static export
  },

  // Output configuration based on deployment mode
  ...(isLiteMode && {
    output: 'export',  // Static HTML export for GitHub Pages
    trailingSlash: true,  // Required for GitHub Pages compatibility
  }),

  // API rewrites (only for full mode with dev server)
  ...(!isLiteMode && {
    async rewrites() {
      return [
        {
          source: '/api/:path*',
          destination: 'http://localhost:8080/api/:path*',
        },
      ]
    },
  }),
}

module.exports = nextConfig