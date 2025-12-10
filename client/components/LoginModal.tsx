"use client";

import { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { useRouter } from "next/navigation";
import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
} from "@heroui/modal";
import { Input } from "@heroui/input";
import { Button } from "@heroui/button";
import { addToast } from "@heroui/toast";
import { Eye, EyeOff } from "lucide-react";

import { loginUser } from "@/lib/hooks";
import { setTokens } from "@/lib/auth";
import api from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { getErrorMessage } from "@/lib/error-utils";
import { GoogleIcon, GithubIcon } from "@/components/icons";

function FortyTwoIcon() {
  return (
    <Image
      src="/images/42_logo.png"
      alt="42"
      width={20}
      height={20}
      className="h-5 w-5"
    />
  );
}

interface LoginModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function LoginModal({ isOpen, onClose }: LoginModalProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isVisible, setIsVisible] = useState(false);
  const router = useRouter();
  const { login } = useAuthStore();

  const toggleVisibility = () => setIsVisible(!isVisible);

  const redirectToOAuth = (provider: "google" | "fortytwo" | "github") => {
    const base = process.env.NEXT_PUBLIC_API_URL || "/api";
    const form = document.createElement("form");

    form.method = "POST";
    form.action = `${base}/oauth2/${provider}`;
    document.body.appendChild(form);
    form.submit();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const response = await loginUser(username, password);

      if (response.AccessToken) {
        setTokens(response);

        try {
          const userResponse = await api.get("/users/me");
          const userData = userResponse.data;

          login(userData, response.AccessToken);
        } catch {
          login(
            {
              id: "unknown",
              email: "unknown",
              username: username,
              firstname: "",
              lastname: "",
            },
            response.AccessToken
          );
        }

        addToast({
          title: "Welcome back",
          description: "You have logged in successfully.",
          severity: "success",
          timeout: 3000,
        });

        onClose();
        router.push("/app/discover");
      }
    } catch (error) {
      const errorMessage = getErrorMessage(error);

      addToast({
        title: "Login failed",
        description: errorMessage,
        severity: "danger",
        timeout: 4000,
      });
    } finally {
      setIsLoading(false);
    }
  };

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
              <h1 className="text-2xl font-extrabold tracking-tight text-center">
                CINÉTHOS
              </h1>
              <p className="text-sm text-default-500 text-center font-normal">
                Welcome back — your cinematic universe awaits.
              </p>
            </ModalHeader>
            <ModalBody>
              <form className="space-y-5" onSubmit={handleSubmit}>
                <Input
                  isRequired
                  label="Username"
                  placeholder="Enter your username"
                  radius="sm"
                  type="text"
                  value={username}
                  variant="bordered"
                  onChange={(e) => setUsername(e.target.value)}
                />
                <Input
                  isRequired
                  endContent={
                    <button
                      className="focus:outline-none"
                      type="button"
                      onClick={toggleVisibility}
                    >
                      {isVisible ? (
                        <EyeOff className="text-2xl text-default-400 pointer-events-none" />
                      ) : (
                        <Eye className="text-2xl text-default-400 pointer-events-none" />
                      )}
                    </button>
                  }
                  label="Password"
                  placeholder="••••••••"
                  radius="sm"
                  type={isVisible ? "text" : "password"}
                  value={password}
                  variant="bordered"
                  onChange={(e) => setPassword(e.target.value)}
                />

                <div className="flex items-center justify-between text-sm">
                  <Link
                    className="text-primary hover:underline"
                    href="/forgot-password"
                  >
                    Forgot password?
                  </Link>
                </div>

                <Button
                  fullWidth
                  color="primary"
                  isLoading={isLoading}
                  radius="sm"
                  type="submit"
                >
                  Log In
                </Button>

                <p className="text-center text-xs text-default-500">
                  Don&apos;t have an account?{" "}
                  <Link
                    className="text-primary hover:underline"
                    href="/auth/register"
                  >
                    Sign up
                  </Link>
                </p>
              </form>
              <div className="my-6 flex items-center gap-3 text-xs text-default-500">
                <div className="h-px w-full bg-default-200" />
                <span>or</span>
                <div className="h-px w-full bg-default-200" />
              </div>
              <div className="grid grid-cols-1 gap-2">
                <Button
                  fullWidth
                  className="justify-center bg-default-100 text-foreground font-medium hover:bg-default-200"
                  radius="sm"
                  onPress={() => redirectToOAuth("fortytwo")}
                >
                  <FortyTwoIcon />
                  <span className="ml-2">Continue with Intra</span>
                </Button>
                <Button
                  fullWidth
                  className="justify-center bg-default-100 text-foreground font-medium hover:bg-default-200"
                  radius="sm"
                  onPress={() => redirectToOAuth("google")}
                >
                  <GoogleIcon size={20} className="ml-2" />
                  <span className="ml-2">Continue with Google</span>
                </Button>
                <Button
                  fullWidth
                  className="justify-center bg-default-100 text-foreground font-medium hover:bg-default-200"
                  radius="sm"
                  onPress={() => redirectToOAuth("github")}
                >
                  <GithubIcon size={20} className="ml-2" />
                  <span className="ml-2">Continue with Github</span>
                </Button>
              </div>
            </ModalBody>
          </>
        )}
      </ModalContent>
    </Modal>
  );
}

