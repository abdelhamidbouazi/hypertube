"use client";

import React from "react";
import type { Movie } from "./MoviesCard";

const MOCK_MOVIES: Movie[] = [
  {
    id: "mock-1",
    title: "The Dark Knight",
    year: 2008,
    rating: 9.0,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Action", "Crime", "Drama"],
    watched: false,
  },
  {
    id: "mock-2", 
    title: "Inception",
    year: 2010,
    rating: 8.8,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Action", "Sci-Fi", "Thriller"],
    watched: true,
  },
  {
    id: "mock-3",
    title: "Interstellar",
    year: 2014,
    rating: 8.6,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Adventure", "Drama", "Sci-Fi"],
    watched: false,
  },
  {
    id: "mock-4",
    title: "The Matrix",
    year: 1999,
    rating: 8.7,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Action", "Sci-Fi"],
    watched: true,
  },
  {
    id: "mock-5",
    title: "Pulp Fiction",
    year: 1994,
    rating: 8.9,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Crime", "Drama"],
    watched: false,
  },
  {
    id: "mock-6",
    title: "The Godfather",
    year: 1972,
    rating: 9.2,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Crime", "Drama"],
    watched: true,
  },
  {
    id: "mock-7",
    title: "Forrest Gump",
    year: 1994,
    rating: 8.8,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Drama", "Romance"],
    watched: false,
  },
  {
    id: "mock-8",
    title: "The Shawshank Redemption",
    year: 1994,
    rating: 9.3,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Drama"],
    watched: true,
  },
  {
    id: "mock-9",
    title: "Avatar",
    year: 2009,
    rating: 7.8,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Action", "Adventure", "Fantasy"],
    watched: false,
  },
  {
    id: "mock-10",
    title: "Titanic",
    year: 1997,
    rating: 7.9,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Drama", "Romance"],
    watched: true,
  },
];

type MoviesResponse = {
  items: Movie[];
  nextCursor?: string | null;
};

export function useInfiniteMovies() {
  const [query, setQuery] = React.useState("");
  const [sort, setSort] = React.useState<"name" | "year" | "rating">("name");
  const [selectedGenres, setSelectedGenres] = React.useState<string[]>([]);
  const [minRating, setMinRating] = React.useState<number>(0);
  const [yearRange, setYearRange] = React.useState<[number, number]>([2000, new Date().getFullYear()]);

  const [movies, setMovies] = React.useState<Movie[]>([]);
  const [cursor, setCursor] = React.useState<string | null>(null);
  const [hasMore, setHasMore] = React.useState(true);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isFetchingMore, setIsFetchingMore] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [useMockData, setUseMockData] = React.useState(false);

  const params = React.useMemo(() => {
    const p = new URLSearchParams();
    if (query.trim()) p.set("q", query.trim());
    p.set("sort", sort);
    if (selectedGenres.length) p.set("genres", selectedGenres.join(","));
    p.set("minRating", String(minRating));
    p.set("yearFrom", String(yearRange[0]));
    p.set("yearTo", String(yearRange[1]));
    return p;
  }, [query, sort, selectedGenres, minRating, yearRange]);

  const fetchPage = React.useCallback(async (reset = false) => {
    try {
      if (reset) {
        setIsLoading(true);
        setCursor(null);
        setHasMore(true);
        setError(null);
      } else {
        setIsFetchingMore(true);
      }

      if (useMockData) {
        const filteredMovies = MOCK_MOVIES.filter(movie => {
          if (query.trim() && !movie.title.toLowerCase().includes(query.toLowerCase())) {
            return false;
          }
          if (selectedGenres.length > 0 && !selectedGenres.some(genre => movie.genres?.includes(genre))) {
            return false;
          }
          if (movie.rating && movie.rating < minRating) {
            return false;
          }
          if (movie.year && (movie.year < yearRange[0] || movie.year > yearRange[1])) {
            return false;
          }
          return true;
        });

        const sortedMovies = [...filteredMovies].sort((a, b) => {
          switch (sort) {
            case "year":
              return (b.year || 0) - (a.year || 0);
            case "rating":
              return (b.rating || 0) - (a.rating || 0);
            case "name":
            default:
              return a.title.localeCompare(b.title);
          }
        });

        setMovies(reset ? sortedMovies : [...movies, ...sortedMovies]);
        setHasMore(false);
        return;
      }

      const qp = new URLSearchParams(params);
      if (!reset && cursor) qp.set("cursor", cursor);

      const res = await fetch(`/api/movies?${qp.toString()}`, { cache: "no-store" });
      if (!res.ok) throw new Error("Failed to fetch movies");
      const data: MoviesResponse = await res.json();

      setMovies((prev) => (reset ? data.items : [...prev, ...data.items]));
      setCursor(data.nextCursor ?? null);
      setHasMore(Boolean(data.nextCursor));
    } catch (e) {
      console.error(e);
      setError("Failed to fetch movies. Please try again later.");
    } finally {
      setIsLoading(false);
      setIsFetchingMore(false);
    }
  }, [params, cursor, useMockData, query, selectedGenres, minRating, yearRange, sort, movies]);

  // Initial + whenever filters change -> reset and fetch
  React.useEffect(() => {
    const id = setTimeout(() => {
      fetchPage(true);
    }, 250); // small debounce for search typing
    return () => clearTimeout(id);
  }, [params, fetchPage]);

  const fetchNext = React.useCallback(() => {
    if (!hasMore || isFetchingMore) return;
    fetchPage(false);
  }, [hasMore, isFetchingMore, fetchPage]);

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
    // fetch will auto-trigger via effect
  }, []);

  return {
    // filters
    query, setQuery,
    sort, setSort,
    selectedGenres, toggleGenre,
    minRating, setMinRating,
    yearRange, setYearRange,

    // data
    movies, isLoading, fetchNext, hasMore, isFetchingMore, error,

    // helpers
    resetAndRefetch, useMockData, setUseMockData
  };
}
