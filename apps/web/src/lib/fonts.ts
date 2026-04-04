import { Plus_Jakarta_Sans, Lora, Roboto_Mono } from "next/font/google";

const plusJakarta = Plus_Jakarta_Sans({ subsets: ["latin"], variable: "--font-plus-jakarta", display: "swap" });
const lora = Lora({ subsets: ["latin"], variable: "--font-lora", display: "swap" });
const robotoMono = Roboto_Mono({ subsets: ["latin"], variable: "--font-roboto-mono", display: "swap" });

const allFonts = [plusJakarta, lora, robotoMono];

export const fontVariables = allFonts.map((f) => f.variable).join(" ");
export const fontClassNames = allFonts.map((f) => f.className).join(" ");
