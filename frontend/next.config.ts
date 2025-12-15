import type { NextConfig } from "next";
import createNextIntlPlugin from 'next-intl/plugin';

const withNextIntl = createNextIntlPlugin('./i18n.ts');

const nextConfig: NextConfig = {
  output: "standalone", // Для Docker-сборки (уменьшает размер образа)
};

export default withNextIntl(nextConfig);
