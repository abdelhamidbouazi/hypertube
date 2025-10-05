import Image from "next/image";
import { notFound } from "next/navigation";

const MOVIES: Record<string, Array<{
  id: string;
  title: string;
  year: number;
  rating?: number;
  poster: string; 
}>> = {
  "action": [
    { id: "m1", title: "Steel Horizon", year: 2024, rating: 7.6, poster: "/posters/steel-horizon.jpg" },
    { id: "m2", title: "Frontline Echo", year: 2023, rating: 7.9, poster: "/posters/frontline-echo.jpg" },
    { id: "m3", title: "Zero Hour", year: 2022, rating: 7.1, poster: "/posters/zero-hour.jpg" },
  ],
  "comedy": [
    { id: "m4", title: "Laugh Track", year: 2023, rating: 7.2, poster: "/posters/laugh-track.jpg" },
    { id: "m5", title: "Odd Couple 2.0", year: 2022, rating: 6.9, poster: "/posters/odd-couple.jpg" },
  ],
  "drama": [
    { id: "m6", title: "Paper Skies", year: 2024, rating: 8.2, poster: "/posters/paper-skies.jpg" },
  ],
};

const VALID = Object.keys(MOVIES);

function MovieCard({
  title,
  year,
  rating,
  poster,
}: {
  title: string;
  year: number;
  rating?: number;
  poster: string;
}) {
  return (
    <div className="group overflow-hidden rounded-2xl border border-gray-200 bg-white/70 shadow-sm backdrop-blur transition hover:shadow-lg dark:border-gray-800 dark:bg-gray-900/60">
      <div className="relative aspect-[2/3] w-full">
        <Image
          src={poster}
          alt={title}
          fill
          sizes="(max-width: 640px) 50vw, 33vw"
          className="object-cover transition duration-300 group-hover:scale-105"
          priority={false}
        />
        <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/50 via-black/0 to-black/0 opacity-70" />
        {typeof rating === "number" && (
          <div className="absolute left-2 top-2 rounded-full bg-black/70 px-2 py-1 text-xs text-white backdrop-blur">
            IMDb {rating.toFixed(1)}
          </div>
        )}
      </div>
      <div className="space-y-1 p-3">
        <h3 className="line-clamp-1 text-sm font-semibold text-gray-900 dark:text-gray-100">
          {title}
        </h3>
        <p className="text-xs text-gray-600 dark:text-gray-400">{year}</p>
        <div className="h-px bg-gradient-to-r from-transparent via-gray-200 to-transparent dark:via-gray-800" />
        <div className="flex gap-2">
          <span className="rounded-full border border-gray-200 bg-gray-50 px-2 py-0.5 text-[10px] text-gray-700 dark:border-gray-800 dark:bg-gray-800 dark:text-gray-200">
            Thumbnail
          </span>
          <span className="rounded-full border border-gray-200 bg-gray-50 px-2 py-0.5 text-[10px] text-gray-700 dark:border-gray-800 dark:bg-gray-800 dark:text-gray-200">
            Sorted &amp; Filterable
          </span>
        </div>
      </div>
    </div>
  );
}

export default function CategoryPage({
  params,
}: {
  params: { "categoryName": string };
}) {
  const category = decodeURIComponent(params["categoryName"]).toLowerCase();
  if (!VALID.includes(category)) return notFound();

  const movies = MOVIES[category] ?? [];

  return (
    <article className="space-y-6 text-left">
      <header className="space-y-2">
        <h1 className="text-3xl font-extrabold tracking-tight">
          <span className="bg-gradient-to-r from-indigo-600 to-pink-600 bg-clip-text text-transparent dark:from-indigo-400 dark:to-pink-400">
            {category.replace("-", " ")} Movies
          </span>
        </h1>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Explore curated titles. Light &amp; dark mode ready.
        </p>
      </header>

      {/* Two columns within BlogLayout’s width; adapts down to 1 on very small screens */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-2">
        {movies.map((m) => (
          <MovieCard
            key={m.id}
            title={m.title}
            year={m.year}
            rating={m.rating}
            poster={m.poster}
          />
        ))}
      </div>

      {/* Replace with infinite scroll when hooking real data */}
      <div className="text-center text-xs text-gray-500 dark:text-gray-400">
        Showing {movies.length} result{movies.length !== 1 ? "s" : ""}.
      </div>
    </article>
  );
}
