"use client";

import React from "react";
import type { Movie } from "./MoviesCard";
import { useMovieDetailsReq } from "@/lib/hooks";

export type MovieDetails = Movie & {
  overview?: string;
  runtime?: number;
  backdrop_path?: string;
  vote_average?: number;
  imdb_id?: string;
  original_language?: string;
  is_available?: boolean;
  stream_url?: string;
  cast?: Array<{ id: number; name: string; character: string; profile_path: string }>;
  director?: Array<{ id: number; name: string }>;
  producer?: Array<{ id: number; name: string }>;
  genres?: Array<{ id: number; name: string }>;
  comments?: Array<{ id: number; movie_id: number; user_id: number; username: string; content: string; created_at: string; updated_at: string }>;
};

export function useMovieDetails(movieId: string) {
  const { movie, isLoading, error } = useMovieDetailsReq(movieId);

  return {
    movie,
    isLoading,
    error,
  };
}
