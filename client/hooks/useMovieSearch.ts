"use client";

import type { SearchResult } from "@/types";

import { useState, useCallback, useRef } from "react";

import api from "@/lib/api";

interface UseMovieSearchResult {
  results: SearchResult[];
  isLoading: boolean;
  error: string | null;
  hasSearched: boolean;
  searchMovies: (query: string) => Promise<void>;
  clearResults: () => void;
}

export function useMovieSearch(): UseMovieSearchResult {
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasSearched, setHasSearched] = useState(false);

  const abortControllerRef = useRef<AbortController>();

  const searchMovies = useCallback(async (query: string) => {
    if (!query.trim()) {
      setResults([]);
      setError(null);
      setHasSearched(false);

      return;
    }

    // Cancel previous request
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // Create new abort controller
    abortControllerRef.current = new AbortController();

    try {
      setIsLoading(true);
      setError(null);
      setHasSearched(true);

      const params = new URLSearchParams({
        q: query.trim(),
      });

      const response = await api.get(`/movies/search?${params}`, {
        signal: abortControllerRef.current.signal,
      });

      if (response.data && Array.isArray(response.data)) {
        setResults(response.data);
      } else {
        setResults([]);
      }
    } catch (err: any) {
      if (err.name !== "AbortError") {
        // eslint-disable-next-line no-console
        console.error("Movie search error:", err);
        setError(err.response?.data?.message || "Search failed");
        setResults([]);
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  const clearResults = useCallback(() => {
    setResults([]);
    setError(null);
    setHasSearched(false);

    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
  }, []);

  return {
    results,
    isLoading,
    error,
    hasSearched,
    searchMovies,
    clearResults,
  };
}
