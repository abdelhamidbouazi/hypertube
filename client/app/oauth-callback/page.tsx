"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { getCookie, setTokens, scheduleAutoRefresh } from "@/lib/auth";

export default function OAuthCallbackPage() {
  const router = useRouter();

  useEffect(() => {
    const accessToken = getCookie("AccessToken");
    const refreshToken = getCookie("RefreshToken");
    const exp = Number(getCookie("AccessTokenExpiresIn") || 0);

    if (accessToken && refreshToken && exp) {
      setTokens({
        AccessToken: accessToken,
        RefreshToken: refreshToken,
        TokenType: "Bearer",
        ExpiresIn: exp,
        RefreshTokenExpiresIn: exp + 60 * 60 * 24 * 7,
      });
      scheduleAutoRefresh(exp);
    }

    document.cookie = `AccessToken=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; samesite=lax`;
    document.cookie = `RefreshToken=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; samesite=lax`;
    document.cookie = `AccessTokenExpiresIn=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; samesite=lax`;

    router.replace(accessToken ? "/app/discover" : "/auth/login");
  }, [router]);

  return (
    <div className="flex min-h-[60vh] items-center justify-center text-white">
      <div className="rounded-2xl border border-white/15 bg-white/10 px-6 py-4 shadow-lg backdrop-blur">
        Finalizing sign-in…
      </div>
    </div>
  );
}
