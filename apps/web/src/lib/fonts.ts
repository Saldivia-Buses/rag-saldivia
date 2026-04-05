import { Source_Sans_3, Source_Serif_4, Source_Code_Pro } from "next/font/google";

const sourceSans = Source_Sans_3({ subsets: ["latin"], variable: "--font-source-sans", display: "swap" });
const sourceSerif = Source_Serif_4({ subsets: ["latin"], variable: "--font-source-serif", display: "swap" });
const sourceCode = Source_Code_Pro({ subsets: ["latin"], variable: "--font-source-code", display: "swap" });

const allFonts = [sourceSans, sourceSerif, sourceCode];

export const fontVariables = allFonts.map((f) => f.variable).join(" ");
export const fontClassNames = allFonts.map((f) => f.className).join(" ");
