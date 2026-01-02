import { defineConfig, globalIgnores } from "eslint/config";
import next from "@next/eslint-plugin-next";

const eslintConfig = defineConfig([
  {
    plugins: { "@next/next": next },
    rules: {
      ...next.configs["core-web-vitals"].rules,
    },
  },
  globalIgnores([
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;
