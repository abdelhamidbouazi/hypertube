"use client";

import React, { useEffect, useState } from "react";
import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
} from "@heroui/modal";
import { Button } from "@heroui/button";
import { Avatar } from "@heroui/avatar";
import { Spinner } from "@heroui/spinner";
import { User, Globe } from "lucide-react";
import { getUserPublicInfo } from "@/lib/hooks";

interface UserPublicInfo {
  username: string;
  firstname: string;
  lastname: string;
  avatar?: string;
  preferred_language?: string;
}

interface PublicProfileProps {
  username: string | null;
  isOpen: boolean;
  onClose: () => void;
}

export default function PublicProfile({
  username,
  isOpen,
  onClose,
}: PublicProfileProps) {
  const [userInfo, setUserInfo] = useState<UserPublicInfo | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && username) {
      setIsLoading(true);
      setError(null);
      getUserPublicInfo(username)
        .then((data) => {
          setUserInfo(data);
        })
        .catch((err) => {
          setError(
            err.response?.data?.message || "Failed to load user information"
          );
        })
        .finally(() => {
          setIsLoading(false);
        });
    } else if (!isOpen) {
      setUserInfo(null);
      setError(null);
    }
  }, [isOpen, username]);

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      size="md"
      scrollBehavior="inside"
      classNames={{
        base: "max-w-md",
        backdrop: "bg-black/50 backdrop-opacity-40",
      }}
    >
      <ModalContent>
        {(onClose: () => void) => (
          <>
            <ModalHeader className="flex flex-col gap-1">
              <span className="text-xl font-semibold">User Profile</span>
            </ModalHeader>
            <ModalBody>
              {isLoading ? (
                <div className="flex flex-col items-center justify-center py-8">
                  <Spinner size="lg" color="primary" />
                  <p className="text-sm text-default-500 mt-4">
                    Loading profile...
                  </p>
                </div>
              ) : error ? (
                <div className="flex flex-col items-center justify-center py-8">
                  <div className="w-16 h-16 rounded-full bg-danger/10 flex items-center justify-center mb-4">
                    <User className="w-8 h-8 text-danger" />
                  </div>
                  <p className="text-sm text-danger text-center">{error}</p>
                </div>
              ) : userInfo ? (
                <div className="flex flex-col items-center gap-6 py-4">
                  <Avatar
                    src={userInfo.avatar}
                    name={`${userInfo.firstname} ${userInfo.lastname}`}
                    size="lg"
                    className="w-24 h-24 text-large"
                  />
                  <div className="flex flex-col items-center gap-2 w-full">
                    <h3 className="text-xl font-semibold">
                      {userInfo.firstname} {userInfo.lastname}
                    </h3>
                    <p className="text-sm text-default-500">@{userInfo.username}</p>
                  </div>
                  <div className="w-full space-y-3 pt-2">
                    {userInfo.preferred_language && (
                      <div className="flex items-center gap-3 p-3 rounded-lg bg-default-100 dark:bg-default-50">
                        <Globe size={18} className="text-default-500" />
                        <div className="flex flex-col">
                          <span className="text-xs text-default-500">
                            Preferred Language
                          </span>
                          <span className="text-sm font-medium">
                            {userInfo.preferred_language.toUpperCase()}
                          </span>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              ) : null}
            </ModalBody>
            <ModalFooter>
              <Button color="primary" variant="flat" onPress={onClose}>
                Close
              </Button>
            </ModalFooter>
          </>
        )}
      </ModalContent>
    </Modal>
  );
}
