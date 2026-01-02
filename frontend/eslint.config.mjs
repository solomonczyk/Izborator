import { defineConfig, globalIgnores } from "eslint/config";
import nextTs from "eslint-config-next/typescript.js";
import next from "@next/eslint-plugin-next";

// В Next.js 15 конфиги могут быть объектами, а не массивами
const nextTsConfig = Array.isArray(nextTs) ? nextTs : [nextTs];

const eslintConfig = defineConfig([
  ...nextTsConfig,
  {
    plugins: { "@next/next": next },
    rules: {
      ...next.configs["core-web-vitals"].rules,
    },
  },
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;
