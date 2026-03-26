/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: '#0D0D0F',
        surface: '#17171A',
        border: '#2A2A2F',
        'text-primary': '#EDEDEF',
        'text-muted': '#71717A',
        accent: '#6366F1',
        success: '#22C55E',
        warning: '#F59E0B',
        danger: '#EF4444',
        openai: '#10A37F',
        anthropic: '#D97706',
        google: '#4285F4',
        ollama: '#7C3AED',
        cohere: '#6B7280',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Consolas', 'monospace'],
      },
      fontSize: {
        micro: '11px',
        cell: '12px',
        body: '13px',
        section: '15px',
        title: '20px',
        hero: '32px',
      },
    },
  },
  plugins: [],
}
