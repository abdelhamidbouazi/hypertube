"use client";

import React from "react";
import type { Movie } from "./MoviesCard";

// hook for managing movie filters and state
export function useMovies() {
  // filter state
  const [query, setQuery] = React.useState("");
  const [sort, setSort] = React.useState<"name" | "year" | "rating">("name");
  const [selectedGenres, setSelectedGenres] = React.useState<string[]>([]);
  const [minRating, setMinRating] = React.useState<number>(0);
  const [yearRange, setYearRange] = React.useState<[number, number]>([2000, new Date().getFullYear()]);

  // movie data state
  const [movies, setMovies] = React.useState<Movie[]>([]);
  const [isLoading, setIsLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  // toggle genre selection
  const toggleGenre = React.useCallback((g: string) => {
    setSelectedGenres((prev) =>
      prev.includes(g) ? prev.filter((x) => x !== g) : [...prev, g]
    );
  }, []);

  // reset all filters to default
  const resetAndRefetch = React.useCallback(() => {
    setQuery("");
    setSort("name");
    setSelectedGenres([]);
    setMinRating(0);
    setYearRange([2000, new Date().getFullYear()]);
  }, []);

  return {
    query, setQuery,
    sort, setSort,
    selectedGenres, toggleGenre,
    minRating, setMinRating,
    yearRange, setYearRange,

    movies, isLoading, error,

    resetAndRefetch
  };
}
