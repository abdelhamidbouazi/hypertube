"use client";

import React, { useMemo, useState, useEffect } from "react";
import { Input } from "@heroui/input";
import { Select, SelectItem } from "@heroui/select";
import { Button } from "@heroui/button";
import { Chip } from "@heroui/chip";
import { Card } from "@heroui/card";
import { Divider } from "@heroui/divider";
import { Slider } from "@heroui/slider";
import { Tooltip } from "@heroui/tooltip";
import { Alert } from "@heroui/alert";
import {
  SearchIcon,
  X,
  SlidersHorizontal,
  Filter,
  FilterX,
} from "lucide-react";

import MovieCard, { Movie } from "@/components/movies/MoviesCard";
import { useInfiniteMovies } from "@/lib/hooks";
import { useFilterStore } from "@/lib/store";

type SortKey = "name" | "year" | "rating";
const SORT_OPTIONS: { key: SortKey; label: string }[] = [
  { key: "name", label: "Name (A→Z)" },
  { key: "year", label: "Year (newest)" },
  { key: "rating", label: "IMDb (high→low)" },
];

const GENRES = [
  "Action",
  "Adventure",
  "Animation",
  "Comedy",
  "Crime",
  "Drama",
  "Fantasy",
  "Horror",
  "Mystery",
  "Romance",
  "Sci-Fi",
  "Thriller",
];

