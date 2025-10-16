"use client";

import { useState, useEffect } from "react";
import Image from "next/image";

type Movie = {
  id: string;
  title: string;
  year: number;
  rating?: number;
  posterUrl: string;
  genres: string[];
};

export default function CategoryPage({
  params,
}: {
  params: Promise<{ categoryName: string }>;
}) {
  const [categoryName, setCategoryName] = useState<string>("");
  const [movies, setMovies] = useState<Movie[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    params.then(({ categoryName }) => {
      setCategoryName(categoryName);
    });
  }, [params]);

  const category = categoryName
    ? decodeURIComponent(categoryName).toLowerCase()
    : "";

  if (isLoading) {
    return (
      <div className="space-y-6">
        <header className="space-y-2">
          <h1 className="text-3xl font-extrabold tracking-tight">
            <span className="bg-gradient-to-r from-indigo-600 to-pink-600 bg-clip-text text-transparent dark:from-indigo-400 dark:to-pink-400">
              {category.replace("-", " ")} Movies
            </span>
          </h1>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Loading movies...
          </p>
        </header>
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <div
              key={i}
              className="h-64 animate-pulse bg-content2 rounded-2xl"
            />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <header className="space-y-2">
          <h1 className="text-3xl font-extrabold tracking-tight">
            <span className="bg-gradient-to-r from-indigo-600 to-pink-600 bg-clip-text text-transparent dark:from-indigo-400 dark:to-pink-400">
              {category.replace("-", " ")} Movies
            </span>
          </h1>
        </header>
        <div className="text-center text-red-500">{error}</div>
      </div>
    );
  }

  return (
    <article className="space-y-6 text-left">
      <header className="space-y-2">
        <h1 className="text-3xl font-extrabold tracking-tight">
          <span className="bg-gradient-to-r from-indigo-600 to-pink-600 bg-clip-text text-transparent dark:from-indigo-400 dark:to-pink-400">
            {category.replace("-", " ")} Movies
          </span>
        </h1>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Explore movies in the {category} genre.
        </p>
      </header>

      {movies.length > 0 ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-2">
          {movies.map((movie) => (
            <div
              key={movie.id}
              className="group overflow-hidden rounded-2xl border border-gray-200 bg-white/70 shadow-sm backdrop-blur transition hover:shadow-lg dark:border-gray-800 dark:bg-gray-900/60"
            >
              <div className="relative aspect-[2/3] w-full">
                <Image
                  fill
                  alt={movie.title}
                  className="object-cover transition duration-300 group-hover:scale-105"
                  priority={false}
                  sizes="(max-width: 640px) 50vw, 33vw"
                  src={movie.posterUrl || "/placeholder-poster.jpg"}
                />
                <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/50 via-black/0 to-black/0 opacity-70" />
                {typeof movie.rating === "number" && (
                  <div className="absolute left-2 top-2 rounded-full bg-black/70 px-2 py-1 text-xs text-white backdrop-blur">
                    IMDb {movie.rating.toFixed(1)}
                  </div>
                )}
              </div>
              <div className="space-y-1 p-3">
                <h3 className="line-clamp-1 text-sm font-semibold text-gray-900 dark:text-gray-100">
                  {movie.title}
                </h3>
                <p className="text-xs text-gray-600 dark:text-gray-400">
                  {movie.year}
                </p>
                <div className="h-px bg-gradient-to-r from-transparent via-gray-200 to-transparent dark:via-gray-800" />
                <div className="flex gap-2">
                  {movie.genres.slice(0, 2).map((genre) => (
                    <span
                      key={genre}
                      className="rounded-full border border-gray-200 bg-gray-50 px-2 py-0.5 text-[10px] text-gray-700 dark:border-gray-800 dark:bg-gray-800 dark:text-gray-200"
                    >
                      {genre}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="text-center text-gray-500 dark:text-gray-400">
          No movies found for the {category} category.
        </div>
      )}

      <div className="text-center text-xs text-gray-500 dark:text-gray-400">
        Showing {movies.length} result{movies.length !== 1 ? "s" : ""}.
      </div>
    </article>
  );
}
