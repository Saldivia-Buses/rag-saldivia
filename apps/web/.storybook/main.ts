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
    return config
  },
}

export default config
