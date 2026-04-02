import {
  Inter,
  DM_Sans,
  Plus_Jakarta_Sans,
  Poppins,
  Montserrat,
  Open_Sans,
  Roboto,
  Outfit,
  Quicksand,
  Oxanium,
  Antic,
  Architects_Daughter,
  Libre_Baskerville,
  Merriweather,
  Playfair_Display,
  Geist,
  JetBrains_Mono,
  Fira_Code,
  Roboto_Mono,
  Source_Code_Pro,
  IBM_Plex_Mono,
  Space_Mono,
  Ubuntu_Mono,
  Lora,
  Source_Serif_4,
} from "next/font/google";

// Sans fonts
const inter = Inter({ subsets: ["latin"], variable: "--font-inter", display: "swap" });
const dmSans = DM_Sans({ subsets: ["latin"], variable: "--font-dm-sans", display: "swap" });
const plusJakarta = Plus_Jakarta_Sans({ subsets: ["latin"], variable: "--font-plus-jakarta", display: "swap" });
const poppins = Poppins({ subsets: ["latin"], variable: "--font-poppins", weight: ["300", "400", "500", "600", "700"], display: "swap" });
const montserrat = Montserrat({ subsets: ["latin"], variable: "--font-montserrat", display: "swap" });
const openSans = Open_Sans({ subsets: ["latin"], variable: "--font-open-sans", display: "swap" });
const roboto = Roboto({ subsets: ["latin"], variable: "--font-roboto", display: "swap" });
const outfit = Outfit({ subsets: ["latin"], variable: "--font-outfit", display: "swap" });
const quicksand = Quicksand({ subsets: ["latin"], variable: "--font-quicksand", display: "swap" });
const oxanium = Oxanium({ subsets: ["latin"], variable: "--font-oxanium", display: "swap" });
const antic = Antic({ subsets: ["latin"], variable: "--font-antic", weight: "400", display: "swap" });
const architectsDaughter = Architects_Daughter({ subsets: ["latin"], variable: "--font-architects-daughter", weight: "400", display: "swap" });
const libreBaskerville = Libre_Baskerville({ subsets: ["latin"], variable: "--font-libre-baskerville", weight: ["400", "700"], display: "swap" });
const merriweather = Merriweather({ subsets: ["latin"], variable: "--font-merriweather", weight: ["300", "400", "700"], display: "swap" });
const playfairDisplay = Playfair_Display({ subsets: ["latin"], variable: "--font-playfair-display", display: "swap" });
const geist = Geist({ subsets: ["latin"], variable: "--font-geist", display: "swap" });

// Mono fonts
const jetbrainsMono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-jetbrains-mono", display: "swap" });
const firaCode = Fira_Code({ subsets: ["latin"], variable: "--font-fira-code", display: "swap" });
const robotoMono = Roboto_Mono({ subsets: ["latin"], variable: "--font-roboto-mono", display: "swap" });
const sourceCodePro = Source_Code_Pro({ subsets: ["latin"], variable: "--font-source-code-pro", display: "swap" });
const ibmPlexMono = IBM_Plex_Mono({ subsets: ["latin"], variable: "--font-ibm-plex-mono", weight: ["400", "500", "600"], display: "swap" });
const spaceMono = Space_Mono({ subsets: ["latin"], variable: "--font-space-mono", weight: ["400", "700"], display: "swap" });
const ubuntuMono = Ubuntu_Mono({ subsets: ["latin"], variable: "--font-ubuntu-mono", weight: ["400", "700"], display: "swap" });

// Serif fonts
const lora = Lora({ subsets: ["latin"], variable: "--font-lora", display: "swap" });
const sourceSerif4 = Source_Serif_4({ subsets: ["latin"], variable: "--font-source-serif-4", display: "swap" });

// All font CSS variable classes combined
export const fontVariables = [
  inter, dmSans, plusJakarta, poppins, montserrat, openSans, roboto,
  outfit, quicksand, oxanium, antic, architectsDaughter, libreBaskerville,
  merriweather, playfairDisplay, geist,
  jetbrainsMono, firaCode, robotoMono, sourceCodePro, ibmPlexMono, spaceMono, ubuntuMono,
  lora, sourceSerif4,
].map((f) => f.variable).join(" ");

// Map font names (as they appear in theme presets) to CSS variable references
export const fontFamilyMap: Record<string, string> = {
  "Inter": "var(--font-inter)",
  "DM Sans": "var(--font-dm-sans)",
  "Plus Jakarta Sans": "var(--font-plus-jakarta)",
  "Poppins": "var(--font-poppins)",
  "Montserrat": "var(--font-montserrat)",
  "Open Sans": "var(--font-open-sans)",
  "Roboto": "var(--font-roboto)",
  "Outfit": "var(--font-outfit)",
  "Quicksand": "var(--font-quicksand)",
  "Oxanium": "var(--font-oxanium)",
  "Antic": "var(--font-antic)",
  "Architects Daughter": "var(--font-architects-daughter)",
  "Libre Baskerville": "var(--font-libre-baskerville)",
  "Merriweather": "var(--font-merriweather)",
  "Playfair Display": "var(--font-playfair-display)",
  "Geist": "var(--font-geist)",
  "JetBrains Mono": "var(--font-jetbrains-mono)",
  "Fira Code": "var(--font-fira-code)",
  "Roboto Mono": "var(--font-roboto-mono)",
  "Source Code Pro": "var(--font-source-code-pro)",
  "IBM Plex Mono": "var(--font-ibm-plex-mono)",
  "Space Mono": "var(--font-space-mono)",
  "Ubuntu Mono": "var(--font-ubuntu-mono)",
  "Lora": "var(--font-lora)",
  "Source Serif 4": "var(--font-source-serif-4)",
  "Geist Mono": "var(--font-jetbrains-mono)", // fallback
};
