import {heroui} from "@heroui/theme"

/** @type {import('tailwindcss').Config} */
const config = {
  content: [
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
    "./node_modules/@heroui/theme/dist/**/*.{js,ts,jsx,tsx}"
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["var(--font-sans)"],
        mono: ["var(--font-mono)"],
      },
    },
  },
  darkMode: "class",
  plugins: [
    heroui({
      themes: {
        dark: {
          colors: {
            background: "#0f172a", // slate-900
            foreground: "#e2e8f0", // slate-200
            content1: "#0b1220", // deep surface
            content2: "#111827", // slate-900/800 blend
            content3: "#0b1220",
            default: {
              DEFAULT: "#111827",
              foreground: "#e5e7eb",
            },
            primary: {
              DEFAULT: "#ef4444", // red-500 accent
              foreground: "#ffffff",
            },
            focus: "#fca5a5", // red-300 focus ring
          },
        },
      },
    }),
  ],
}

module.exports = config;