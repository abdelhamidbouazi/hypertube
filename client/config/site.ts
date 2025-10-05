export type SiteConfig = typeof siteConfig;

export const siteConfig = {
  name: "Cinéthos",
  description: "Your best watching destintion, Watch with love",
  navItems: [
    {
      label: "Movies",
      href: "/app/discover",
    },
    {
      label: "Categories",
      href: "/app/categories",
    },
    {
      label: "About",
      href: "/app/about",
    },
  ],
  navMenuItems: [
    {
      label: "Discover",
      href: "/app/discover",
    },
    
    {
      label: "Team",
      href: "/app/team",
    },
    {
      label: "Settings",
      href: "/app/settings",
    },
    {
      label: "Logout",
      href: "/app/logout",
    },
  ],
  links: {
    github: "https://github.com/abdelhamidbouazi",
    X: "https://x.com/abdelhamidbouazi",
    porfolio: "https://abouazi.me",
  },
};
