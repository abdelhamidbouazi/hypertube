"use client";
import {
  Navbar as HeroUINavbar,
  NavbarContent,
  NavbarMenu,
  NavbarMenuToggle,
  NavbarMenuItem,
  NavbarItem,
} from "@heroui/navbar";
import { Input } from "@heroui/input";
import { Kbd } from "@heroui/kbd";
import NextLink from "next/link";
import NextImage from "next/image";
import { usePathname } from "next/navigation";
import { useEffect, useState, useRef } from "react";
import { SearchIcon } from "@/components/icons";

import { siteConfig } from "@/config/site";
import { useSearchMovies } from "@/lib/hooks";
import { Movie } from "./movies/MoviesCard";
import { useAuthStore } from "@/lib/store";
import { logoutUser } from "@/lib/hooks";

const MovieSearchCard = ({ movie }: { movie: Movie }) => {
  return (
    <NextLink
      href={`/app/movie/${movie.id}`}
      className="flex gap-3 p-3 hover:bg-default-100 dark:hover:bg-default-50 transition-colors rounded-lg cursor-pointer group"
    >
      <div className="relative w-16 h-24 flex-shrink-0 rounded-md overflow-hidden bg-default-200">
        {movie.poster_path ? (
          <NextImage
            src={`https://image.tmdb.org/t/p/w185${movie.poster_path}`}
            alt={movie.title}
            fill
            className="object-cover group-hover:scale-105 transition-transform"
            sizes="64px"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-default-400 text-xs">
            No Image
          </div>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <h3 className="font-semibold text-sm text-foreground group-hover:text-primary transition-colors line-clamp-1">
          {movie.title}
        </h3>
        {movie.release_date && (
          <p className="text-xs text-default-500 mt-1">
            {new Date(movie.release_date).getFullYear()}
          </p>
        )}
        {movie.overview && (
          <p className="text-xs text-default-400 mt-2 line-clamp-2">
            {movie.overview}
          </p>
        )}
      </div>
    </NextLink>
  );
};

export const Navbar = ({
  onSidebarToggle,
  onSidebarCollapseToggle,
  isSidebarCollapsed,
}: {
  onSidebarToggle?: () => void;
  onSidebarCollapseToggle?: () => void;
  isSidebarCollapsed?: boolean;
}) => {
  const pathname = usePathname();
  const { isAuthenticated } = useAuthStore();
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [isSearchFocused, setIsSearchFocused] = useState(false);
  const searchContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(searchQuery);
    }, 500);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  const { movies, isLoading, error } = useSearchMovies(debouncedQuery);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        searchContainerRef.current &&
        !searchContainerRef.current.contains(event.target as Node)
      ) {
        setIsSearchFocused(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === "k") {
        const target = event.target as HTMLElement;
        const isInputFocused =
          target.tagName === "INPUT" ||
          target.tagName === "TEXTAREA" ||
          target.isContentEditable;

        if (!isInputFocused) {
          event.preventDefault();
          if (searchContainerRef.current) {
            const inputElement = searchContainerRef.current.querySelector(
              'input[type="search"]'
            ) as HTMLInputElement;
            if (inputElement) {
              inputElement.focus();
              setIsSearchFocused(true);
            }
          }
        }
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  const showResults = isSearchFocused && debouncedQuery.trim().length > 0;
  const hasResults = movies.length > 0;
  const isDiscoverPage = pathname === "/app/discover";
  const isMovieDetailsPage = pathname?.startsWith("/app/movie/") && pathname !== "/app/movie";
  const isHistoryPage = pathname === "/app/history";
  const isWatchLaterPage = pathname === "/app/watch-later";
  const isSettingsPage = pathname === "/app/settings";
  const shouldShowSearch = isDiscoverPage || isMovieDetailsPage || isHistoryPage || isWatchLaterPage || isSettingsPage;

  const navigationItems = isAuthenticated
    ? [
        { label: "Discover", href: "/app/discover" },
        { label: "Watch Later", href: "/app/watch-later" },
        { label: "Watch History", href: "/app/history" },
        { label: "Settings", href: "/app/settings" },
      ]
    : [{ label: "Discover", href: "/app/discover" }];

  const searchInput = (
    <div ref={searchContainerRef} className="relative w-full max-w-md">
      <Input
        variant="flat"
        size="lg"
        color="primary"
        aria-label="Search"
        value={searchQuery}
        onValueChange={setSearchQuery}
        onFocus={() => setIsSearchFocused(true)}
        classNames={{
          inputWrapper:
            "bg-content1/50 backdrop-blur-sm border-primary/20 hover:border-primary/40 focus-within:border-primary shadow-sm",
          input: "text-foreground placeholder:text-default-500",
        }}
        endContent={
          <Kbd
            className="hidden lg:inline-block bg-primary/10 text-primary border-primary/20"
            keys={["command"]}
          >
            K
          </Kbd>
        }
        labelPlacement="outside"
        placeholder="Search movies..."
        startContent={
          <SearchIcon className="text-base text-primary/70 pointer-events-none flex-shrink-0" />
        }
        type="search"
      />

      {showResults && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-content1 border border-default-200 rounded-lg shadow-lg z-50 max-h-96 overflow-hidden flex flex-col">
          <div className="overflow-y-auto">
            {isLoading ? (
              <div className="p-6 text-center text-default-500">
                <div className="flex items-center justify-center gap-2">
                  <div className="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                  <span className="text-sm">Searching...</span>
                </div>
              </div>
            ) : error ? (
              <div className="p-6 text-center text-danger text-sm">
                Error loading search results
              </div>
            ) : hasResults ? (
              <div className="py-2">
                {movies.map((movie: Movie) => (
                  <MovieSearchCard key={movie.id} movie={movie} />
                ))}
              </div>
            ) : (
              <div className="p-6 text-center text-default-500 text-sm">
                No movies found
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );

  return (
    <HeroUINavbar position="sticky">
      <NavbarContent justify="start">
        <NavbarItem></NavbarItem>
      </NavbarContent>

      <NavbarContent justify="center">
        <NavbarItem className="hidden lg:flex w-full max-w-md">
          {shouldShowSearch && searchInput}
        </NavbarItem>
      </NavbarContent>

      <NavbarContent justify="end">
        <NavbarItem></NavbarItem>
      </NavbarContent>

      <NavbarContent className="lg:hidden" justify="end">
        <NavbarMenuToggle />
      </NavbarContent>

      <NavbarMenu>
        {shouldShowSearch && (
          <div className="px-4 pt-4 pb-2">
            {searchInput}
          </div>
        )}
        <div className="mx-4 mt-2 flex flex-col gap-2">
          {navigationItems.map((item, index) => {
            const isActive = pathname === item.href;
            return (
              <NavbarMenuItem key={`${item.href}-${index}`}>
                <NextLink
                  href={item.href}
                  className={`w-full text-lg ${
                    isActive
                      ? "text-primary font-semibold"
                      : "text-foreground hover:text-primary"
                  } transition-colors`}
                >
                  {item.label}
                </NextLink>
              </NavbarMenuItem>
            );
          })}
          {isAuthenticated && (
            <NavbarMenuItem>
              <button
                onClick={() => logoutUser()}
                className="w-full text-left text-lg text-danger hover:text-danger-600 transition-colors"
              >
                Logout
              </button>
            </NavbarMenuItem>
          )}
        </div>
      </NavbarMenu>
    </HeroUINavbar>
  );
};
