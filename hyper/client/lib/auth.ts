"use client";
/* eslint-disable */

import api from "./api";
import { addToast } from "@heroui/toast";
import { getErrorMessage } from "./error-utils";

type RevokeTokenRes = {
  AccessToken: string;
  RefreshToken: string;
  TokenType: string;
  ExpiresIn: number;
  RefreshTokenExpiresIn: number;
};

const COOKIE_MAX_AGE = 60 * 60 * 24 * 7;

let refreshTimer: ReturnType<typeof setTimeout> | null = null;

export function getCookie(name: string): string | null {
  if (typeof document === "undefined") return null;
  const match = document.cookie
    .split(";")
    .map((c) => c.trim())
    .find((c) => c.startsWith(name + "="));

  return match ? decodeURIComponent(match.split("=")[1]) : null;
}

function setCookie(name: string, value: string, maxAgeSec: number) {
  const secure =
    typeof location !== "undefined" && location.protocol === "https:";

  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${maxAgeSec}; samesite=lax${secure ? "; secure" : ""}`;
}

export function clearTokens() {
  try {
    localStorage.removeItem("token");
  } catch {}
  document.cookie =
    "token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; samesite=lax";
  document.cookie =
    "refreshToken=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; samesite=lax";
  document.cookie =
    "token_exp=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; samesite=lax";
  document.cookie =
    "refresh_exp=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; samesite=lax";
  if (refreshTimer) {
    clearTimeout(refreshTimer);
    refreshTimer = null;
  }
}

export function setTokens(data: RevokeTokenRes) {
  try {
    localStorage.setItem("token", data.AccessToken);
  } catch {}
  setCookie("token", data.AccessToken, COOKIE_MAX_AGE);
  setCookie("refreshToken", data.RefreshToken, COOKIE_MAX_AGE);
  setCookie("token_exp", String(data.ExpiresIn), COOKIE_MAX_AGE);
  setCookie("refresh_exp", String(data.RefreshTokenExpiresIn), COOKIE_MAX_AGE);
  scheduleAutoRefresh(data.ExpiresIn);
}

export function getAccessToken(): string | null {
  if (typeof window !== "undefined") {
    const ls = localStorage.getItem("token");

    if (ls) return ls;
  }

  return getCookie("token");
}

export function getRefreshToken(): string | null {
  return getCookie("refreshToken");
}

export async function refreshAccessToken(): Promise<boolean> {
  const access = getAccessToken();
  const refresh = getRefreshToken();

  if (!access || !refresh) {
    addToast({
      title: "Authentication error",
      description: "Missing tokens, cannot refresh session",
      severity: "warning",
      timeout: 3000,
    });
    return false;
  }
  try {
    const res = await api.post("/auth/refreshToken", {}, {
      headers: { RefreshToken: refresh },
      skipAuthRedirect: true as any,
    } as any);
    const data: RevokeTokenRes = res.data;

    setTokens(data);
    return true;
  } catch (error) {
    addToast({
      title: "Session expired",
      description: "Please log in again to continue",
      severity: "warning",
      timeout: 4000,
    });
    clearTokens();
    if (typeof window !== "undefined") {
      window.location.href = "/auth/login";
    }

    return false;
  }
}

export function scheduleAutoRefresh(expUnixSeconds?: number) {
  if (refreshTimer) {
    clearTimeout(refreshTimer);
    refreshTimer = null;
  }

  const now = Date.now();
  const expSec = expUnixSeconds ?? Number(getCookie("token_exp") || 0);

  if (!expSec || Number.isNaN(expSec)) {
    return;
  }

  const msUntilRefresh = Math.max(expSec * 1000 - now - 60_000, 5_000);

  refreshTimer = setTimeout(() => {
    refreshAccessToken();
  }, msUntilRefresh);
}

export function initAuthRefresh() {
  const exp = Number(getCookie("token_exp") || 0);

  if (exp) {
    scheduleAutoRefresh(exp);
  }

  if (typeof document !== "undefined") {
    const handler = () => {
      if (document.visibilityState === "visible") {
        const exp = Number(getCookie("token_exp") || 0);
        const now = Date.now();
        const timeLeft = exp * 1000 - now;

        if (exp && timeLeft < 90_000) {
          refreshAccessToken();
        } else if (exp) {
          scheduleAutoRefresh(exp);
        }
      }
    };

    document.addEventListener("visibilitychange", handler);
  }
}
