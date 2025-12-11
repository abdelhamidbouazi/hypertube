"use client";

import React from "react";
import Link from "next/link";
import { Tooltip } from "@heroui/tooltip";
import { Bookmark, BookmarkCheck } from "lucide-react";
import { Swiper, SwiperSlide } from "swiper/react";
import { Autoplay, Pagination, Parallax, EffectFade } from "swiper/modules";
import "swiper/css";
import "swiper/css/pagination";
import "swiper/css/effect-fade";
import "swiper/css/parallax";
import { useWatchlistStore, useAuthStore } from "@/lib/store";

type Slide = {
  id: number;
  title: string;
  backdrop_path?: string;
  poster_path?: string;
  overview?: string;
};

export default function HeroSection({ slides }: { slides: Slide[] }) {
  const { user } = useAuthStore();
  const { isInWatchlist, toggleWatchlist } = useWatchlistStore();
  return (
    <section className="relative">
      <Swiper
        modules={[Autoplay, Pagination, Parallax, EffectFade]}
        autoplay={{ delay: 4200, disableOnInteraction: false, pauseOnMouseEnter: true }}
        pagination={{ clickable: true }}
        effect="fade"
        parallax
        speed={900}
        loop
        className="h-[50vh] sm:h-[60vh] lg:h-[70vh] rounded-2xl overflow-hidden"
      >
        {slides.map((s) => (
          <SwiperSlide key={s.id}>
            <div className="block h-full">
              <div className="relative h-full w-full">
                <img
                  src={
                    s.backdrop_path
                      ? `https://image.tmdb.org/t/p/original${s.backdrop_path}`
                      : s.poster_path
                      ? `https://image.tmdb.org/t/p/w780${s.poster_path}`
                      : "/movies-bg.jpg"
                  }
                  alt={s.title}
                  className="absolute inset-0 h-full w-full object-cover"
                  data-swiper-parallax="30%"
                />

                <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/30 to-transparent" />

                {/* Watchlist button (matches card style) */}
                <div className="absolute top-4 left-4 z-40">
                  <Tooltip content={isInWatchlist(s.id) ? "In Watchlist" : "Add to Watchlist"}>
                    <button
                      type="button"
                      aria-label={isInWatchlist(s.id) ? "In Watchlist" : "Add to Watchlist"}
                      aria-pressed={isInWatchlist(s.id)}
                      onClick={(e) => { e.preventDefault(); e.stopPropagation(); toggleWatchlist(s.id); }}
                      className={`grid place-items-center w-10 h-10 rounded-full shadow-lg transition-all duration-300 hover:scale-110 active:scale-95 border backdrop-blur-md hover:ring-2 hover:ring-white/40 ${isInWatchlist(s.id) ? "bg-red-600 border-red-500" : "bg-black/40 border-white/20"}`}
                    >
                      {isInWatchlist(s.id) ? (
                        <BookmarkCheck className="w-5 h-5 text-white" />
                      ) : (
                        <Bookmark className="w-5 h-5 text-white" />
                      )}
                    </button>
                  </Tooltip>
                </div>

                <div className="absolute bottom-0 left-0 right-0 p-6 sm:p-10">
                  <div className="max-w-3xl">
                    <h2 className="text-white text-3xl sm:text-4xl lg:text-5xl font-extrabold tracking-tight drop-shadow" data-swiper-parallax="-120">
                      {s.title}
                    </h2>
                    {s.overview && (
                      <p className="mt-4 text-white/85 text-base sm:text-lg line-clamp-3" data-swiper-parallax="-200">
                        {s.overview}
                      </p>
                    )}
                    <div className="mt-6" data-swiper-parallax="-300">
                      <Link href={`/app/movie/${s.id}`} className="group relative inline-flex">
                        <span
                          className="absolute -inset-1 rounded-full bg-red-500/40 blur-md opacity-0 transition duration-300 transform group-hover:opacity-100 group-hover:scale-110"
                          aria-hidden
                        />
                        <span className="relative inline-flex items-center gap-2 rounded-full bg-red-600 text-white px-5 py-2.5 text-sm font-semibold shadow-md transition-all duration-300 group-hover:scale-105 active:scale-95 ring-0 group-hover:ring-2 group-hover:ring-red-300/60">
                          Watch now
                          <svg className="w-4 h-4 fill-current transition-transform duration-300 group-hover:translate-x-1" viewBox="0 0 24 24">
                            <path d="M8 5v14l11-7z" />
                          </svg>
                        </span>
                      </Link>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </SwiperSlide>
        ))}
      </Swiper>

      {/* Fancy pagination styling */}
      <style jsx global>{`
        .swiper-pagination-bullets { bottom: 14px !important; }
        .swiper-pagination-bullet {
          width: 8px; height: 8px; background: rgba(255,255,255,0.6); opacity: 1; transition: all .3s ease;
        }
        .swiper-pagination-bullet-active {
          width: 22px; border-radius: 9999px; background: #ffffff;
        }
      `}</style>
    </section>
  );
}


