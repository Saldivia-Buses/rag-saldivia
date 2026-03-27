module.exports = {
  "apps/web/src/**/*.{ts,tsx}": [
    "bunx eslint --fix --max-warnings 0 --config apps/web/eslint.config.js",
  ],
}
