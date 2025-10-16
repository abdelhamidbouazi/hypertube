"use client";

import React from "react";
import useSWR from "swr";

import api from "@/lib/api";
import { useAuthStore } from "@/lib/store";

export interface Comment {
  id: number;
  movie_id: number;
  user_id: number;
  username: string;
  content: string;
  created_at: string;
  updated_at: string;
}

interface CommentResponse {
  id: number;
  username: string;
  date: string;
  content: string;
}

export function useComments(movieId: string) {
  const { user } = useAuthStore();
  const {
    data: allComments,
    error,
    mutate,
  } = useSWR<CommentResponse[]>(
    "/comments",
    async (url: string) => {
      try {
        const response = await api.get(url);

        return response.data;
      } catch (error: any) {
        if (
          error?.response?.status === 404 ||
          error?.code === "ECONNREFUSED" ||
          error?.name === "AxiosError"
        ) {
          return [];
        }
        throw error;
      }
    },
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      fallbackData: [], // Provide fallback data
      errorRetryCount: 1, // Only retry once
      errorRetryInterval: 5000, // Wait 5 seconds before retry
    }
  );

  // Since we're working with existing endpoints and the backend
  // doesn't filter by movie_id yet, we'll show all comments for now
  // In a real implementation, you'd want a /comments/movie/:id endpoint
  const comments = React.useMemo(() => {
    if (!allComments) return undefined;

    // Convert response format to our format
    return allComments.map(
      (comment): Comment => ({
        id: comment.id,
        movie_id: parseInt(movieId), // Associate with current movie
        user_id: 0, // Not available in current response
        username: comment.username,
        content: comment.content,
        created_at: comment.date,
        updated_at: comment.date,
      })
    );
  }, [allComments, movieId]);

  const addComment = async (content: string) => {
    // Try to get username from user store, fallback to a default if not available
    const username = user?.username || "Anonymous User";

    try {
      await api.post("/comments/add", {
        movie_id: parseInt(movieId),
        username,
        content,
      });

      // Refresh comments after adding
      await mutate();
    } catch (error) {
      throw error;
    }
  };

  const refreshComments = async () => {
    await mutate();
  };

  return {
    comments,
    isLoading: !allComments && !error,
    error,
    addComment,
    refreshComments,
  };
}
