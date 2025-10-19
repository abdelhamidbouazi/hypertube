import { useApi } from './swr';
import api from './api';
// import type { Movie } from '@/components/movies/MoviesCard';
import { useEffect, useState } from 'react';

export const useMovies = () => {
  const [mounted, setMounted] = useState(false);
  const { data, error, isLoading, mutate } = useApi(mounted ? '/movies' : null);
  
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

export const useAuth = () => {
  const { data, error, isLoading } = useApi('/users/me');
  
  return {
    user: data,
    isLoading,
    error,
  };
};

export const useMe = () => {
  const { data, error, isLoading, mutate } = useApi('/users/me');
  return { user: data, isLoading, error, refetch: mutate };
};

export const loginUser = async (email: string, password: string) => {
  const response = await api.post('/auth/login', { email, password });
  return response.data;
};

export const registerUser = async (email: string, password: string, firstName: string, lastName: string) => {
  const response = await api.post('/auth/register', { email, password, FirstName: firstName, LastName: lastName });
  return response.data;
};

export const logoutUser = async () => {
  try {
    await api.post('/auth/logout');
  } catch (error) {
    console.error('Logout error:', error);
  } finally {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
      document.cookie = 'token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
      window.location.href = '/auth/login';
    }
  }
};
