// src/components/movies/useInfiniteMovies.ts
"use client";

import React from "react";
import type { Movie } from "./MoviesCard";

type MoviesResponse = {
  items: Movie[];
  nextCursor?: string | null;
};

export function useInfiniteMovies() {
  const [query, setQuery] = React.useState("");
  const [sort, setSort] = React.useState<"name" | "year" | "rating">("name");
  const [selectedGenres, setSelectedGenres] = React.useState<string[]>([]);
  const [minRating, setMinRating] = React.useState<number>(0);
  const [yearRange, setYearRange] = React.useState<[number, number]>([2000, new Date().getFullYear()]);

  const [movies, setMovies] = React.useState<Movie[]>([]);
  const [cursor, setCursor] = React.useState<string | null>(null);
  const [hasMore, setHasMore] = React.useState(true);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isFetchingMore, setIsFetchingMore] = React.useState(false);

  const params = React.useMemo(() => {
    const p = new URLSearchParams();
    if (query.trim()) p.set("q", query.trim());
    p.set("sort", sort);
    if (selectedGenres.length) p.set("genres", selectedGenres.join(","));
    p.set("minRating", String(minRating));
    p.set("yearFrom", String(yearRange[0]));
    p.set("yearTo", String(yearRange[1]));
    return p;
  }, [query, sort, selectedGenres, minRating, yearRange]);

  const fetchPage = React.useCallback(async (reset = false) => {
    try {
      if (reset) {
        setIsLoading(true);
        setCursor(null);
        setHasMore(true);
      } else {
        setIsFetchingMore(true);
      }

      const qp = new URLSearchParams(params);
      if (!reset && cursor) qp.set("cursor", cursor);

      const res = await fetch(`/api/movies?${qp.toString()}`, { cache: "no-store" });
      if (!res.ok) throw new Error("Failed to fetch movies");
      const data: MoviesResponse = await res.json();

      setMovies((prev) => (reset ? data.items : [...prev, ...data.items]));
      setCursor(data.nextCursor ?? null);
      setHasMore(Boolean(data.nextCursor));
    } catch (e) {
      console.error(e);
    } finally {
      setIsLoading(false);
      setIsFetchingMore(false);
    }
  }, [params, cursor]);

  // Initial + whenever filters change -> reset and fetch
  React.useEffect(() => {
    const id = setTimeout(() => {
      fetchPage(true);
    }, 250); // small debounce for search typing
    return () => clearTimeout(id);
  }, [params, fetchPage]);

  const fetchNext = React.useCallback(() => {
    if (!hasMore || isFetchingMore) return;
    fetchPage(false);
  }, [hasMore, isFetchingMore, fetchPage]);

  const toggleGenre = React.useCallback((g: string) => {
    setSelectedGenres((prev) =>
      prev.includes(g) ? prev.filter((x) => x !== g) : [...prev, g]
    );
  }, []);

  const resetAndRefetch = React.useCallback(() => {
    setQuery("");
    setSort("name");
    setSelectedGenres([]);
    setMinRating(0);
    setYearRange([2000, new Date().getFullYear()]);
    // fetch will auto-trigger via effect
  }, []);

  return {
    // filters
    query, setQuery,
    sort, setSort,
    selectedGenres, toggleGenre,
    minRating, setMinRating,
    yearRange, setYearRange,

    // data
    movies, isLoading, fetchNext, hasMore, isFetchingMore,

    // helpers
    resetAndRefetch
  };
}
