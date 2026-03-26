import type { Preview, Decorator } from "@storybook/react"
import { withThemeByClassName } from "@storybook/addon-themes"
import "../src/app/globals.css"

// Wrapper que aplica el fondo del tema al canvas de cada story
const withBackground: Decorator = (Story, context) => (
  <div
    className={context.globals.theme === "dark" ? "dark" : ""}
    style={{ background: "var(--bg)", minHeight: "100%", padding: "1.5rem" }}
  >
    <Story />
  </div>
)

const preview: Preview = {
  parameters: {
    backgrounds: { disable: true },
    layout: "centered",
    controls: { matchers: { color: /(background|color)$/i, date: /Date$/i } },
  },
  decorators: [
    withBackground,
    withThemeByClassName({
      themes: {
        light: "",
        dark: "dark",
      },
      defaultTheme: "light",
    }),
  ],
}

export default preview
