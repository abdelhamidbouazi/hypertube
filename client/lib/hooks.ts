import { useEffect, useState } from "react";
import useSWRInfinite from "swr/infinite";
import { useApi } from "./swr";
import api from "./api";
import fetcher from "./swr";
import { MovieFilters, useAuthStore } from "./store";
import { ContinueWatchingMovie } from "@/types";
import { addToast } from "@heroui/toast";
import { getErrorMessage } from "./error-utils";

export const useSearchMovies = (query: string) => {
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

export const useMovies = (filters?: MovieFilters) => {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const getKey = (pageIndex: number, previousPageData: any[]) => {
    if (!mounted) return null;
    if (previousPageData && !previousPageData.length) return null;

    if (filters?.query && filters.query.trim().length > 0) {
      return pageIndex === 0
        ? `/movies/search?q=${encodeURIComponent(filters.query)}`
        : null;
    }

    const hasFilters =
      (filters?.selectedGenres && filters.selectedGenres.length > 0) ||
      (filters?.minRating && filters.minRating > 0) ||
      (filters?.sort && filters.sort !== "popularity") ||
      (filters?.yearRange &&
        (filters.yearRange[0] !== 2000 ||
          filters.yearRange[1] !== new Date().getFullYear()));

    if (!hasFilters) {
      const page = pageIndex + 1;
      return `/movies/popular?page=${page}`;
    }

    const params = new URLSearchParams();
    params.set("page", (pageIndex + 1).toString());

    if (filters) {
      if (filters.selectedGenres.length > 0)
        params.set("genres", filters.selectedGenres.join(","));
      if (filters.minRating > 0)
        params.set("minRating", filters.minRating.toString());
      if (filters.sort && filters.sort !== "popularity")
        params.set("sort", filters.sort);
      if (filters.yearRange) {
        params.set("yearFrom", filters.yearRange[0].toString());
        params.set("yearTo", filters.yearRange[1].toString());
      }
    }

    return `/movies/popular?${params.toString()}`;
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
    data?.length === 0 || (data && data[data.length - 1]?.length < 20);

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

export const useAuth = () => {
  const [hasToken, setHasToken] = useState<boolean | null>(null);

  useEffect(() => {
    if (typeof window !== "undefined") {
      const token = localStorage.getItem("token");
      setHasToken(!!token);
    }
  }, []);

  const { data, error, isLoading } = useApi(
    hasToken ? "/users/me" : null
  );

  return {
    user: data,
    isLoading,
    error,
  };
};

export const useMe = () => {
  const [hasToken, setHasToken] = useState<boolean | null>(null);

  useEffect(() => {
    if (typeof window !== "undefined") {
      const token = localStorage.getItem("token");
      setHasToken(!!token);
    }
  }, []);

  const { data, error, isLoading, mutate } = useApi(
    hasToken ? "/users/me" : null
  );

  return { user: data, isLoading, error, refetch: mutate };
};

export const loginUser = async (username: string, password: string) => {
  const response = await api.post("/auth/login", { username, password });
  return response.data;
};

export const registerUser = async (
  username: string,
  email: string,
  password: string,
  firstName: string,
  lastName: string,
  avatar?: File
) => {
  const formData = new FormData();
  formData.append("username", username);
  formData.append("email", email);
  formData.append("password", password);
  formData.append("firstname", firstName);
  formData.append("lastname", lastName);

  if (avatar) {
    formData.append("picture", avatar);
  }

  const response = await api.post("/auth/register", formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });
  return response.data;
};

export const useUserStats = () => {
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

export const useWatchHistory = () => {
  const [mounted, setMounted] = useState(false);
  const [hasToken, setHasToken] = useState<boolean | null>(null);

  useEffect(() => {
    setMounted(true);
    if (typeof window !== "undefined") {
      const token = localStorage.getItem("token");
      setHasToken(!!token);
    }
  }, []);

  const { data, error, isLoading, mutate } = useApi(
    mounted && hasToken ? "/users/watch-history" : null
  );

  const apiResponse = data as WatchHistoryResponse | undefined;
  const history = apiResponse?.history || [];

  const firstItem = history.length > 0 ? history[0] : null;
  const movieId = firstItem?.movie_id;
  const { movie: movieData } = useMovieDetailsReq(
    movieId && (!firstItem.movie_title || firstItem.movie_title.trim() === "")
      ? movieId.toString()
      : ""
  );

  const continueWatching = firstItem
    ? (() => {
        const watchedDate = firstItem.watched_at
          ? new Date(firstItem.watched_at)
          : null;
        const year = watchedDate
          ? watchedDate.getFullYear()
          : new Date().getFullYear();

        const title =
          firstItem.movie_title && firstItem.movie_title.trim() !== ""
            ? firstItem.movie_title
            : movieData?.title || "Unknown title";

        const rawPosterPath = firstItem.poster_path || movieData?.poster_path;
        const posterPath = rawPosterPath
          ? rawPosterPath.startsWith("http")
            ? rawPosterPath
            : `https://image.tmdb.org/t/p/w185${rawPosterPath}`
          : undefined;

        return {
          id: firstItem.movie_id,
          title: title,
          posterPath: posterPath,
          genre: "Movie",
          year: year,
        } as ContinueWatchingMovie;
      })()
    : null;

  return {
    watchHistory: history,
    continueWatching,
    isLoading: !mounted || hasToken === null || isLoading,
    error,
    refetch: mutate,
  };
};

export const logoutUser = async () => {
  try {
    await api.post("/auth/logout");
  } catch (error) {
    addToast({
      title: "Logout error",
      description: getErrorMessage(error),
      severity: "warning",
      timeout: 3000,
    });
  } finally {
    if (typeof window !== "undefined") {
      localStorage.removeItem("token");
      document.cookie = "token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";

      useAuthStore.getState().logout();

      window.location.href = "/auth/login";
    }
  }
};

export const forgotPassword = async (email: string) => {
  const response = await api.post("/forgot-password", { email });

  return response.data;
};

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

export const getUserPublicInfo = async (username: string) => {
  const response = await api.get(`/users/${encodeURIComponent(username)}`);
  return response.data;
};
