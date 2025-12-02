import React, { useState } from "react";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Button } from "@heroui/button";
import { Textarea } from "@heroui/input";
import { Avatar } from "@heroui/avatar";
import { Divider } from "@heroui/divider";
import { Send, MessageSquare } from "lucide-react";
import { addComment } from "@/lib/hooks";
import { formatDistanceToNow } from "date-fns";

interface Comment {
  id: number;
  username: string;
  content: string;
  date?: string; // Backend sends 'date' but sometimes we might want to handle it differently
  created_at?: string; // Fallback
}

interface User {
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

  // Sync local state with props when they change
  React.useEffect(() => {
    setDisplayComments(comments);
  }, [comments]);

  const handleSubmit = async () => {
    if (!newComment.trim() || !currentUser) return;

    setIsSubmitting(true);
    try {
      const username = `${currentUser.firstname} ${currentUser.lastname}`;
      const addedComment = await addComment(movieId, newComment, username);

      // Optimistically update local state
      setDisplayComments((prev) => [addedComment, ...prev]);

      setNewComment("");
      onCommentAdded();
    } catch (error) {
      console.error("Failed to post comment:", error);
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
              name={`${currentUser.firstname} ${currentUser.lastname}`}
              className="flex-shrink-0"
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
            <div className="text-center text-default-500 py-8">
              No comments yet. Be the first to share your thoughts!
            </div>
          ) : (
            displayComments.map((comment) => (
              <div key={comment.id} className="flex gap-4">
                <Avatar
                  name={comment.username}
                  className="flex-shrink-0"
                  size="sm"
                />
                <div className="flex-grow">
                  <div className="flex items-baseline gap-2">
                    <span className="font-semibold text-small">
                      {comment.username}
                    </span>
                    <span className="text-tiny text-default-400">
                      {formatDate(comment.date || comment.created_at)}
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
    </Card>
  );
}
