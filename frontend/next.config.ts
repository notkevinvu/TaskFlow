import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  outputFileTracingIncludes: {
    '/': ['./public/**/*'],
  },
};

export default nextConfig;
