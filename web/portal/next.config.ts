import type { NextConfig } from "next";

const apiTarget = process.env.SUPAY_API_PROXY_TARGET
  ?? (process.env.NODE_ENV === "development" ? "http://127.0.0.1:8084" : "http://gateway:8080");

const nextConfig: NextConfig = {
  output: "standalone",
  poweredByHeader: false,
  // Let devices on the local network load Next's dev client and HMR resources.
  allowedDevOrigins: ["192.168.1.7"],
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${apiTarget}/api/:path*` }];
  },
};

export default nextConfig;
