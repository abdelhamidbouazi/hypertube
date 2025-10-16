/* eslint-disable */
import axios from "axios";

import { clearTokens, getAccessToken, refreshAccessToken } from "./auth";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "/api",
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true,
});

api.interceptors.request.use(
  (config) => {
    if (typeof window !== "undefined") {
      let token = localStorage.getItem("token");

      if (!token) {
        const cookies = document.cookie.split(";");
        const tokenCookie = cookies.find((cookie) =>
          cookie.trim().startsWith("token=")
        );

        if (tokenCookie) {
          token = tokenCookie.split("=")[1];
        }
      }

      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }

    return config;
  },
  (error) => Promise.reject(error)
);

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error?.config || {};
    const skipAuthRedirect = (original as any)?.skipAuthRedirect;

    // Try a single refresh when we get a 401 and we haven't retried yet
    if (
      error?.response?.status === 401 &&
      !skipAuthRedirect &&
      !original._retry &&
      typeof window !== "undefined"
    ) {
      original._retry = true;
      const refreshed = await refreshAccessToken();

      if (refreshed) {
        const token = getAccessToken();

        if (token)
          original.headers = {
            ...original.headers,
            Authorization: `Bearer ${token}`,
          };

        return api(original);
      }
    }

    if (
      !skipAuthRedirect &&
      error?.response?.status === 401 &&
      typeof window !== "undefined"
    ) {
      clearTokens();
      window.location.href = "/auth/login";
    }

    return Promise.reject(error);
  }
);

export default api;
