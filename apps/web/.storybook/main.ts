import type { StorybookConfig } from "@storybook/react-vite"
import { resolve } from "path"
import type { UserConfig } from "vite"

const config: StorybookConfig = {
  stories: ["../stories/**/*.stories.@(ts|tsx)"],
  addons: [
    "@storybook/addon-essentials",
    "@storybook/addon-a11y",
    "@storybook/addon-themes",
  ],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
  docs: {
    autodocs: "tag",
  },
  typescript: {
    check: false,
    reactDocgen: "react-docgen-typescript",
  },
  async viteFinal(config: UserConfig) {
    const { default: tailwindcss } = await import("@tailwindcss/vite")
    config.plugins = [...(config.plugins ?? []), tailwindcss()]
    config.resolve = {
      ...config.resolve,
      alias: {
        ...((config.resolve as any)?.alias ?? {}),
        "@": resolve(__dirname, "../src"),
      },
    }
    config.esbuild = {
      ...config.esbuild,
      jsx: "automatic",
    }
    return config
  },
}

export default config
