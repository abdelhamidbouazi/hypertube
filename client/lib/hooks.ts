import { useApi } from './swr';
import api from './api';
import { useEffect, useState } from 'react';
import { ContinueWatchingMovie } from '@/types';

// hook for fetching movies list
export const useMovies = () => {
  const [mounted, setMounted] = useState(false);
  const { data, error, isLoading, mutate } = useApi(mounted ? '/movies' : null);
  
  // prevent hydration mismatch by waiting for client mount
  useEffect(() => {
    setMounted(true);
  }, []);
  
  return {
    movies: data || [],
    isLoading: !mounted || isLoading,
    error,
    refetch: mutate,
  };
};

// hook for fetching single movie details
export const useMovieDetailsReq = (movieId: string) => {
  const url = movieId ? `/movies/${movieId}` : null;
  const { data, error, isLoading, mutate } = useApi(url);
  
  return {
    movie: data,
    isLoading,
    error,
    refetch: mutate,
  };
};

// hook for user authentication status
export const useAuth = () => {
  const { data, error, isLoading } = useApi('/users/me');
  
  return {
    user: data,
    isLoading,
    error,
  };
};

// hook for current user data
export const useMe = () => {
  const { data, error, isLoading, mutate } = useApi('/users/me');
  return { user: data, isLoading, error, refetch: mutate };
};

// authenticate user with email and password
export const loginUser = async (email: string, password: string) => {
  const response = await api.post('/auth/login', { email, password });
  return response.data;
};

// register new user account
export const registerUser = async (email: string, password: string, firstName: string, lastName: string) => {
  const response = await api.post('/auth/register', { email, password, FirstName: firstName, LastName: lastName });
  return response.data;
};

// hook for user statistics (movies watched, hours, favorites)
export const useUserStats = () => {
  // TODO: Replace with real API endpoint when available
  // const { data, error, isLoading } = useApi('/users/stats');
  
  // For now, return placeholder data
  return {
    stats: {
      moviesWatched: 0,
      hoursWatched: 0,
      favorites: 0,
    },
    isLoading: false,
    error: null,
  };
};

// hook for user watch history and continue watching
export const useWatchHistory = () => {
  // TODO: Replace with real API endpoint when available
  // const { data, error, isLoading } = useApi('/users/watch-history');
  
  // For now, return placeholder data
  return {
    watchHistory: [] as ContinueWatchingMovie[],
    continueWatching: null as ContinueWatchingMovie | null,
    isLoading: false,
    error: null,
  };
};

// logout user and clear tokens
export const logoutUser = async () => {
  try {
    await api.post('/auth/logout');
  } catch (error) {
    console.error('logout error:', error);
  } finally {
    // clear tokens from storage and redirect
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
      document.cookie = 'token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
      window.location.href = '/auth/login';
    }
  }
};
