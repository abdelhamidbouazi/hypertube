import { create } from "zustand";
import { persist } from "zustand/middleware";

interface User {
  id: string;
  email: string;
  username?: string;
  firstname?: string;
  lastname?: string;
  FirstName?: string;
  LastName?: string;
  avatar?: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (user: User, token: string) => void;
  logout: () => void;
  setUser: (user: User) => void;
}

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

interface MovieFilters {
  query: string;
  sort: "name" | "year" | "rating";
  selectedGenres: string[];
  minRating: number;
  yearRange: [number, number];
}

interface FilterState {
  filters: MovieFilters;
  currentPage: number;
  hasMore: boolean;
  setQuery: (query: string) => void;
  setSort: (sort: "name" | "year" | "rating") => void;
  toggleGenre: (genre: string) => void;
  setMinRating: (rating: number) => void;
  setYearRange: (range: [number, number]) => void;
  setCurrentPage: (page: number) => void;
  setHasMore: (hasMore: boolean) => void;
  resetFilters: () => void;
}

export const useFilterStore = create<FilterState>()(
  persist(
    (set) => ({
      filters: {
        query: "",
        sort: "name",
        selectedGenres: [],
        minRating: 0,
        yearRange: [1950, new Date().getFullYear()],
      },
      currentPage: 1,
      hasMore: true,
      setQuery: (query) =>
        set((state) => ({
          filters: { ...state.filters, query },
          currentPage: 1,
        })),
      setSort: (sort) =>
        set((state) => ({
          filters: { ...state.filters, sort },
          currentPage: 1,
        })),
      toggleGenre: (genre) =>
        set((state) => ({
          filters: {
            ...state.filters,
            selectedGenres: state.filters.selectedGenres.includes(genre)
              ? state.filters.selectedGenres.filter((g) => g !== genre)
              : [...state.filters.selectedGenres, genre],
          },
          currentPage: 1,
        })),
      setMinRating: (minRating) =>
        set((state) => ({
          filters: { ...state.filters, minRating },
          currentPage: 1,
        })),
      setYearRange: (yearRange) =>
        set((state) => ({
          filters: { ...state.filters, yearRange },
          currentPage: 1,
        })),
      setCurrentPage: (page) => set(() => ({ currentPage: page })),
      setHasMore: (hasMore) => set(() => ({ hasMore })),
      resetFilters: () =>
        set({
          filters: {
            query: "",
            sort: "name",
            selectedGenres: [],
            minRating: 0,
            yearRange: [1950, new Date().getFullYear()],
          },
          currentPage: 1,
          hasMore: true,
        }),
    }),
    {
      name: "filter-storage",
    }
  )
);
