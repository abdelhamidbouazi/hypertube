"use client";

import React from "react";
import type { Movie } from "./MoviesCard";

export type MovieDetails = Movie & {
  description?: string;
  director?: string;
  cast?: string[];
  duration?: number; // in minutes
  releaseDate?: string;
  language?: string;
  country?: string;
  budget?: number;
  revenue?: number;
  imdbId?: string;
  tmdbId?: string;
};

export function useMovieDetails(movieId: string) {
  const [movie, setMovie] = React.useState<MovieDetails | null>(null);
  const [isLoading, setIsLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  const toggleWatched = React.useCallback(() => {
    if (!movie) return;
    setMovie(prev => prev ? { ...prev, watched: !prev.watched } : null);
  }, [movie]);

  return {
    movie,
    isLoading,
    error,
    toggleWatched,
  };
}
