"use client";

import React, { useMemo } from "react";
import { Card } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Alert } from "@heroui/alert";
import MovieCard, { Movie } from "@/components/movies/MoviesCard";
import { useWatchHistory, useMovies } from "@/lib/hooks";
import { getErrorMessage } from "@/lib/error-utils";

export default function History() {
  const { watchHistory, isLoading: historyLoading, error: historyError } = useWatchHistory();
  const { movies: allMovies, isLoading: moviesLoading, error: moviesError } = useMovies();

  const historyMovieIds = useMemo(() => {
    if (!watchHistory || watchHistory.length === 0) return [];
    return watchHistory.map((item: any) => item.movie_id);
  }, [watchHistory]);

  const historyMovies = useMemo(() => {
    if (!watchHistory || watchHistory.length === 0) return [];
    
    if (!allMovies || allMovies.length === 0) {
      return watchHistory.map((item: any) => ({
        id: item.movie_id,
        title: item.movie_title,
        poster_path: item.poster_path,
        release_date: undefined,
        overview: undefined,
        original_language: undefined,
        watched: true,
        vote_average: undefined,
      } as Movie));
    }
    
    const moviesMap = new Map<number, Movie>();
    allMovies.forEach((movie: Movie) => {
      moviesMap.set(movie.id, movie);
    });
    
    return watchHistory
      .map((item: any): Movie | null => {
        const fullMovie = moviesMap.get(item.movie_id);
        if (fullMovie) {
          return {
            ...fullMovie,
            watched: true,
          };
        }
        return {
          id: item.movie_id,
          title: item.movie_title,
          poster_path: item.poster_path,
          release_date: undefined,
          overview: undefined,
          original_language: undefined,
          watched: true,
          vote_average: undefined,
        } as Movie;
      })
      .filter((movie: Movie | null): movie is Movie => movie !== null);
  }, [allMovies, watchHistory]);

  const isLoading = historyLoading || moviesLoading;
  const error = historyError || moviesError;

  return (
    <div className="w-full">
      {/* Header */}
      <div className="relative overflow-hidden rounded-3xl bg-content1/70 p-6 shadow-sm backdrop-blur-md mb-6">
        <div className="pointer-events-none absolute inset-0 -z-10 opacity-40 blur-3xl">
          <div className="absolute -top-24 right-0 h-48 w-48 rounded-full bg-gradient-to-br from-indigo-500 to-pink-500 dark:from-indigo-400 dark:to-pink-400" />
          <div className="absolute -bottom-24 left-0 h-48 w-48 rounded-full bg-gradient-to-br from-cyan-500 to-sky-500" />
        </div>
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight">
            <span className="bg-gradient-to-r from-indigo-600 to-pink-600 bg-clip-text text-transparent dark:from-indigo-400 dark:to-pink-400">
              Watch History
            </span>
          </h1>
          <p className="text-small text-foreground-500 mt-2">
            Movies you've watched. {historyMovieIds.length > 0 && `${historyMovieIds.length} movie${historyMovieIds.length !== 1 ? 's' : ''} in your history.`}
          </p>
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="mb-6">
          <Alert
            color="danger"
            variant="flat"
            title="Failed to fetch watch history"
            description={getErrorMessage(error)}
          />
        </div>
      )}

      {/* Loading State */}
      {isLoading ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
          {Array.from({ length: 10 }).map((_, i) => (
            <Card key={i} className="h-64 animate-pulse bg-content2" />
          ))}
        </div>
      ) : (
        <>
          {/* Empty State */}
          {historyMovies.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-2 py-16">
              <p className="text-medium">No movies in your watch history</p>
              <p className="text-small text-foreground-500">
                Start watching movies to see them appear here.
              </p>
            </div>
          ) : (
            /* Movies Grid */
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
              {historyMovies.map((movie: Movie) => (
                <MovieCard key={movie.id} movie={movie} />
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}