"use client";

import React, { useMemo } from "react";
import { Card } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Alert } from "@heroui/alert";
import MovieCard, { Movie } from "@/components/movies/MoviesCard";
import { useMovies } from "@/lib/hooks";
import { useWatchlistStore } from "@/lib/store";
import { getErrorMessage } from "@/lib/error-utils";

export default function WatchLater() {
  const { movies: allMovies, isLoading, error } = useMovies();
  const { watchlistIds } = useWatchlistStore();

  // Filter movies to only show those in watchlist
  const watchlistMovies = useMemo(() => {
    if (!allMovies || watchlistIds.length === 0) return [];
    return allMovies.filter((movie: Movie) => watchlistIds.includes(movie.id));
  }, [allMovies, watchlistIds]);

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
              Watch Later
            </span>
          </h1>
          <p className="text-small text-foreground-500 mt-2">
            Your saved movies to watch later. {watchlistMovies.length > 0 && `${watchlistMovies.length} movie${watchlistMovies.length !== 1 ? 's' : ''} saved.`}
          </p>
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="mb-6">
          <Alert
            color="danger"
            variant="flat"
            title="Failed to fetch movies"
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
          {watchlistMovies.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-2 py-16">
              <p className="text-medium">No movies in your watchlist</p>
              <p className="text-small text-foreground-500">
                Start adding movies to your watchlist by clicking the bookmark icon on any movie.
              </p>
            </div>
          ) : (
            /* Movies Grid */
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
              {watchlistMovies.map((movie: Movie) => (
                <MovieCard key={movie.id} movie={movie} />
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}