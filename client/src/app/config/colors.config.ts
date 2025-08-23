// Color palette configuration for HyperTybe
export const colors = {
  // Primary grayscale palette
  primary: {
    darkest: '#595959', // Main headings, primary buttons
    dark: '#7f7f7f', // Hover states, labels, secondary text
    medium: '#a5a5a5', // Placeholder text, descriptions
    light: '#cccccc', // Borders, dividers
    lightest: '#f2f2f2', // Input backgrounds, subtle fills
  },

  // Semantic colors using the palette
  text: {
    primary: '#595959',
    secondary: '#7f7f7f',
    muted: '#a5a5a5',
  },

  background: {
    primary: '#ffffff',
    secondary: '#f2f2f2',
    muted: '#cccccc',
  },

  border: {
    primary: '#cccccc',
    secondary: '#a5a5a5',
  },

  button: {
    primary: {
      bg: '#595959',
      hover: '#7f7f7f',
      focus: '#7f7f7f',
    },
    secondary: {
      bg: 'transparent',
      border: '#cccccc',
      text: '#595959',
      hover: '#f2f2f2',
    },
  },

  input: {
    bg: '#f2f2f2',
    focus: '#ffffff',
    border: '#cccccc',
    placeholder: '#a5a5a5',
  },
} as const;

// Tailwind class names for easy use
export const tailwindColors = {
  primary: 'ht-darkest',
  secondary: 'ht-dark',
  muted: 'ht-medium',
  light: 'ht-light',
  lightest: 'ht-lightest',

  text: {
    primary: 'text-ht-darkest',
    secondary: 'text-ht-dark',
    muted: 'text-ht-medium',
  },

  bg: {
    primary: 'bg-white',
    secondary: 'bg-ht-lightest',
    muted: 'bg-ht-light',
  },

  border: {
    primary: 'border-ht-light',
    secondary: 'border-ht-medium',
  },
} as const;

// CSS custom properties generator
export const generateCSSVars = () => {
  return `
    :root {
      --color-primary-darkest: ${colors.primary.darkest};
      --color-primary-dark: ${colors.primary.dark};
      --color-primary-medium: ${colors.primary.medium};
      --color-primary-light: ${colors.primary.light};
      --color-primary-lightest: ${colors.primary.lightest};
      
      --color-text-primary: ${colors.text.primary};
      --color-text-secondary: ${colors.text.secondary};
      --color-text-muted: ${colors.text.muted};
      
      --color-bg-primary: ${colors.background.primary};
      --color-bg-secondary: ${colors.background.secondary};
      --color-bg-muted: ${colors.background.muted};
      
      --color-border-primary: ${colors.border.primary};
      --color-border-secondary: ${colors.border.secondary};
      
      --color-btn-primary-bg: ${colors.button.primary.bg};
      --color-btn-primary-hover: ${colors.button.primary.hover};
      --color-btn-primary-focus: ${colors.button.primary.focus};
      
      --color-btn-secondary-bg: ${colors.button.secondary.bg};
      --color-btn-secondary-border: ${colors.button.secondary.border};
      --color-btn-secondary-text: ${colors.button.secondary.text};
      --color-btn-secondary-hover: ${colors.button.secondary.hover};
      
      --color-input-bg: ${colors.input.bg};
      --color-input-focus: ${colors.input.focus};
      --color-input-border: ${colors.input.border};
      --color-input-placeholder: ${colors.input.placeholder};
    }
  `;
};
