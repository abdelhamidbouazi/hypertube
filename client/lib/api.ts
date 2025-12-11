import axios from "axios";
import { addToast } from "@heroui/toast";
import { getErrorMessage } from "./error-utils";

export const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "/api";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "/api",
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: false,
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

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: any) => void;
  reject: (reason?: any) => void;
}> = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });

  failedQueue = [];
};

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    const skipRefresh =
      originalRequest?.skipAuthRedirect ||
      error.response?.status !== 401 ||
      typeof window === "undefined" ||
      window.location.pathname.includes("/auth/login");

    // 401 Unauthorized error handling

    if (skipRefresh) {
      return Promise.reject(error);
    }

    if (originalRequest._retry) {
      addToast({
        title: "Authentication failed",
        description: "Session expired. Please log in again.",
        severity: "danger",
        timeout: 4000,
      });
      localStorage.removeItem("token");
      document.cookie = "token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";
      document.cookie =
        "refreshToken=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";
      window.location.href = "/auth/login";
      return Promise.reject(error);
    }

    if (isRefreshing) {
      // Queue request while token is being refreshed
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      })
        .then(() => {
          return api(originalRequest);
        })
        .catch((err) => {
          return Promise.reject(err);
        });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    return new Promise((resolve, reject) => {
      import("./auth")
        .then((authModule) => authModule.refreshAccessToken())
        .then((success) => {
          if (success) {
            processQueue(null, null);
            resolve(api(originalRequest));
          } else {
            addToast({
              title: "Authentication failed",
              description: "Token refresh failed. Please log in again.",
              severity: "danger",
              timeout: 4000,
            });
            processQueue(error, null);
            reject(error);
          }
        })
        .catch((err) => {
          addToast({
            title: "Authentication error",
            description: getErrorMessage(err),
            severity: "danger",
            timeout: 4000,
          });
          processQueue(err, null);
          reject(err);
        })
        .finally(() => {
          isRefreshing = false;
        });
    });
  }
);

export default api;
