import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: '#4F46E5',
          dark: '#3730A3',
        },
      },
    },
  },
  plugins: [],
}

export default config

import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#f5f7ff',
          100: '#ebefff',
          200: '#cfd9ff',
          300: '#a6baff',
          400: '#7893ff',
          500: '#4f6cff',
          600: '#2d4bff',
          700: '#1f37db',
          800: '#1b2fab',
          900: '#1a2a85',
        },
      },
    },
  },
  plugins: [],
}

export default config


