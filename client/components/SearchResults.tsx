"use client";

import type { SearchResult } from "@/types";

import React from "react";
import { Card, CardBody } from "@heroui/card";
import { Avatar } from "@heroui/avatar";
import { Divider } from "@heroui/divider";
import { Spinner } from "@heroui/spinner";
import { Button } from "@heroui/button";
import Link from "next/link";
import { CalendarDays, Film } from "lucide-react";

interface SearchResultsProps {
  results: SearchResult[];
  isLoading: boolean;
  error: string | null;
  query: string;
  onClose: () => void;
  maxResults?: number;
}

interface SearchResultItemProps {
  movie: SearchResult;
  onClose: () => void;
}

function SearchResultItem({ movie, onClose }: SearchResultItemProps) {
  const year = movie.release_date
    ? new Date(movie.release_date).getFullYear()
    : null;
  const posterUrl = movie.poster_path
    ? `https://image.tmdb.org/t/p/w92${movie.poster_path}`
    : "/placeholder-poster.jpg";

  return (
    <Link
      className="block w-full"
      href={`/app/movie/${movie.id}`}
      onClick={onClose}
    >
      <div className="flex items-start gap-3 p-3 hover:bg-default-100 transition-colors rounded-lg cursor-pointer">
        <Avatar
          className="flex-shrink-0"
          radius="sm"
          size="lg"
          src={posterUrl}
        />
        <div className="flex-1 min-w-0">
          <p className="font-medium text-sm text-foreground truncate">
            {movie.title}
          </p>
          <div className="flex items-center gap-2 mt-1">
            {year && (
              <div className="flex items-center gap-1 text-xs text-default-600">
                <CalendarDays className="h-3 w-3" />
                <span>{year}</span>
              </div>
            )}
            <div className="flex items-center gap-1 text-xs text-default-600">
              <Film className="h-3 w-3" />
              <span>Movie</span>
            </div>
          </div>
          {movie.overview && (
            <p className="text-xs text-default-500 mt-1 line-clamp-3">
              {movie.overview}
            </p>
          )}
        </div>
      </div>
    </Link>
  );
}

export function SearchResults({
  results,
  isLoading,
  error,
  query,
  onClose,
  maxResults = 8,
}: SearchResultsProps) {
  const displayResults = results.slice(0, maxResults);
  const hasMoreResults = results.length > maxResults;

  if (!query.trim()) {
    return null;
  }

  return (
    <Card className="absolute top-full left-0 mt-2 z-50 max-h-96 overflow-hidden w-[90vw] max-w-96 sm:w-[500px] sm:max-w-[500px] lg:w-[600px] lg:max-w-[600px] shadow-lg border border-default-200">
      <CardBody className="p-0">
        {isLoading && (
          <div className="flex items-center justify-center p-4">
            <Spinner size="sm" />
            <span className="ml-2 text-sm text-default-600">Searching...</span>
          </div>
        )}

        {error && (
          <div className="p-4 text-center">
            <p className="text-sm text-danger">{error}</p>
            <Button
              className="mt-2"
              size="sm"
              variant="light"
              onPress={onClose}
            >
              Close
            </Button>
          </div>
        )}

        {!isLoading &&
          !error &&
          displayResults.length === 0 &&
          query.trim() && (
            <div className="p-4 text-center">
              <p className="text-sm text-default-600">
                No movies found for &ldquo;{query}&rdquo;
              </p>
              <p className="text-xs text-default-500 mt-1">
                Try adjusting your search terms
              </p>
            </div>
          )}

        {!isLoading && !error && displayResults.length > 0 && (
          <div className="max-h-80 overflow-y-auto">
            <div className="p-2">
              <p className="text-xs text-default-600 font-medium px-3 py-2">
                Found {results.length} result{results.length !== 1 ? "s" : ""}{" "}
                for &ldquo;{query}&rdquo;
              </p>
            </div>
            <Divider />

            {displayResults.map((movie, index) => (
              <React.Fragment key={movie.id}>
                <SearchResultItem movie={movie} onClose={onClose} />
                {index < displayResults.length - 1 && <Divider />}
              </React.Fragment>
            ))}

            {hasMoreResults && (
              <>
                <Divider />
                <div className="p-3 text-center bg-default-50">
                  <Link href={`/app/search?q=${encodeURIComponent(query)}`}>
                    <Button
                      className="w-full"
                      size="sm"
                      variant="light"
                      onPress={onClose}
                    >
                      View all {results.length} results
                    </Button>
                  </Link>
                </div>
              </>
            )}
          </div>
        )}
      </CardBody>
    </Card>
  );
}
