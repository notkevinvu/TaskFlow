import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  outputFileTracingIncludes: {
    '/': ['./public/**/*'],
  },
  // Enable tree-shaking for large icon libraries
  // This ensures only imported icons are bundled, not the entire library
  experimental: {
    optimizePackageImports: ['lucide-react'],
  },
};

export default nextConfig;
