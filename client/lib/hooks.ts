import { useEffect, useState } from "react";
import useSWRInfinite from "swr/infinite";
import { useApi } from "./swr";
import api from "./api";
import fetcher from "./swr";
import { MovieFilters, useAuthStore } from "./store";
import { ContinueWatchingMovie } from "@/types";

export const useSearchMovies = (query: string) => {
  // Only make API call if query is not empty
  const url = query.trim()
    ? `/movies/search?q=${encodeURIComponent(query.trim())}`
    : null;
  const { data, error, isLoading, mutate } = useApi(url);
  return {
    movies: data || [],
    isLoading,
    error,
    refetch: mutate,
  };
};

// hook for fetching movies list with infinite scroll
export const useMovies = (filters?: MovieFilters) => {
  const [mounted, setMounted] = useState(false);

  // prevent hydration mismatch by waiting for client mount
  useEffect(() => {
    setMounted(true);
  }, []);

  const getKey = (pageIndex: number, previousPageData: any[]) => {
    if (!mounted) return null;
    if (previousPageData && !previousPageData.length) return null; // reached end

    // If searching, use search endpoint (no pagination support in backend yet)
    if (filters?.query && filters.query.trim().length > 0) {
      return pageIndex === 0
        ? `/movies/search?q=${encodeURIComponent(filters.query)}`
        : null;
    }

    // Construct URL with filters for discovery endpoint
    const params = new URLSearchParams();

    params.set("page", (pageIndex + 1).toString());

    if (filters) {
      if (filters.selectedGenres.length > 0)
        params.set("genres", filters.selectedGenres.join(","));
      if (filters.minRating > 0)
        params.set("minRating", filters.minRating.toString());
      if (filters.sort) params.set("sort", filters.sort);
      if (filters.yearRange) {
        params.set("yearFrom", filters.yearRange[0].toString());
        params.set("yearTo", filters.yearRange[1].toString());
      }
    }

    return `/movies?${params.toString()}`;
  };

  const { data, error, isLoading, size, setSize, mutate } = useSWRInfinite(
    getKey,
    fetcher,
    {
      revalidateFirstPage: false,
      persistSize: true,
    }
  );

  const movies = data ? [].concat(...data) : [];
  const isLoadingMore =
    isLoading || (size > 0 && data && typeof data[size - 1] === "undefined");
  const isReachingEnd =
    data?.length === 0 || (data && data[data.length - 1]?.length < 20); // Assuming default limit is 20

  return {
    movies,
    isLoading: isLoading && movies.length === 0,
    isLoadingMore,
    isReachingEnd,
    size,
    setSize,
    error,
    refetch: mutate,
  };
};

// hook for fetching single movie details
export const useMovieDetailsReq = (movieId: string) => {
  const url = movieId ? `/movies/${movieId}` : null;
  const { data, error, isLoading, mutate } = useApi(url);

  return {
    movie: data,
    isLoading,
    error,
    refetch: mutate,
  };
};

// hook for user authentication status
export const useAuth = () => {
  const { data, error, isLoading } = useApi("/users/me");

  return {
    user: data,
    isLoading,
    error,
  };
};

// hook for current user data
export const useMe = () => {
  const { data, error, isLoading, mutate } = useApi("/users/me");

  return { user: data, isLoading, error, refetch: mutate };
};

// authenticate user with username and password
export const loginUser = async (username: string, password: string) => {
  const response = await api.post("/auth/login", { username, password });
  return response.data;
};

// register new user account
export const registerUser = async (
  username: string,
  email: string,
  password: string,
  firstName: string,
  lastName: string
) => {
  const response = await api.post("/auth/register", {
    username,
    email,
    password,
    FirstName: firstName,
    LastName: lastName,
  });
  return response.data;
};

// hook for user statistics (movies watched, hours, favorites)
export const useUserStats = () => {
  // TODO: Replace with real API endpoint when available
  // const { data, error, isLoading } = useApi('/users/stats');

  // For now, return placeholder data
  return {
    stats: {
      moviesWatched: 0,
      hoursWatched: 0,
      favorites: 0,
    },
    isLoading: false,
    error: null,
  };
};

// Type for API response (snake_case from backend)
type WatchHistoryItemResponse = {
  id?: number;
  movie_id: number;
  movie_title: string;
  poster_path?: string;
  watched_at?: string;
  duration?: number;
  last_position?: number;
  created_at?: string;
  updated_at?: string;
};

type WatchHistoryResponse = {
  history: WatchHistoryItemResponse[];
  total?: number;
};

// hook for user watch history and continue watching
export const useWatchHistory = () => {
  const [mounted, setMounted] = useState(false);
  const { data, error, isLoading, mutate } = useApi(
    mounted ? "/users/watch-history" : null
  );

  // prevent hydration mismatch by waiting for client mount
  useEffect(() => {
    setMounted(true);
  }, []);

  const apiResponse = data as WatchHistoryResponse | undefined;
  const history = apiResponse?.history || [];

  // Transform API response to match ContinueWatchingMovie type
  const continueWatching =
    history.length > 0
      ? (() => {
          const item = history[0];
          // Extract year from watched_at or use current year as default
          const watchedDate = item.watched_at
            ? new Date(item.watched_at)
            : null;
          const year = watchedDate
            ? watchedDate.getFullYear()
            : new Date().getFullYear();

          return {
            id: item.movie_id,
            title: item.movie_title || "Unknown Movie",
            posterPath: item.poster_path || undefined,
            genre: "Movie", // Default since API doesn't provide genre - could be enhanced by fetching movie details
            year: year,
          } as ContinueWatchingMovie;
        })()
      : null;

  return {
    watchHistory: history,
    continueWatching,
    isLoading: !mounted || isLoading,
    error,
    refetch: mutate,
  };
};

// logout user and clear tokens
export const logoutUser = async () => {
  try {
    await api.post("/auth/logout");
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error("logout error:", error);
  } finally {
    // clear tokens from storage and redirect
    if (typeof window !== "undefined") {
      localStorage.removeItem("token");
      document.cookie = "token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";

      // Clear Zustand store
      useAuthStore.getState().logout();

      window.location.href = "/auth/login";
    }
  }
};

// request password reset email
export const forgotPassword = async (email: string) => {
  const response = await api.post("/forgot-password", { email });

  return response.data;
};

// reset password with token
export const resetPassword = async (
  token: string,
  email: string,
  password: string
) => {
  const response = await api.post("/reset-password", {
    token,
    email,
    password,
  });

  return response.data;
};

// post a new comment
export const addComment = async (
  movieId: number,
  content: string,
  username: string
) => {
  const response = await api.post("/comments/add", {
    movie_id: movieId,
    content,
    username,
  });

  return response.data;
};

// update user profile
export const updateUser = async (data: {
  firstname?: string;
  lastname?: string;
  email?: string;
  username?: string;
  password?: string;
  preferred_language?: string;
}) => {
  const response = await api.patch("/users/me", data);

  return response.data;
};

// upload user avatar
export const uploadAvatar = async (file: File) => {
  const formData = new FormData();
  formData.append("picture", file);

  const response = await api.post("/users/me/upload-avatar", formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });

  return response.data;
};
