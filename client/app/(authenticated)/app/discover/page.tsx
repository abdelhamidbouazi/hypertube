"use client";

import React, { useMemo, useRef, useEffect, useState } from "react";
import { Input } from "@heroui/input";
import { Select, SelectItem } from "@heroui/select";
import { Button } from "@heroui/button";
import { Chip } from "@heroui/chip";
import { Card } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Divider } from "@heroui/divider";
import { Slider } from "@heroui/slider";
import { Tooltip } from "@heroui/tooltip";
import { SearchIcon, X, SlidersHorizontal, RefreshCcw, ListFilterPlus } from "lucide-react";
import MovieCard from "@/components/movies/MoviesCard";
import { useInfiniteMovies } from "@/components/movies/useInfiniteMovies";

type SortKey = "name" | "year" | "rating";
const SORT_OPTIONS: { key: SortKey; label: string }[] = [
  { key: "name", label: "Name (A→Z)" },
  { key: "year", label: "Year (newest)" },
  { key: "rating", label: "IMDb (high→low)" },
];

const GENRES = [
  "Action","Adventure","Animation","Comedy","Crime","Drama",
  "Fantasy","Horror","Mystery","Romance","Sci-Fi","Thriller",
];

export default function DiscoverPage() {
  const {
    query, setQuery,
    sort, setSort,
    selectedGenres, toggleGenre,
    minRating, setMinRating,
    yearRange, setYearRange,
    movies, isLoading, fetchNext, hasMore, isFetchingMore, resetAndRefetch
  } = useInfiniteMovies();
  const [filters, setFilters] = useState(false)

  const sentinelRef = useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    if (!sentinelRef.current) return;
    const el = sentinelRef.current;
    const io = new IntersectionObserver((entries) => {
      const first = entries[0];
      if (first.isIntersecting && hasMore && !isFetchingMore) fetchNext();
    });
    io.observe(el);
    return () => io.unobserve(el);
  }, [hasMore, isFetchingMore, fetchNext]);

  const genreChips = useMemo(
    () =>
      GENRES.map((g) => {
        const active = selectedGenres.includes(g);
        return (
          <Chip
            key={g}
            variant={active ? "solid" : "flat"}
            color={active ? "primary" : "default"}
            onClick={() => toggleGenre(g)}
            className="cursor-pointer transition-transform hover:scale-[1.02]"
            radius="sm"
            size="sm"
          >
            {g}
          </Chip>
        );
      }),
    [selectedGenres, toggleGenre]
  );

  const hasActiveFilters =
    query.trim().length > 0 ||
    selectedGenres.length > 0 ||
    minRating > 0 ||
    yearRange[0] !== 1950 ||
    yearRange[1] !== new Date().getFullYear() ||
    sort !== "name";

    
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
            <p className="text-small text-foreground-500">
              Search, filter, and scroll to explore. Watched titles are marked.
            </p>
          </div>
          <div className="flex gap-2">
            {filters ? <>
            <Tooltip content="Reset filters">
              <Button variant="flat" startContent={<RefreshCcw size={14} />} onPress={() => resetAndRefetch()}>
                Reset
              </Button>
            </Tooltip>
            <Tooltip content="Reset filters">
              <Button variant="flat" color="secondary" startContent={<ListFilterPlus size={14} />} onPress={() => {setFilters(false)}}>
                Close Filters
              </Button>
            </Tooltip> 
            </>: <Tooltip content="Reset filters">
              <Button variant="flat" color="secondary" startContent={<ListFilterPlus size={14} />} onPress={() => {setFilters(true)}}>
                Filters
              </Button>
            </Tooltip> }
            
          </div>
        </div>

        {/* Filters Bar */}
       { filters && <Card className="mt-5 border border-default-200 bg-content1/80 p-4 shadow-sm backdrop-blur-md">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-12">
            {/* 1. Search */}
            <div className="md:col-span-5">
              <Input
                aria-label="Search movies"
                placeholder="Search by title, cast, etc."
                startContent={<SearchIcon size={16} />}
                value={query}
                onValueChange={setQuery}
                variant="flat"
              />
            </div>

            {/* 2. Sort */}
            <div className="md:col-span-3">
              <Select
                aria-label="Sort by"
                label="Sort"
                labelPlacement="outside-left"
                selectedKeys={[sort]}
                onSelectionChange={(keys) => {
                  const k = Array.from(keys)[0] as SortKey;
                  if (k) setSort(k);
                }}
                variant="flat"
                startContent={<SlidersHorizontal size={14} />}
              >
                {SORT_OPTIONS.map(({ key, label }) => (
                  <SelectItem key={key} value={key}>
                    {label}
                  </SelectItem>
                ))}
              </Select>
            </div>

            {/* 3. Min IMDb */}
            <div className="md:col-span-4">
              <div className="flex items-center justify-between">
                <label className="text-small font-medium">Min IMDb</label>
                <span className="text-tiny text-foreground-500">{minRating.toFixed(1)}</span>
              </div>
              <Slider
                aria-label="Minimum IMDb rating"
                minValue={0}
                maxValue={10}
                step={0.1}
                value={minRating}
                onChange={(v) => {
                  const val = Array.isArray(v) ? v[0] : v;
                  setMinRating(Number(val));
                }}
                className="mt-2"
              />
            </div>

            {/* 4. Year Range */}
            <div className="md:col-span-12">
              <div className="flex items-center justify-between">
                <label className="text-small font-medium">Year Range</label>
                <span className="text-tiny text-foreground-500">
                  {yearRange[0]} – {yearRange[1]}
                </span>
              </div>
              <Slider
                aria-label="Year range"
                minValue={1950}
                maxValue={new Date().getFullYear()}
                step={1}
                value={yearRange}
                onChange={(v) => {
                  const arr = Array.isArray(v) ? (v as number[]) : [Number(v), Number(v)];
                  if (arr.length === 2) setYearRange([arr[0], arr[1]]);
                }}
                className="mt-2"
              />
            </div>

            {/* 5. Genres */}
            <div className="md:col-span-12">
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-small font-medium">Genres:</span>
                {genreChips}
              </div>
            </div>
          </div>

          {/* Active Filters Summary */}
          {hasActiveFilters && (
            <div className="mt-4 flex flex-wrap items-center gap-2">
              {!!query.trim() && (
                <Chip
                  variant="flat"
                  color="primary"
                  endContent={<X size={12} className="cursor-pointer" onClick={() => setQuery("")} />}
                  className="text-tiny"
                  radius="sm"
                >
                  “{query.trim()}”
                </Chip>
              )}
              {selectedGenres.map((g) => (
                <Chip
                  key={g}
                  variant="flat"
                  color="secondary"
                  endContent={<X size={12} className="cursor-pointer" onClick={() => toggleGenre(g)} />}
                  className="text-tiny"
                  radius="sm"
                >
                  {g}
                </Chip>
              ))}
              {minRating > 0 && (
                <Chip
                  variant="flat"
                  color="success"
                  endContent={<X size={12} className="cursor-pointer" onClick={() => setMinRating(0)} />}
                  className="text-tiny"
                  radius="sm"
                >
                  IMDb ≥ {minRating.toFixed(1)}
                </Chip>
              )}
              {(yearRange[0] !== 1950 || yearRange[1] !== new Date().getFullYear()) && (
                <Chip
                  variant="flat"
                  color="warning"
                  endContent={
                    <X
                      size={12}
                      className="cursor-pointer"
                      onClick={() => setYearRange([1950, new Date().getFullYear()])}
                    />
                  }
                  className="text-tiny"
                  radius="sm"
                >
                  {yearRange[0]}–{yearRange[1]}
                </Chip>
              )}
              {sort !== "name" && (
                <Chip
                  variant="flat"
                  color="default"
                  endContent={<X size={12} className="cursor-pointer" onClick={() => setSort("name")} />}
                  className="text-tiny"
                  radius="sm"
                >
                  Sort: {SORT_OPTIONS.find(s => s.key === sort)?.label}
                </Chip>
              )}
              <div className="ml-auto">
                <Tooltip content="Reset all">
                  <Button size="sm" variant="light" onPress={() => resetAndRefetch()}>
                    Reset
                  </Button>
                </Tooltip>
              </div>
            </div>
          )}
        </Card>}
      </div>

      <Divider className="my-6" />

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
              <p className="text-small text-foreground-500">Try adjusting your filters.</p>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
              {movies.map((m) => (
                <MovieCard key={m.id} movie={m} />
              ))}
            </div>
          )}

          <div ref={sentinelRef} className="h-12 flex items-center justify-center">
            {isFetchingMore ? (
              <Spinner size="sm" label="Loading more…" />
            ) : hasMore ? (
              <span className="text-tiny text-foreground-500">Scroll to load more</span>
            ) : movies.length > 0 ? (
              <span className="text-tiny text-foreground-500">You’ve reached the end</span>
            ) : null}
          </div>
        </>
      )}
    </div>
  );
}
