"use client";

import React from "react";
import { Card, CardBody, CardFooter } from "@heroui/card";
import { Image } from "@heroui/image";
import { Tooltip } from "@heroui/tooltip";
import Link from "next/link";
import { useWatchlistStore } from "@/lib/store";
import { useMovieDetailsReq } from "@/lib/hooks";
import { Bookmark, BookmarkCheck } from "lucide-react";

export type Movie = {
  id: number;
  title: string;
  release_date?: string;
  poster_path?: string;
  overview?: string;
  original_language?: string;
  watched?: boolean;
  vote_average?: number;
};

export default function MovieCard({ movie }: { movie: Movie }) {
  const { id, title, release_date, poster_path, watched, vote_average } = movie;
  const { isInWatchlist, toggleWatchlist } = useWatchlistStore();
  const [shouldFetchDetails, setShouldFetchDetails] = React.useState(false);
  const cardRef = React.useRef<HTMLDivElement | null>(null);
  const { movie: details } = useMovieDetailsReq(shouldFetchDetails ? String(id) : "");
  const year = release_date ? new Date(release_date).getFullYear() : undefined;

  // use existing rating or fetch from details
  const ratingSource = vote_average ?? details?.vote_average;
  const formattedRating = ratingSource !== undefined && ratingSource !== null
    ? Number(ratingSource).toFixed(1)
    : null;
  const inWatchlist = isInWatchlist(id);

  // lazy load movie details when card becomes visible
  React.useEffect(() => {
    if (!cardRef.current || shouldFetchDetails) return;
    const observer = new IntersectionObserver((entries) => {
      const entry = entries[0];
      if (entry.isIntersecting) {
        setShouldFetchDetails(true);
        observer.disconnect();
      }
    }, { rootMargin: "200px" });
    observer.observe(cardRef.current);
    return () => observer.disconnect();
  }, [shouldFetchDetails]);

  return (
    <Link href={`/app/movie/${id}`} className="group block" onMouseEnter={() => setShouldFetchDetails(true)}>
      <Card className="h-full bg-white dark:bg-gray-800 border-0 shadow-md hover:shadow-xl transition-all duration-300 overflow-hidden group-hover:scale-105 group-hover:-translate-y-1">
        <CardBody className="p-0">
          {/* movie poster container */}
          <div ref={cardRef} className="relative aspect-[2/3] overflow-hidden">
            <Image
              removeWrapper
              src={poster_path ? `https://image.tmdb.org/t/p/w500${poster_path}` : "/placeholder-poster.jpg"}
              alt={`${title} poster`}
              className="w-full h-full object-cover transition-all duration-300 group-hover:blur-sm group-hover:scale-110 z-0"
            />

            {/* hover overlay */}
            <div className="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-all duration-300 z-10" />

            {/* bottom gradient for text readability */}
            <div className="absolute inset-x-0 bottom-0 h-24 bg-gradient-to-t from-black/80 via-black/30 to-transparent z-20" />

            {/* play button overlay */}
            <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-300 z-30 pointer-events-none">
              <div className="relative w-16 h-16">
                <div className="absolute inset-0 rounded-full bg-red-500/40 animate-ping" />
                <div className="relative w-full h-full rounded-full bg-gradient-to-br from-rose-500 to-red-600 shadow-2xl ring-2 ring-white/20 grid place-items-center transition-transform duration-300 group-hover:scale-110">
                  <svg className="w-7 h-7 text-white drop-shadow-[0_2px_4px_rgba(0,0,0,0.35)] fill-current ml-0.5" viewBox="0 0 24 24">
                    <path d="M8 5v14l11-7z" />
                  </svg>
                </div>
              </div>
            </div>

            {/* movie rating badge */}
            {formattedRating !== null && (
              <div className="absolute top-3 right-3 bg-yellow-400 text-gray-900 rounded-full px-2 py-1 shadow-md z-40">
                <div className="flex items-center gap-1">
                  <svg className="w-3 h-3 fill-current" viewBox="0 0 24 24">
                    <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                  </svg>
                  <span className="text-xs font-bold">{formattedRating}</span>
                </div>
              </div>
            )}

            {/* watchlist button */}
            <div className="absolute top-3 left-3 z-40">
              <Tooltip content={inWatchlist ? "In Watchlist" : "Add to Watchlist"}>
                <button
                  type="button"
                  aria-label={inWatchlist ? "In Watchlist" : "Add to Watchlist"}
                  aria-pressed={inWatchlist}
                  onClick={(e) => { e.preventDefault(); e.stopPropagation(); toggleWatchlist(id); }}
                  className={`grid place-items-center w-9 h-9 rounded-full shadow-lg transition-all duration-300 hover:scale-110 active:scale-95 border backdrop-blur-md hover:ring-2 hover:ring-white/40 ${inWatchlist ? "bg-red-600 border-red-500" : "bg-black/40 border-white/20"}`}
                >
                  {inWatchlist ? (
                    <BookmarkCheck className="w-5 h-5 text-white" />
                  ) : (
                    <Bookmark className="w-5 h-5 text-white" />
                  )}
                </button>
              </Tooltip>
            </div>

            {/* movie title and year */}
            <div className="absolute inset-x-0 bottom-0 z-40 p-3">
              <div className="flex items-end justify-between">
                <div className="max-w-[80%]">
                  <h3 className="text-white font-semibold text-sm sm:text-base leading-snug line-clamp-2 drop-shadow">
                    {title}
                  </h3>
                  <p className="text-white/80 text-xs">{year ?? "—"}</p>
                </div>
              </div>
            </div>
          </div>
        </CardBody>

      </Card>
    </Link>
  );
}
