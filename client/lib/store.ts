import { create } from "zustand";
import { persist } from "zustand/middleware";

// user data interface
interface User {
  id: string;
  email: string;
  username: string;
  firstname: string;
  lastname: string;
  avatar?: string;
  preferred_language?: string;
}

// authentication state interface
interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (user: User, token: string) => void;
  logout: () => void;
  setUser: (user: User) => void;
}

// global authentication store with persistence
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      login: (user, token) => set({ user, token, isAuthenticated: true }),
      logout: () => set({ user: null, token: null, isAuthenticated: false }),
      setUser: (user) => set({ user }),
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

// movie filtering options interface
export interface MovieFilters {
  query: string;
  sort: "name" | "year" | "rating";
  selectedGenres: string[];
  minRating: number;
  yearRange: [number, number];
}

// movie filter state interface
interface FilterState {
  filters: MovieFilters;
  setQuery: (query: string) => void;
  setSort: (sort: "name" | "year" | "rating") => void;
  toggleGenre: (genre: string) => void;
  setMinRating: (rating: number) => void;
  setYearRange: (range: [number, number]) => void;
  resetFilters: () => void;
}

// movie filter store with persistence
export const useFilterStore = create<FilterState>()(
  persist(
    (set) => ({
      filters: {
        query: "",
        sort: "name",
        selectedGenres: [],
        minRating: 0,
        yearRange: [2000, new Date().getFullYear()],
      },
      setQuery: (query) =>
        set((state) => ({ filters: { ...state.filters, query } })),
      setSort: (sort) =>
        set((state) => ({ filters: { ...state.filters, sort } })),
      toggleGenre: (genre) =>
        set((state) => ({
          filters: {
            ...state.filters,
            selectedGenres: state.filters.selectedGenres.includes(genre)
              ? state.filters.selectedGenres.filter((g) => g !== genre)
              : [...state.filters.selectedGenres, genre],
          },
        })),
      setMinRating: (minRating) =>
        set((state) => ({ filters: { ...state.filters, minRating } })),
      setYearRange: (yearRange) =>
        set((state) => ({ filters: { ...state.filters, yearRange } })),
      resetFilters: () =>
        set({
          filters: {
            query: "",
            sort: "name",
            selectedGenres: [],
            minRating: 0,
            yearRange: [2000, new Date().getFullYear()],
          },
        }),
    }),
    {
      name: "filter-storage",
    }
  )
);

// watchlist state interface
interface WatchlistState {
  watchlistIds: number[];
  addToWatchlist: (id: number) => void;
  removeFromWatchlist: (id: number) => void;
  toggleWatchlist: (id: number) => void;
  isInWatchlist: (id: number) => boolean;
}

// watchlist store for managing saved movies
export const useWatchlistStore = create<WatchlistState>()(
  persist(
    (set, get) => ({
      watchlistIds: [],
      addToWatchlist: (id) =>
        set((state) => ({
          watchlistIds: state.watchlistIds.includes(id)
            ? state.watchlistIds
            : [...state.watchlistIds, id],
        })),
      removeFromWatchlist: (id) =>
        set((state) => ({
          watchlistIds: state.watchlistIds.filter((x) => x !== id),
        })),
      toggleWatchlist: (id) => {
        const isIn = get().watchlistIds.includes(id);

        if (isIn) {
          set((state) => ({
            watchlistIds: state.watchlistIds.filter((x) => x !== id),
          }));
        } else {
          set((state) => ({ watchlistIds: [...state.watchlistIds, id] }));
        }
      },
      isInWatchlist: (id) => get().watchlistIds.includes(id),
    }),
    {
      name: "watchlist-storage",
      partialize: (state) => ({ watchlistIds: state.watchlistIds }),
    }
  )
);
