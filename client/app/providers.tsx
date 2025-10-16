"use client";

import type { ThemeProviderProps } from "next-themes";

import * as React from "react";
import { HeroUIProvider } from "@heroui/system";
import { ToastProvider } from "@heroui/toast";
import { useRouter } from "next/navigation";
import { ThemeProvider as NextThemesProvider } from "next-themes";
import { SWRConfig } from "swr";

import fetcher from "@/lib/swr";
import { initAuthRefresh } from "@/lib/auth";

export interface ProvidersProps {
  children: React.ReactNode;
  themeProps?: ThemeProviderProps;
}

declare module "@react-types/shared" {
  interface RouterConfig {
    routerOptions: NonNullable<
      Parameters<ReturnType<typeof useRouter>["push"]>[1]
    >;
  }
}

export function Providers({ children, themeProps }: ProvidersProps) {
  const router = useRouter();

  React.useEffect(() => {
    initAuthRefresh();
  }, []);

  return (
    <SWRConfig value={{ fetcher }}>
      <HeroUIProvider navigate={router.push}>
        <NextThemesProvider
          enableSystem
          attribute="class"
          defaultTheme="light"
          {...themeProps}
        >
          {children}
        </NextThemesProvider>
        {/* Toast region (does not accept children) */}
        <ToastProvider placement="top-right" />
      </HeroUIProvider>
    </SWRConfig>
  );
}
