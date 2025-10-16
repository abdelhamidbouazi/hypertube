/* eslint-disable */
import { useEffect, useState } from "react";

import { useApi } from "./swr";
import api from "./api";
import { clearTokens, getRefreshToken } from "./auth";
// import type { Movie } from '@/components/movies/MoviesCard';

export const useMovies = () => {
  const [mounted, setMounted] = useState(false);
  const { data, error, isLoading, mutate } = useApi(mounted ? "/movies" : null);

  useEffect(() => {
    setMounted(true);
  }, []);

  return {
    movies: data || [],
    isLoading: !mounted || isLoading,
    error,
    refetch: mutate,
  };
};

export const useInfiniteMovies = () => {
  const [movies, setMovies] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadMovies = async (page: number, append: boolean = false) => {
    try {
      if (append) {
        setIsLoadingMore(true);
      } else {
        setIsLoading(true);
        setMovies([]);
      }
      setError(null);

      const response = await api.get(`/movies?page=${page}`);
      const newMovies = response.data || [];

      if (append) {
        setMovies((prev) => {
          const existingIds = new Set(prev.map((movie: any) => movie.id));
          const uniqueNewMovies = newMovies.filter(
            (movie: any) => !existingIds.has(movie.id)
          );
          return [...prev, ...uniqueNewMovies];
        });
      } else {
        setMovies(newMovies);
      }

      const hasMore = newMovies.length === 20;
      return { movies: newMovies, hasMore };
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load movies");
      return { movies: [], hasMore: false };
    } finally {
      setIsLoading(false);
      setIsLoadingMore(false);
    }
  };

  return {
    movies,
    isLoading,
    isLoadingMore,
    error,
    loadMovies,
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
  const { data, error, isLoading } = useApi("/users/me");

  return {
    user: data,
    isLoading,
    error,
  };
};

export const loginUser = async (email: string, password: string) => {
  const response = await api.post(
    "/auth/login",
    { email, password },

    { skipAuthRedirect: true } as any
  );

  return response.data;
};

export const registerUser = async (
  email: string,
  password: string,
  firstName: string,
  lastName: string
) => {
  const response = await api.post("/auth/register", {
    email,
    password,
    FirstName: firstName,
    LastName: lastName,
  });

  return response.data;
};

export const logoutUser = async () => {
  try {
    const refresh = getRefreshToken();
    await api.delete("/auth/logout", {
      headers: refresh ? { RefreshToken: refresh } : undefined,
    });
  } catch {
    // ignore
  } finally {
    clearTokens();
    if (typeof window !== "undefined") window.location.href = "/auth/login";
  }
};