export default function DiscoverPage() {
  const { movies, isLoading, isLoadingMore, error, loadMovies } =
    useInfiniteMovies();
  const {
    filters,
    currentPage,
    hasMore,
    setQuery,
    setSort,
    toggleGenre,
    setMinRating,
    setYearRange,
    setCurrentPage,
    setHasMore,
    resetFilters,
  } = useFilterStore();
  const [showFilters, setShowFilters] = useState(false);

  useEffect(() => {
    const initialLoad = async () => {
      const { hasMore: moreAvailable } = await loadMovies(1, false);

      setHasMore(moreAvailable);
      setCurrentPage(1);
    };

    initialLoad();
  }, []);

  const handleLoadMore = async () => {
    if (!isLoadingMore && hasMore) {
      const nextPage = currentPage + 1;
      const { hasMore: moreAvailable } = await loadMovies(nextPage, true);

      setHasMore(moreAvailable);
      setCurrentPage(nextPage);
    }
  };

  const genreChips = useMemo(
    () =>
      GENRES.map((g) => {
        const active = filters.selectedGenres.includes(g);

        return (
          <Chip
            key={g}
            aria-pressed={active}
            className="cursor-pointer transition-transform hover:scale-[1.02]"
            color={active ? "primary" : "default"}
            radius="sm"
            role="button"
            size="sm"
            variant={active ? "solid" : "flat"}
            onClick={() => toggleGenre(g)}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                toggleGenre(g);
              }
            }}
          >
            {g}
          </Chip>
        );
      }),
    [filters.selectedGenres, toggleGenre]
  );

  const hasActiveFilters =
    filters.query.trim().length > 0 ||
    filters.selectedGenres.length > 0 ||
    filters.minRating > 0 ||
    filters.yearRange[0] !== 1950 ||
    filters.yearRange[1] !== new Date().getFullYear() ||
    filters.sort !== "name";

  return (
    <div className="mx-auto w-full">
      <div className="relative overflow-hidden rounded-3xl border border-default-200 bg-content1/70 p-6 shadow-sm backdrop-blur-md dark:border-default-100">
        <div className="pointer-events-none absolute inset-0 -z-10 opacity-40 blur-3xl">
          <div className="absolute -top-24 right-0 h-48 w-48 rounded-full bg-gradient-to-br from-indigo-500 to-pink-500 dark:from-indigo-400 dark:to-pink-400" />
          <div className="absolute -bottom-24 left-0 h-48 w-48 rounded-full bg-gradient-to-br from-cyan-500 to-sky-500" />
        </div>
        <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
          <div>
            <h1 className="text-3xl font-extrabold tracking-tight">
              <span className="bg-gradient-to-r from-indigo-600 to-pink-600 bg-clip-text text-transparent dark:from-indigo-400 dark:to-pink-400">
                Discover movies
              </span>
            </h1>
            <p className="text-default-600 dark:text-default-400">
              Search, filter, and discover movies. Explore our vast collection.
            </p>
          </div>
          <div className="flex gap-2">
            {showFilters ? (
              <Tooltip content="Hide filters">
                <Button
                  aria-label="Hide filters"
                  color="secondary"
                  startContent={<FilterX size={14} />}
                  variant="flat"
                  onPress={() => {
                    setShowFilters(false);
                  }}
                >
                  Hide Filters
                </Button>
              </Tooltip>
            ) : (
              <Tooltip content="Show filters">
                <Button
                  aria-label="Show filters"
                  color="secondary"
                  startContent={<Filter size={14} />}
                  variant="flat"
                  onPress={() => {
                    setShowFilters(true);
                  }}
                >
                  Show Filters
                  {hasActiveFilters && (
                    <Chip
                      className="ml-1"
                      color="primary"
                      size="sm"
                      variant="dot"
                    />
                  )}
                </Button>
              </Tooltip>
            )}
          </div>
        </div>

        {/* Filters Bar */}
        {showFilters && (
          <Card className="mt-5 border border-default-200 bg-content1/80 p-4 shadow-sm backdrop-blur-md">
            <div className="grid grid-cols-1 gap-4 md:grid-cols-12">
              {/* 1. Search */}
              <div className="md:col-span-5">
                <Input
                  aria-label="Search movies"
                  placeholder="Search by title, cast, etc."
                  startContent={<SearchIcon size={16} />}
                  value={filters.query}
                  variant="flat"
                  onValueChange={setQuery}
                />
              </div>

              {/* 2. Sort */}
              <div className="md:col-span-3">
                <Select
                  aria-label="Sort by"
                  label="Sort"
                  labelPlacement="outside-left"
                  selectedKeys={[filters.sort]}
                  startContent={<SlidersHorizontal size={14} />}
                  variant="flat"
                  onSelectionChange={(keys) => {
                    const k = Array.from(keys)[0] as SortKey;

                    if (k) setSort(k);
                  }}
                >
                  {SORT_OPTIONS.map(({ key, label }) => (
                    <SelectItem key={key}>{label}</SelectItem>
                  ))}
                </Select>
              </div>

              {/* 3. Min IMDb */}
              <div className="md:col-span-4">
                <div className="flex items-center justify-between">
                  <label
                    className="text-small font-medium"
                    htmlFor="min-rating-slider"
                  >
                    Min IMDb Rating
                  </label>
                  <span className="text-tiny text-foreground-500">
                    {filters.minRating.toFixed(1)}
                  </span>
                </div>
                <Slider
                  aria-label="Minimum IMDb rating filter"
                  className="mt-2"
                  id="min-rating-slider"
                  maxValue={10}
                  minValue={0}
                  step={0.1}
                  value={filters.minRating}
                  onChange={(v) => {
                    const val = Array.isArray(v) ? v[0] : v;

                    setMinRating(Number(val));
                  }}
                />
              </div>

              {/* 4. Year Range */}
              <div className="md:col-span-12">
                <div className="flex items-center justify-between">
                  <label
                    className="text-small font-medium"
                    htmlFor="year-range-slider"
                  >
                    Release Year Range
                  </label>
                  <span className="text-tiny text-foreground-500">
                    {filters.yearRange[0]} – {filters.yearRange[1]}
                  </span>
                </div>
                <Slider
                  aria-label="Movie release year range filter"
                  className="mt-2"
                  id="year-range-slider"
                  maxValue={new Date().getFullYear()}
                  minValue={1950}
                  step={1}
                  value={filters.yearRange}
                  onChange={(v) => {
                    const arr = Array.isArray(v)
                      ? (v as number[])
                      : [Number(v), Number(v)];

                    if (arr.length === 2) setYearRange([arr[0], arr[1]]);
                  }}
                />
              </div>

              {/* 5. Genres */}
              <div className="md:col-span-12">
                <fieldset className="space-y-2">
                  <legend className="text-small font-medium">
                    Filter by Genres
                  </legend>
                  <div
                    aria-label="Movie genre filters"
                    className="flex flex-wrap gap-2"
                    role="group"
                  >
                    {genreChips}
                  </div>
                </fieldset>
              </div>
            </div>

            {/* Active Filters Summary */}
            {hasActiveFilters && (
              <div className="mt-4 flex flex-wrap items-center gap-2">
                {!!filters.query.trim() && (
                  <Chip
                    className="text-tiny"
                    color="primary"
                    endContent={
                      <X
                        className="cursor-pointer"
                        size={12}
                        onClick={() => setQuery("")}
                      />
                    }
                    radius="sm"
                    variant="flat"
                  >
                    Search: {filters.query.trim()}
                  </Chip>
                )}
                {filters.selectedGenres.map((g) => (
                  <Chip
                    key={g}
                    className="text-tiny"
                    color="secondary"
                    endContent={
                      <X
                        className="cursor-pointer"
                        size={12}
                        onClick={() => toggleGenre(g)}
                      />
                    }
                    radius="sm"
                    variant="flat"
                  >
                    {g}
                  </Chip>
                ))}
                {filters.minRating > 0 && (
                  <Chip
                    className="text-tiny"
                    color="success"
                    endContent={
                      <X
                        className="cursor-pointer"
                        size={12}
                        onClick={() => setMinRating(0)}
                      />
                    }
                    radius="sm"
                    variant="flat"
                  >
                    IMDb ≥ {filters.minRating.toFixed(1)}
                  </Chip>
                )}
                {(filters.yearRange[0] !== 1950 ||
                  filters.yearRange[1] !== new Date().getFullYear()) && (
                  <Chip
                    className="text-tiny"
                    color="warning"
                    endContent={
                      <X
                        className="cursor-pointer"
                        size={12}
                        onClick={() =>
                          setYearRange([1950, new Date().getFullYear()])
                        }
                      />
                    }
                    radius="sm"
                    variant="flat"
                  >
                    Years: {filters.yearRange[0]}–{filters.yearRange[1]}
                  </Chip>
                )}
                {filters.sort !== "name" && (
                  <Chip
                    className="text-tiny"
                    color="default"
                    endContent={
                      <X
                        className="cursor-pointer"
                        size={12}
                        onClick={() => setSort("name")}
                      />
                    }
                    radius="sm"
                    variant="flat"
                  >
                    Sort:{" "}
                    {SORT_OPTIONS.find((s) => s.key === filters.sort)?.label}
                  </Chip>
                )}
                <div className="ml-auto">
                  <Tooltip content="Reset all">
                    <Button
                      color="secondary"
                      size="sm"
                      variant="flat"
                      onPress={() => resetFilters()}
                    >
                      Reset
                    </Button>
                  </Tooltip>
                </div>
              </div>
            )}
          </Card>
        )}
      </div>

      <Divider className="my-6" />

      {/* Error Alert */}
      {error && (
        <div className="mb-6">
          <Alert
            color="danger"
            description={
              typeof error === "string"
                ? error
                : "Something went wrong while loading movies. Please try again."
            }
            title="Unable to load movies"
            variant="bordered"
          />
        </div>
      )}

      {isLoading ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
          {Array.from({ length: 15 }).map((_, i) => (
            <Card key={i} className="h-80 p-0 overflow-hidden">
              <div className="w-full h-60 bg-content2 animate-pulse" />
              <div className="p-3 space-y-2">
                <div className="h-4 bg-content2 animate-pulse rounded" />
                <div className="h-3 bg-content2 animate-pulse rounded w-3/4" />
              </div>
            </Card>
          ))}
        </div>
      ) : (
        <>
          {movies.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-3 py-16">
              <div className="text-6xl">🎬</div>
              <h3 className="text-lg font-semibold">No movies found</h3>
              <p className="text-default-500 text-center max-w-md">
                Try adjusting your search terms or filters to discover more
                movies.
              </p>
              {hasActiveFilters && (
                <Button
                  color="primary"
                  size="sm"
                  variant="flat"
                  onPress={resetFilters}
                >
                  Clear All Filters
                </Button>
              )}
            </div>
          ) : (
            <>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
                {movies.map((m: Movie) => (
                  <MovieCard key={m.id} movie={m} />
                ))}
              </div>

              {hasMore && (
                <div className="flex justify-center mt-8">
                  <Button
                    className="px-8"
                    color="primary"
                    isLoading={isLoadingMore}
                    size="lg"
                    variant="flat"
                    onPress={handleLoadMore}
                  >
                    {isLoadingMore ? "Loading more..." : "Load More Movies"}
                  </Button>
                </div>
              )}
            </>
          )}
        </>
      )}
    </div>
  );
}
