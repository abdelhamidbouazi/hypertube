"use client";

import type { ThemeProviderProps } from "next-themes";

import * as React from "react";
import { HeroUIProvider } from "@heroui/system";
import { SWRConfig } from "swr";
import { useRouter } from "next/navigation";
import { ToastProvider } from "@heroui/toast";
import { ThemeProvider as NextThemesProvider } from "next-themes";

import fetcher from "@/lib/swr";

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

  return (
    <SWRConfig value={{ fetcher }}>
      <NextThemesProvider {...themeProps}>
        <HeroUIProvider navigate={router.push}>
          {children}
          <ToastProvider placement="top-right" />
        </HeroUIProvider>
      </NextThemesProvider>
    </SWRConfig>
  );
}
