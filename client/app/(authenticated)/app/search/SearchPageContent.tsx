"use client";

import React, { useState, useEffect } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { Input } from "@heroui/input";
import { Button } from "@heroui/button";
import { Spinner } from "@heroui/spinner";
import { Card, CardBody } from "@heroui/card";
import Link from "next/link";
import { Search, ArrowLeft } from "lucide-react";

import { useMovieSearch } from "@/hooks/useMovieSearch";
import { SearchIcon } from "@/components/icons";
import MovieCard from "@/components/movies/MoviesCard";

export default function SearchPageContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const initialQuery = searchParams.get("q") || "";

  const { searchMovies, results, isLoading, error } = useMovieSearch();
  const [query, setQuery] = useState(initialQuery);
  const [hasSearched, setHasSearched] = useState(false);

  useEffect(() => {
    if (initialQuery) {
      searchMovies(initialQuery);
      setHasSearched(true);
    }
  }, [initialQuery, searchMovies]);

  useEffect(() => {
    if (hasSearched && query) {
      const params = new URLSearchParams();

      params.set("q", query);

      const newUrl = `/app/search?${params.toString()}`;

      router.replace(newUrl, { scroll: false });
    }
  }, [query, hasSearched, router]);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      await searchMovies(query);
      setHasSearched(true);
    }
  };

  const handleQueryChange = (value: string) => {
    setQuery(value);
  };

  const handleGoBack = () => {
    router.back();
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button isIconOnly size="sm" variant="light" onPress={handleGoBack}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-2xl font-bold text-foreground">Search Movies</h1>
          <p className="text-sm text-default-600">
            Find your favorite movies and discover new ones
          </p>
        </div>
      </div>

      {/* Search Form */}
      <Card>
        <CardBody className="p-6">
          <form className="space-y-4" onSubmit={handleSearch}>
            <div className="flex gap-4">
              <Input
                className="flex-1"
                classNames={{
                  inputWrapper: "bg-default-100",
                  input: "text-sm",
                }}
                // label="Movie Title"
                placeholder="Enter movie title..."
                startContent={
                  <SearchIcon className="text-base text-default-400 pointer-events-none flex-shrink-0" />
                }
                value={query}
                onValueChange={handleQueryChange}
              />
              <div className="flex items-center justify-center">
                <Button
                  className=""
                  color="primary"
                  isDisabled={!query.trim()}
                  isLoading={isLoading}
                  startContent={
                    !isLoading ? <Search className="h-4 w-4" /> : undefined
                  }
                  type="submit"
                >
                  {isLoading ? "Searching..." : "Search"}
                </Button>
              </div>
            </div>
          </form>
        </CardBody>
      </Card>

      {/* Results */}
      <div className="space-y-4">
        {hasSearched && (
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-foreground">
                Search Results
              </h2>
              {!isLoading && !error && (
                <p className="text-sm text-default-600">
                  Found {results.length} result{results.length !== 1 ? "s" : ""}{" "}
                  {query && <>for &ldquo;{query}&rdquo;</>}
                </p>
              )}
            </div>
          </div>
        )}

        {/* Loading State */}
        {isLoading && (
          <div className="flex items-center justify-center py-12">
            <div className="text-center space-y-3">
              <Spinner size="lg" />
              <p className="text-default-600">Searching for movies...</p>
            </div>
          </div>
        )}

        {/* Error State */}
        {error && (
          <Card className="border-danger-200 bg-danger-50">
            <CardBody className="text-center py-8">
              <p className="text-danger font-medium">Search Error</p>
              <p className="text-danger-600 text-sm mt-1">{error}</p>
              <Button
                className="mt-4"
                color="danger"
                size="sm"
                variant="light"
                onPress={() => searchMovies(query)}
              >
                Try Again
              </Button>
            </CardBody>
          </Card>
        )}

        {/* No Results */}
        {!isLoading &&
          !error &&
          hasSearched &&
          results.length === 0 &&
          query && (
            <Card className="bg-default-50">
              <CardBody className="text-center py-12">
                <div className="space-y-3">
                  <div className="text-default-400">
                    <Search className="h-12 w-12 mx-auto" />
                  </div>
                  <div>
                    <p className="font-medium text-foreground">
                      No movies found
                    </p>
                    <p className="text-sm text-default-600 mt-1">
                      Try adjusting your search terms or removing filters
                    </p>
                  </div>
                  <Button
                    size="sm"
                    variant="light"
                    onPress={() => {
                      setQuery("");
                      setHasSearched(false);
                    }}
                  >
                    Clear Search
                  </Button>
                </div>
              </CardBody>
            </Card>
          )}

        {/* Results Grid */}
        {!isLoading && !error && results.length > 0 && (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
            {results.map((movie: any) => (
              <MovieCard key={movie.id} movie={movie} />
            ))}
          </div>
        )}

        {/* Popular Movies Suggestion */}
        {!hasSearched && (
          <Card className="bg-gradient-to-r from-primary-50 to-secondary-50">
            <CardBody className="text-center py-12">
              <div className="space-y-3">
                <div className="text-primary">
                  <Search className="h-12 w-12 mx-auto" />
                </div>
                <div>
                  <p className="font-medium text-foreground">Start searching</p>
                  <p className="text-sm text-default-600 mt-1">
                    Enter a movie title above to find movies
                  </p>
                </div>
                <Link href="/app/discover">
                  <Button color="primary" size="sm" variant="light">
                    Browse Popular Movies
                  </Button>
                </Link>
              </div>
            </CardBody>
          </Card>
        )}
      </div>
    </div>
  );
}
