"use client";

import React from "react";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Button } from "@heroui/button";
import { Textarea } from "@heroui/input";
import { Avatar } from "@heroui/avatar";
import { Spinner } from "@heroui/spinner";
import { MessageCircle, Send, Clock } from "lucide-react";

import { useAuthStore } from "@/lib/store";
import { useAuth } from "@/lib/hooks";
import api from "@/lib/api";

export interface Comment {
  id: number;
  username: string;
  content: string;
  date?: string;
  CreatedAt?: string;
  UpdatedAt?: string;
  DeletedAt?: string | null;
  ID?: number;
  movie_id?: number;
  user_id?: number;
}

interface CommentSectionProps {
  movieId: string;
  comments?: Comment[];
  isLoading?: boolean;
  error?: any;
  onCommentAdded?: () => void;
}

export function CommentSection({
  movieId,
  comments = [],
  isLoading = false,
  error = null,
  onCommentAdded,
}: CommentSectionProps) {
  const [newComment, setNewComment] = React.useState("");
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [isClient, setIsClient] = React.useState(false);
  const { user: storeUser, isAuthenticated } = useAuthStore();
  const { user: apiUser } = useAuth();

  // Use API user data as primary source, fallback to store
  const user = apiUser || storeUser;

  React.useEffect(() => {
    setIsClient(true);
  }, []);

  const hasToken = React.useMemo(() => {
    if (typeof window === "undefined") return false;

    return document.cookie.includes("token=");
  }, [isClient]);

  const userIsAuthenticated = isAuthenticated || hasToken;

  const handleSubmitComment = async () => {
    if (!newComment.trim()) return;

    setIsSubmitting(true);
    try {
      // Get username from user data - try multiple fallback options
      let username = user?.username;

      // Try FirstName/LastName (from API) first, then lowercase versions
      if (!username && user?.FirstName && user?.LastName) {
        username = `${user.FirstName} ${user.LastName}`;
      } else if (!username && user?.firstname && user?.lastname) {
        username = `${user.firstname} ${user.lastname}`;
      } else if (!username && user?.email) {
        username = user.email.split("@")[0];
      }

      if (!username) {
        username = "Anonymous User";
      }

      await api.post("/comments/add", {
        movie_id: parseInt(movieId),
        username, // Keep sending username for now since backend expects it
        content: newComment,
      });

      setNewComment("");

      // Trigger refresh of movie details to get updated comments
      if (onCommentAdded) {
        onCommentAdded();
      }
    } catch {
      // Error handling - could show a toast notification here
    } finally {
      setIsSubmitting(false);
    }
  };

  const formatDate = (dateString: string) => {
    // Handle various cases where date might be missing or invalid
    if (
      !dateString ||
      dateString === "null" ||
      dateString === "undefined" ||
      dateString.trim() === ""
    ) {
      return "Just now"; // Default for new comments
    }

    // The backend sends dates in ISO format: "2006-01-02T15:04:05Z07:00"
    // JavaScript Date constructor handles this format well
    const date = new Date(dateString);

    // Check if date is valid
    if (isNaN(date.getTime())) {
      return "Just now"; // Fallback for invalid dates
    }

    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 30) return `${diffDays}d ago`;

    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  const getInitials = (username: string) => {
    return username
      .split(" ")
      .map((name) => name[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  if (error) {
    return (
      <Card>
        <CardBody className="p-6">
          <div className="flex items-center justify-center py-8">
            <div className="text-center">
              <p className="text-foreground-500 mb-2">
                Failed to load comments
              </p>
              <p className="text-xs text-foreground-400">
                {error?.message || "Something went wrong!"}
              </p>
              <Button
                className="mt-3"
                size="sm"
                variant="flat"
                onPress={() => onCommentAdded && onCommentAdded()}
              >
                Try again
              </Button>
            </div>
          </div>
        </CardBody>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center gap-2">
          <MessageCircle className="text-primary" size={24} />
          <h2 className="text-2xl font-bold">
            Comments {comments && `(${comments.length})`}
          </h2>
        </div>
      </CardHeader>

      <CardBody className="px-6 pb-6">
        {/* Add Comment Form */}
        {isClient && userIsAuthenticated ? (
          <div className="mb-6 space-y-3">
            <Textarea
              classNames={{
                input: "resize-none",
              }}
              maxRows={6}
              minRows={3}
              placeholder="Share your thoughts about this movie..."
              value={newComment}
              variant="bordered"
              onChange={(e) => setNewComment(e.target.value)}
            />
            <div className="flex justify-end">
              <Button
                color="primary"
                isDisabled={!newComment.trim() || isSubmitting}
                isLoading={isSubmitting}
                startContent={<Send size={16} />}
                onPress={handleSubmitComment}
              >
                {isSubmitting ? "Posting..." : "Post Comment"}
              </Button>
            </div>
          </div>
        ) : !isClient ? (
          <div className="mb-6 p-4 text-center bg-content1/50 rounded-lg border border-divider">
            <Spinner size="sm" />
            <p className="text-sm text-foreground-400 mt-2">Loading...</p>
          </div>
        ) : (
          <div className="mb-6 p-4 text-center bg-content1/50 rounded-lg border border-divider">
            <p className="text-foreground-500 mb-2">
              Sign in to share your thoughts
            </p>
            <p className="text-sm text-foreground-400">
              You need to be logged in to post comments
            </p>
          </div>
        )}

        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Spinner label="Loading comments..." />
          </div>
        ) : comments && comments.length > 0 ? (
          <div className="space-y-4">
            {comments.map((comment: Comment) => {
              // Check if this comment belongs to the current user
              const isCurrentUser =
                user &&
                (comment.username === user.username ||
                  comment.username === `${user.FirstName} ${user.LastName}` ||
                  comment.username === `${user.firstname} ${user.lastname}` ||
                  comment.username === user.email?.split("@")[0]);

              return (
                <div
                  key={comment.id}
                  className={`flex gap-3 p-4 rounded-lg border w-full ${
                    isCurrentUser
                      ? "bg-primary/10 border-primary/30" // Current user: primary color
                      : "bg-content1/50 border-divider" // Other users: default style
                  }`}
                >
                  <Avatar
                    className="flex-shrink-0"
                    color={isCurrentUser ? "success" : "primary"}
                    name={getInitials(comment.username)}
                    size="sm"
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-semibold text-sm">
                        {comment.username}
                        {isCurrentUser && (
                          <span className="text-xs text-success ml-1">
                            (You)
                          </span>
                        )}
                      </span>
                      <div className="flex items-center gap-1 text-xs text-foreground-500">
                        <Clock size={12} />
                        <span>
                          {formatDate(comment.CreatedAt || comment.date || "")}
                        </span>
                      </div>
                    </div>
                    <p className="text-sm text-foreground-700 leading-relaxed break-words">
                      {comment.content}
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <MessageCircle className="text-foreground-300 mb-3" size={48} />
            <p className="text-foreground-500 mb-1">No comments yet</p>
            <p className="text-sm text-foreground-400">
              Be the first to share your thoughts about this movie!
            </p>
          </div>
        )}
      </CardBody>
    </Card>
  );
}
