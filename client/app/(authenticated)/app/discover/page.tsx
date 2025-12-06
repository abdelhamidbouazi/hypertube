"use client";

import React, { useMemo, useState } from "react";
import { Input } from "@heroui/input";
import { Select, SelectItem } from "@heroui/select";
import { Button } from "@heroui/button";
import { Chip } from "@heroui/chip";
import { Card } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Divider } from "@heroui/divider";
import { Slider } from "@heroui/slider";
import { Tooltip } from "@heroui/tooltip";
import { Alert } from "@heroui/alert";
import { SearchIcon, X, SlidersHorizontal, ListFilterPlus } from "lucide-react";
import MovieCard, { Movie } from "@/components/movies/MoviesCard";
import HeroSection from "@/components/HeroSection";
import { useMovies } from "@/lib/hooks";
import { useFilterStore } from "@/lib/store";
import { getErrorMessage } from "@/lib/error-utils";

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
  const {
    filters,
    setQuery,
    setSort,
    toggleGenre,
    setMinRating,
    setYearRange,
    resetFilters,
  } = useFilterStore();

  const {
    movies,
    isLoading,
    error,
    isLoadingMore,
    isReachingEnd,
    setSize,
    size,
  } = useMovies(filters);
  const [showFilters, setShowFilters] = useState(false);

  const genreChips = useMemo(
    () =>
      GENRES.map((g) => {
        const active = filters.selectedGenres.includes(g);
        return (
          <Chip
            className="cursor-pointer transition-transform hover:scale-[1.02]"
            key={g}
            onClick={() => toggleGenre(g)}
            radius="sm"
            size="sm"
            variant={active ? "solid" : "flat"}
            color={active ? "primary" : "default"}
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
    filters.yearRange[0] !== 2000 ||
    filters.yearRange[1] !== new Date().getFullYear() ||
    filters.sort !== "name";

  const isSearchActive = filters.query.trim().length > 0;

  // Infinite scroll observer
  const observerTarget = React.useRef(null);

  React.useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && !isLoadingMore && !isReachingEnd) {
          setSize(size + 1);
        }
      },
      { threshold: 1.0 }
    );

    if (observerTarget.current) {
      observer.observe(observerTarget.current);
    }

    return () => {
      if (observerTarget.current) {
        observer.unobserve(observerTarget.current);
      }
    };
  }, [observerTarget, isLoadingMore, isReachingEnd, setSize, size]);

  return (
    <div className="w-full">
      {/* Hero section (top picks) – FIRST */}
      {movies.length > 0 && !filters.query && (
        <div className="mb-8">
          <HeroSection
            slides={movies
              .filter((m: any) => m && m.id)
              .slice(0, Math.min(8, movies.length))}
          />
        </div>
      )}

      {/* Header and Filters */}
      <div className="relative overflow-hidden rounded-3xl  bg-content1/70 p-6 shadow-sm backdrop-blur-md ">
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
            <p className="text-small text-foreground-500">
              Search, filter, and scroll to explore. Watched titles are marked.
            </p>
          </div>
          <div className="flex gap-2">
            {showFilters ? (
              <>
                <Tooltip content="Close filters">
                  <Button
                    color="secondary"
                    startContent={<ListFilterPlus size={14} />}
                    variant="flat"
                    onPress={() => {
                      setShowFilters(false);
                    }}
                  >
                    Close Filters
                  </Button>
                </Tooltip>
              </>
            ) : (
              <Tooltip content="Open filters">
                <Button
                  color="secondary"
                  startContent={<ListFilterPlus size={14} />}
                  variant="flat"
                  onPress={() => {
                    setShowFilters(true);
                  }}
                >
                  Filters
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
                  description={
                    isSearchActive ? "Filters are disabled while searching" : ""
                  }
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
                  isDisabled={isSearchActive}
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
                  <span
                    className={`text-small font-medium ${isSearchActive ? "text-default-400" : ""}`}
                  >
                    Min IMDb
                  </span>
                  <span className="text-tiny text-foreground-500">
                    {filters.minRating.toFixed(1)}
                  </span>
                </div>
                <Slider
                  aria-label="Minimum IMDb rating"
                  className="mt-2"
                  maxValue={10}
                  minValue={0}
                  step={0.1}
                  value={filters.minRating}
                  isDisabled={isSearchActive}
                  onChange={(v) => {
                    const val = Array.isArray(v) ? v[0] : v;
                    setMinRating(Number(val));
                  }}
                />
              </div>

              {/* 4. Year Range */}
              <div className="md:col-span-12">
                <div className="flex items-center justify-between">
                  <span
                    className={`text-small font-medium ${isSearchActive ? "text-default-400" : ""}`}
                  >
                    Year Range
                  </span>
                  <span className="text-tiny text-foreground-500">
                    {filters.yearRange[0]} – {filters.yearRange[1]}
                  </span>
                </div>
                <Slider
                  aria-label="Year range"
                  className="mt-2"
                  maxValue={new Date().getFullYear()}
                  minValue={1950}
                  step={1}
                  value={filters.yearRange}
                  isDisabled={isSearchActive}
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
                <div className="flex flex-wrap items-center gap-2">
                  <span
                    className={`text-small font-medium ${isSearchActive ? "text-default-400" : ""}`}
                  >
                    Genres:
                  </span>
                  {GENRES.map((g) => {
                    const active = filters.selectedGenres.includes(g);
                    return (
                      <Chip
                        key={g}
                        className={`transition-transform ${!isSearchActive ? "cursor-pointer hover:scale-[1.02]" : "opacity-50"}`}
                        onClick={() => !isSearchActive && toggleGenre(g)}
                        radius="sm"
                        size="sm"
                        variant={active ? "solid" : "flat"}
                        color={active ? "primary" : "default"}
                      >
                        {g}
                      </Chip>
                    );
                  })}
                </div>
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
                        size={12}
                        className="cursor-pointer"
                        onClick={() => setQuery("")}
                      />
                    }
                    radius="sm"
                    variant="flat"
                  >
                    "{filters.query.trim()}"
                  </Chip>
                )}
                {filters.selectedGenres.map((g) => (
                  <Chip
                    key={g}
                    className="text-tiny"
                    color="secondary"
                    endContent={
                      <X
                        size={12}
                        className="cursor-pointer"
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
                        size={12}
                        className="cursor-pointer"
                        onClick={() => setMinRating(0)}
                      />
                    }
                    radius="sm"
                    variant="flat"
                  >
                    IMDb ≥ {filters.minRating.toFixed(1)}
                  </Chip>
                )}
                {(filters.yearRange[0] !== 2000 ||
                  filters.yearRange[1] !== new Date().getFullYear()) && (
                  <Chip
                    className="text-tiny"
                    color="warning"
                    endContent={
                      <X
                        size={12}
                        className="cursor-pointer"
                        onClick={() =>
                          setYearRange([2000, new Date().getFullYear()])
                        }
                      />
                    }
                    radius="sm"
                    variant="flat"
                  >
                    {filters.yearRange[0]}–{filters.yearRange[1]}
                  </Chip>
                )}
                {filters.sort !== "name" && (
                  <Chip
                    className="text-tiny"
                    color="default"
                    endContent={
                      <X
                        size={12}
                        className="cursor-pointer"
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

      <Divider className="my-6 bg-transparent" />

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

      {isLoading ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
          {Array.from({ length: 10 }).map((_, i) => (
            <Card key={i} className="h-64 animate-pulse bg-content2" />
          ))}
        </div>
      ) : (
        <>
          {movies.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-2 py-16">
              <p className="text-medium">No movies found</p>
              <p className="text-small text-foreground-500">
                Try adjusting your filters.
              </p>
            </div>
          ) : (
            <>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
                {movies
                  .filter((m: any) => m && m.id)
                  .map((m: Movie) => (
                    <MovieCard key={m.id} movie={m} />
                  ))}
              </div>

              {/* Infinite Scroll Sentinel */}
              <div
                ref={observerTarget}
                className="flex justify-center py-8 w-full"
              >
                {isLoadingMore && <Spinner color="primary" />}
              </div>
            </>
          )}
        </>
      )}
    </div>
  );
}
