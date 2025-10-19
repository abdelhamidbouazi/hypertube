"use client";

import React from "react";
import type { Movie } from "./MoviesCard";

export function useMovies() {
  const [query, setQuery] = React.useState("");
  const [sort, setSort] = React.useState<"name" | "year" | "rating">("name");
  const [selectedGenres, setSelectedGenres] = React.useState<string[]>([]);
  const [minRating, setMinRating] = React.useState<number>(0);
  const [yearRange, setYearRange] = React.useState<[number, number]>([2000, new Date().getFullYear()]);

  const [movies, setMovies] = React.useState<Movie[]>([]);
  const [isLoading, setIsLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

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
  }, []);

  return {
    // filters
    query, setQuery,
    sort, setSort,
    selectedGenres, toggleGenre,
    minRating, setMinRating,
    yearRange, setYearRange,

    // data
    movies, isLoading, error,

    // helpers
    resetAndRefetch
  };
}
