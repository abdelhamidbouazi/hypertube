"use client";
import Link from "next/link";
import Image from "next/image";
import { Input } from "@heroui/input";
import { Button } from "@heroui/button";
import { useState } from "react";
import { useRouter } from "next/navigation";
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

export default function LoginPage() {
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
    <div className="rounded-3xl border border-white/15 bg-white/10 p-8 text-white shadow-2xl backdrop-blur-md">
      <h1 className="text-4xl font-extrabold tracking-tight text-white/95 text-center">
        CINÉTHOS
      </h1>
      <p className="mt-2 text-sm text-white/85 text-center">
        Welcome back — your cinematic universe awaits.
      </p>
      <form className="space-y-5 pt-4" onSubmit={handleSubmit}>
        <Input
          isRequired
          classNames={{
            input: "!text-white",
            label: "!text-white/90",
          }}
          label="Username"
          placeholder="Enter your username"
          radius="sm"
          type="text"
          value={username}
          variant="faded"
          onChange={(e) => setUsername(e.target.value)}
        />
        <Input
          isRequired
          classNames={{
            input: "!text-white",
            label: "!text-white/90",
          }}
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
          variant="faded"
          onChange={(e) => setPassword(e.target.value)}
        />

        <div className="flex items-center justify-between text-sm text-white/75">
          <Link
            className="hover:text-white hover:underline"
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

        <p className="mt-4 text-center text-xs text-white/75">
          Don&apos;t have an account?{" "}
          <Link
            className="text-primary hover:text-pink-200 hover:underline"
            href="/auth/register"
          >
            Sign up
          </Link>
        </p>
      </form>
      <div className="my-6 flex items-center gap-3 text-xs text-white/70">
        <div className="h-px w-full bg-white/20" />
        <span>or</span>
        <div className="h-px w-full bg-white/20" />
      </div>
      <div className="mt-6 grid grid-cols-1 gap-2">
        <Button
          fullWidth
          className="justify-center bg-white text-gray-900 font-medium hover:brightness-95"
          radius="sm"
          onPress={() => redirectToOAuth("fortytwo")}
        >
          <FortyTwoIcon />
          <span className="ml-2">Continue with Intra</span>
        </Button>
        <Button
          fullWidth
          className="justify-center bg-white text-gray-900 font-medium hover:brightness-95"
          radius="sm"
          onPress={() => redirectToOAuth("google")}
        >
          <GoogleIcon size={20} className="ml-2" />
          <span className="ml-2">Continue with Google</span>
        </Button>
        <Button
          fullWidth
          className="justify-center bg-white text-gray-900 font-medium hover:brightness-95"
          radius="sm"
          onPress={() => redirectToOAuth("github")}
        >
          <GithubIcon size={20} className="ml-2" />
          <span className="ml-2">Continue with Github</span>
        </Button>
      </div>
    </div>
  );
}
