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

const MOCK_MOVIE_DETAILS: Record<string, MovieDetails> = {
  "mock-1": {
    id: "mock-1",
    title: "The Dark Knight",
    year: 2008,
    rating: 9.0,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Action", "Crime", "Drama"],
    watched: false,
    description: "When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests of his ability to fight injustice.",
    director: "Christopher Nolan",
    cast: ["Christian Bale", "Heath Ledger", "Aaron Eckhart", "Michael Caine", "Maggie Gyllenhaal"],
    duration: 152,
    releaseDate: "2008-07-18",
    language: "English",
    country: "United States",
    budget: 185000000,
    revenue: 1004558444,
    imdbId: "tt0468569",
  },
  "mock-2": {
    id: "mock-2",
    title: "Inception",
    year: 2010,
    rating: 8.8,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Action", "Sci-Fi", "Thriller"],
    watched: true,
    description: "A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.",
    director: "Christopher Nolan",
    cast: ["Leonardo DiCaprio", "Marion Cotillard", "Tom Hardy", "Elliot Page", "Michael Caine"],
    duration: 148,
    releaseDate: "2010-07-16",
    language: "English",
    country: "United States",
    budget: 160000000,
    revenue: 836836967,
    imdbId: "tt1375666",
  },
  "mock-3": {
    id: "mock-3",
    title: "Interstellar",
    year: 2014,
    rating: 8.6,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Adventure", "Drama", "Sci-Fi"],
    watched: false,
    description: "When Earth becomes uninhabitable in the future, a farmer and ex-NASA pilot, Joseph Cooper, is tasked to pilot a spacecraft, along with a team of researchers, to find a new planet for humans.",
    director: "Christopher Nolan",
    cast: ["Matthew McConaughey", "Anne Hathaway", "Jessica Chastain", "Michael Caine", "Matt Damon"],
    duration: 169,
    releaseDate: "2014-11-07",
    language: "English",
    country: "United States",
    budget: 165000000,
    revenue: 677463813,
    imdbId: "tt0816692",
  },
  "mock-4": {
    id: "mock-4",
    title: "The Matrix",
    year: 1999,
    rating: 8.7,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Action", "Sci-Fi"],
    watched: true,
    description: "A computer hacker learns from mysterious rebels about the true nature of his reality and his role in the war against its controllers.",
    director: "Lana Wachowski",
    cast: ["Keanu Reeves", "Laurence Fishburne", "Carrie-Anne Moss", "Hugo Weaving", "Joe Pantoliano"],
    duration: 136,
    releaseDate: "1999-03-31",
    language: "English",
    country: "United States",
    budget: 63000000,
    revenue: 463517383,
    imdbId: "tt0133093",
  },
  "mock-5": {
    id: "mock-5",
    title: "Pulp Fiction",
    year: 1994,
    rating: 8.9,
    posterUrl: "/placeholder-poster.jpg",
    genres: ["Crime", "Drama"],
    watched: false,
    description: "The lives of two mob hitmen, a boxer, a gangster and his wife, and a pair of diner bandits intertwine in four tales of violence and redemption.",
    director: "Quentin Tarantino",
    cast: ["John Travolta", "Samuel L. Jackson", "Uma Thurman", "Bruce Willis", "Ving Rhames"],
    duration: 154,
    releaseDate: "1994-10-14",
    language: "English",
    country: "United States",
    budget: 8000000,
    revenue: 213928762,
    imdbId: "tt0110912",
  },
};

export function useMovieDetails(movieId: string) {
  const [movie, setMovie] = React.useState<MovieDetails | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    const fetchMovieDetails = async () => {
      try {
        setIsLoading(true);
        setError(null);

        // Check if it's a mock movie ID
        if (movieId.startsWith('mock-')) {
          const mockMovie = MOCK_MOVIE_DETAILS[movieId];
          if (mockMovie) {
            setMovie(mockMovie);
            return;
          }
        }

        // Try to fetch from API
        const res = await fetch(`/api/movies/${movieId}`, { cache: "no-store" });
        if (!res.ok) {
          throw new Error("Failed to fetch movie details");
        }
        const data: MovieDetails = await res.json();
        setMovie(data);
      } catch (e) {
        console.error(e);
        setError("Failed to fetch movie details. Please try again later.");
      } finally {
        setIsLoading(false);
      }
    };

    if (movieId) {
      fetchMovieDetails();
    }
  }, [movieId]);

  const toggleWatched = React.useCallback(async () => {
    if (!movie) return;

    try {
      const res = await fetch(`/api/movies/${movieId}/watch`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ watched: !movie.watched }),
      });

      if (!res.ok) {
        throw new Error("Failed to update watch status");
      }

      setMovie(prev => prev ? { ...prev, watched: !prev.watched } : null);
    } catch (e) {
      console.error(e);
    }
  }, [movie, movieId]);

  return {
    movie,
    isLoading,
    error,
    toggleWatched,
  };
}
