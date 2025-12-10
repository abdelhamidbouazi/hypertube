import React, { useState } from "react";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Button } from "@heroui/button";
import { Textarea } from "@heroui/input";
import { Avatar } from "@heroui/avatar";
import { Divider } from "@heroui/divider";
import { addToast } from "@heroui/toast";
import { Send, MessageSquare } from "lucide-react";
import { addComment } from "@/lib/hooks";
import { formatDistanceToNow } from "date-fns";
import { getErrorMessage } from "@/lib/error-utils";
import PublicProfile from "@/components/PublicProfile";

// Backend returns gorm.Model fields with PascalCase (ID, CreatedAt)
// and custom fields with snake_case (movie_id, user_id, username, content, avatar)
interface Comment {
  id?: number;
  ID?: number; // Backend returns ID (uppercase) from gorm.Model
  username: string;
  avatar?: string; // User's profile picture URL
  content: string;
  date?: string;
  created_at?: string;
  CreatedAt?: string; // Backend returns CreatedAt from gorm.Model
}

interface User {
  username?: string;
  firstname: string;
  lastname: string;
  avatar?: string;
}

interface CommentSectionProps {
  movieId: number;
  comments: Comment[];
  currentUser?: User;
  onCommentAdded: () => void; // Callback to refresh data
}

export default function CommentSection({
  movieId,
  comments,
  currentUser,
  onCommentAdded,
}: CommentSectionProps) {
  const [newComment, setNewComment] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [displayComments, setDisplayComments] = useState<Comment[]>(comments);
  const [isProfileModalOpen, setIsProfileModalOpen] = useState(false);
  const [selectedUsername, setSelectedUsername] = useState<string | null>(null);

  // Sync local state with props when they change
  React.useEffect(() => {
    setDisplayComments(comments);
  }, [comments]);

  // Debug logging for user avatar
  React.useEffect(() => {
    if (currentUser) {
      console.log("Current user data:", currentUser);
      console.log("Avatar URL:", currentUser.avatar);
    }
  }, [currentUser]);

  const handleSubmit = async () => {
    if (!newComment.trim() || !currentUser) return;

    setIsSubmitting(true);
    try {
      const username =
        currentUser.username ||
        `${currentUser.firstname}${currentUser.lastname ? " " + currentUser.lastname : ""}`.trim();
      const apiResponse = await addComment(movieId, newComment, username);

      // Normalize API response to match Comment interface
      // The backend returns ID (uppercase) and CreatedAt
      const normalizedComment: Comment = {
        ID: apiResponse.ID || apiResponse.id,
        username: apiResponse.username || username,
        avatar: apiResponse.avatar || apiResponse.Avatar || currentUser.avatar,
        content: apiResponse.content || newComment,
        CreatedAt:
          apiResponse.CreatedAt || apiResponse.date || new Date().toISOString(),
      };

      // Optimistically update local state
      setDisplayComments((prev) => [normalizedComment, ...prev]);

      setNewComment("");
      onCommentAdded();
    } catch (error) {
      addToast({
        title: "Failed to post comment",
        description: getErrorMessage(error),
        severity: "danger",
        timeout: 4000,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "Just now";
    try {
      return formatDistanceToNow(new Date(dateString), { addSuffix: true });
    } catch (e) {
      return dateString;
    }
  };

  const handleUserClick = (username: string) => {
    setSelectedUsername(username);
    setIsProfileModalOpen(true);
  };

  return (
    <Card className="w-full">
      <CardHeader className="flex gap-3 px-6 pt-6">
        <MessageSquare size={24} className="text-primary" />
        <div className="flex flex-col">
          <p className="text-md font-bold">Comments</p>
          <p className="text-small text-default-500">
            {displayComments.length}{" "}
            {displayComments.length === 1 ? "Comment" : "Comments"}
          </p>
        </div>
      </CardHeader>
      <Divider />
      <CardBody className="px-6 py-6 gap-6">
        {/* Comment Input */}
        {currentUser ? (
          <div className="flex gap-4">
            <Avatar
              src={currentUser.avatar}
              name={`${currentUser.firstname}${currentUser.lastname ? " " + currentUser.lastname : ""}`.trim()}
              className="flex-shrink-0"
                  showFallback
                  imgProps={{
                    onError: () => {
                    },
                  }}
            />
            <div className="flex-grow space-y-2">
              <Textarea
                placeholder="Write a comment..."
                value={newComment}
                onValueChange={setNewComment}
                minRows={2}
                variant="bordered"
              />
              <div className="flex justify-end">
                <Button
                  color="primary"
                  endContent={<Send size={16} />}
                  isLoading={isSubmitting}
                  isDisabled={!newComment.trim()}
                  onPress={handleSubmit}
                  size="sm"
                >
                  Post Comment
                </Button>
              </div>
            </div>
          </div>
        ) : (
          <div className="text-center py-4 text-default-500">
            Please log in to leave a comment.
          </div>
        )}

        {/* Comments List */}
        <div className="space-y-6 mt-4">
          {displayComments.length === 0 ? (
            <div className="text-center text-default-500 py-8" key={movieId}>
              No comments yet. Be the first to share your thoughts!
            </div>
          ) : (
            displayComments.map((comment) => (
              <div key={comment.ID} className="flex gap-4">
                <Avatar
                  src={comment.avatar}
                  name={comment.username}
                  className="flex-shrink-0 cursor-pointer hover:opacity-80 transition-opacity"
                  size="sm"
                  showFallback
                  onClick={() => handleUserClick(comment.username)}
                />
                <div className="flex-grow">
                  <div className="flex items-baseline gap-2">
                    <span
                      className="font-semibold text-small cursor-pointer hover:text-primary transition-colors"
                      onClick={() => handleUserClick(comment.username)}
                    >
                      {comment.username}
                    </span>
                    <span className="text-tiny text-default-400">
                      {formatDate(
                        comment.date || comment.created_at || comment.CreatedAt
                      )}
                    </span>
                  </div>
                  <p className="text-small text-foreground-600 mt-1">
                    {comment.content}
                  </p>
                </div>
              </div>
            ))
          )}
        </div>
      </CardBody>
      <PublicProfile
        username={selectedUsername}
        isOpen={isProfileModalOpen}
        onClose={() => {
          setIsProfileModalOpen(false);
          setSelectedUsername(null);
        }}
      />
    </Card>
  );
}
