import type { NextConfig } from "next";
import createNextIntlPlugin from 'next-intl/plugin';

const withNextIntl = createNextIntlPlugin('./i18n.ts');

const nextConfig: NextConfig = {
  output: "standalone", // Docker standalone output
  experimental: {
    allowedDevOrigins: ["http://127.0.0.1:3000"],
  },
};

export default withNextIntl(nextConfig);

