/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        neocode: {
          bg: '#0d1117',
          surface: '#161b22',
          border: '#30363d',
          text: '#c9d1d9',
          muted: '#8b949e',
          primary: '#58a6ff',
          success: '#3fb950',
          warning: '#d29922',
          error: '#f85149',
          accent: '#bc8cff',
        },
      },
    },
  },
  plugins: [],
}
