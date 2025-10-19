import { useApi } from './swr';
import api from './api';
import { useEffect, useState } from 'react';

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
