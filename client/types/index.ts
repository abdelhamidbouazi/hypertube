import { SVGProps } from "react";

export type IconSvgProps = SVGProps<SVGSVGElement> & {
  size?: number;
};

// Search-related types
export interface SearchFilters {
  query: string;
  year?: string;
  genre?: string;
  minRating?: number;
}

export interface SearchState {
  isLoading: boolean;
  error: string | null;
  results: SearchResult[];
  totalResults: number;
  currentPage: number;
  hasMore: boolean;
}

export interface SearchResult {
  id: number;
  title: string;
  release_date?: string;
  poster_path?: string;
  overview?: string;
  original_language?: string;
  vote_average?: number;
  genre_ids?: number[];
  watched?: boolean;
}
