"use client";

import Link from "next/link";
import { Input } from "@heroui/input";
import { Button } from "@heroui/button";
import Image from "next/image";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { addToast } from "@heroui/toast";

import { loginUser } from "@/lib/hooks";
import { setTokens } from "@/lib/auth";
import { useAuthStore } from "@/lib/store";
import api from "@/lib/api";

function GoogleIcon() {
  return (
    <Image alt="Google" height={20} src="/icons/google-icon.png" width={20} />
  );
}
function GithubIcon() {
  return (
    <Image alt="GitHub" height={20} src="/icons/github-icon.png" width={20} />
  );
}
function FortyTwoIcon() {
  return <Image alt="42" height={20} src="/icons/42-icon.png" width={22} />;
}

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();
  const { login } = useAuthStore();

  // Post to backend OAuth endpoints using a temporary form (to ensure a POST + navigation)
  const redirectToOAuth = (provider: "google" | "fortytwo") => {
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
      const response = await loginUser(email, password);

      if (response.AccessToken) {
        setTokens(response);

        // Fetch user data and store it
        try {
          const userResponse = await api.get("/users/me");
          const userData = userResponse.data;

          login(userData, response.AccessToken);
        } catch {
          // If we can't fetch user data, just set basic info
          login(
            {
              id: "unknown",
              email,
              username: email.split("@")[0],
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
    } catch {
      const errorMessage = "Invalid email or password. Please try again.";

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

      {/* {error && (
        <div className="mt-4 p-3 bg-red-500/20 border border-red-500/30 rounded-lg text-red-200 text-sm">
          {error}
        </div>
      )} */}

      <form className="space-y-5 pt-4" onSubmit={handleSubmit}>
        <Input
          isRequired
          className="text-slate-800"
          classNames={{ input: "text-black" }}
          label="Email"
          placeholder="you@example.com"
          radius="sm"
          type="email"
          value={email}
          variant="faded"
          onChange={(e) => setEmail(e.target.value)}
        />
        <Input
          isRequired
          className="text-slate-800"
          classNames={{ input: "text-black" }}
          label="Password"
          placeholder="••••••••"
          radius="sm"
          type="password"
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
          className="mt-1 bg-gradient-to-r from-indigo-500 to-pink-500 text-white font-semibold shadow-lg transition hover:brightness-110"
          isLoading={isLoading}
          radius="sm"
          type="submit"
        >
          Log In
        </Button>

        <p className="mt-4 text-center text-xs text-white/75">
          Don&apos;t have an account?{" "}
          <Link
            className="text-pink-300 hover:text-pink-200 hover:underline"
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
          <span className="ml-2">Continue with</span>
          <FortyTwoIcon />
        </Button>
        <Button
          fullWidth
          className="justify-center bg-white text-gray-900 font-medium hover:brightness-95"
          radius="sm"
          onPress={() => redirectToOAuth("google")}
        >
          <span className="ml-2">Continue with</span>
          <GoogleIcon />
        </Button>
        <Button
          fullWidth
          className="justify-center bg-white text-gray-900 font-medium hover:brightness-95"
          radius="sm"
          onPress={() =>
            addToast({
              title: "Unavailable",
              description: "GitHub OAuth isn't configured on the server.",
              severity: "warning",
            })
          }
        >
          <span className="ml-2">Continue with</span>
          <GithubIcon />
        </Button>
      </div>
    </div>
  );
}
