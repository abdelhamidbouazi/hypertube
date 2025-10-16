import "@/styles/globals.css";
import { Metadata, Viewport } from "next";
import clsx from "clsx";

import { Providers } from "./providers";

import { siteConfig } from "@/config/site";
import { fontSans } from "@/config/fonts";

export const metadata: Metadata = {
  title: {
    default: siteConfig.name,
    template: `%s - ${siteConfig.name}`,
  },
  description: siteConfig.description,
  icons: {
    icon: "/favicon.ico",
  },
};

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "white" },
    { media: "(prefers-color-scheme: dark)", color: "black" },
  ],
};

export default function AuthenthicatedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      suppressHydrationWarning
      className={clsx(fontSans.variable, "min-h-screen bg-background")}
      lang="en"
    >
      <head />
      <body>
        <Providers themeProps={{ attribute: "class", defaultTheme: "light" }}>
          <main>{children}</main>
        </Providers>
      </body>
    </html>
  );
}

// <html suppressHydrationWarning lang="en">
//   <head />
//   <body
//     className={clsx(
//       fontSans.variable,
//       "min-h-screen text-foreground bg-background font-sans antialiased"
//     )}
//   >
// </body>
// </html>
