"use client";

import type { SearchResult } from "@/types";

import { useState, useEffect, useCallback, useRef } from "react";

import api from "@/lib/api";

interface UseSearchResult {
  query: string;
  setQuery: (query: string) => void;
  results: SearchResult[];
  isLoading: boolean;
  error: string | null;
  isOpen: boolean;
  setIsOpen: (open: boolean) => void;
  clearSearch: () => void;
  searchMovie: (movieQuery: string, year?: string) => Promise<void>;
}

export function useSearch(debounceMs: number = 300): UseSearchResult {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isOpen, setIsOpen] = useState(false);

  const debounceTimeoutRef = useRef<NodeJS.Timeout>();
  const abortControllerRef = useRef<AbortController>();

  const searchMovie = useCallback(async (movieQuery: string, year?: string) => {
    if (!movieQuery.trim()) {
      setResults([]);
      setIsOpen(false);

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

      const params = new URLSearchParams({
        q: movieQuery.trim(),
      });

      if (year) {
        params.append("year", year);
      }

      const response = await api.get(`/movies/search?${params}`, {
        signal: abortControllerRef.current.signal,
      });

      if (response.data && Array.isArray(response.data)) {
        setResults(response.data);
        setIsOpen(response.data.length > 0);
      } else {
        setResults([]);
        setIsOpen(false);
      }
    } catch (err: any) {
      if (err.name !== "AbortError") {
        // eslint-disable-next-line no-console
        console.error("Search error:", err);
        setError(err.response?.data?.message || "Search failed");
        setResults([]);
        setIsOpen(false);
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  const debouncedSearch = useCallback(
    (searchQuery: string) => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }

      debounceTimeoutRef.current = setTimeout(() => {
        searchMovie(searchQuery);
      }, debounceMs);
    },
    [searchMovie, debounceMs]
  );

  const handleQueryChange = useCallback(
    (newQuery: string) => {
      setQuery(newQuery);
      debouncedSearch(newQuery);
    },
    [debouncedSearch]
  );

  const clearSearch = useCallback(() => {
    setQuery("");
    setResults([]);
    setIsOpen(false);
    setError(null);

    if (debounceTimeoutRef.current) {
      clearTimeout(debounceTimeoutRef.current);
    }

    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  return {
    query,
    setQuery: handleQueryChange,
    results,
    isLoading,
    error,
    isOpen,
    setIsOpen,
    clearSearch,
    searchMovie,
  };
}
