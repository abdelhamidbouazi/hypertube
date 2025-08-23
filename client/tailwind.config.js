/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{html,ts}'],
  theme: {
    extend: {
      colors: {
        // HyperTybe custom color palette
        'ht-darkest': '#595959',
        'ht-dark': '#7f7f7f',
        'ht-medium': '#a5a5a5',
        'ht-light': '#cccccc',
        'ht-lightest': '#f2f2f2',

        // Semantic color aliases
        'text-primary': '#595959',
        'text-secondary': '#7f7f7f',
        'text-muted': '#a5a5a5',

        'bg-primary': '#ffffff',
        'bg-secondary': '#f2f2f2',
        'bg-muted': '#cccccc',

        'border-primary': '#cccccc',
        'border-secondary': '#a5a5a5',
      },
    },
  },
  plugins: [],
};
