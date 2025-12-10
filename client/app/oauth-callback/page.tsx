"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { addToast } from "@heroui/toast";

import { useAuthStore } from "@/lib/store";
import { setTokens, getCookie } from "@/lib/auth";
import api from "@/lib/api";

export default function OAuthCallbackPage() {
  const router = useRouter();
  const { login } = useAuthStore();

  useEffect(() => {
    const handleOAuthCallback = async () => {
      const accessToken = getCookie("AccessToken");
      const refreshToken = getCookie("RefreshToken");
      const expiresIn = getCookie("AccessTokenExpiresIn");

      if (accessToken && refreshToken) {
        type RevokeTokenRes = {
          AccessToken: string;
          RefreshToken: string;
          TokenType: string;
          ExpiresIn: number;
          RefreshTokenExpiresIn: number;
        };

        const tokenData: RevokeTokenRes = {
          AccessToken: accessToken,
          RefreshToken: refreshToken,
          TokenType: "Bearer",
          ExpiresIn: Number(expiresIn) || 3600,
          RefreshTokenExpiresIn: 60 * 60 * 24 * 7,
        };

        setTokens(tokenData);

        document.cookie = "AccessToken=; path=/; max-age=0";
        document.cookie = "RefreshToken=; path=/; max-age=0";
        document.cookie = "AccessTokenExpiresIn=; path=/; max-age=0";

        try {
          const userResponse = await api.get("/users/me");
          const userData = userResponse.data;

          login(userData, accessToken);

          addToast({
            title: "Login successful",
            description: "Welcome back!",
            severity: "success",
            timeout: 3000,
          });

          router.push("/app/discover");
        } catch (error) {
          addToast({
            title: "Login failed",
            description: "Could not retrieve user information. Please try again.",
            severity: "danger",
            timeout: 4000,
          });
          router.push("/auth/login");
        }
      } else {
        addToast({
          title: "Login failed",
          description: "No authentication tokens received.",
          severity: "danger",
          timeout: 4000,
        });
        router.push("/auth/login");
      }
    };

    handleOAuthCallback();
  }, [router, login]);

  return (
    <div className="flex h-screen w-full items-center justify-center bg-black text-white">
      <div className="flex flex-col items-center gap-4">
        <div className="h-12 w-12 animate-spin rounded-full border-4 border-white/20 border-t-white" />
        <p className="text-white/80 animate-pulse">Authenticating...</p>
      </div>
    </div>
  );
}
