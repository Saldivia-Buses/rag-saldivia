const coreWebVitals = require("eslint-config-next/core-web-vitals")
const reactCompiler = require("eslint-plugin-react-compiler")

module.exports = [
  ...coreWebVitals,
  {
    plugins: {
      "react-compiler": reactCompiler,
    },
    rules: {
      "no-console": ["warn", { allow: ["warn", "error"] }],
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
      "react-hooks/exhaustive-deps": "warn",
      // Next 16 / React Compiler rules — patrones válidos (effects de montaje, tablas, RHF)
      "react-hooks/purity": "off",
      "react-hooks/set-state-in-effect": "off",
      "react-hooks/incompatible-library": "off",
      // React Compiler healthcheck
      "react-compiler/react-compiler": "warn",
    },
  },
]
