"use client";

import Link from "next/link";
import { Input } from "@heroui/input";
import { Button } from "@heroui/button";
import { useState } from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { addToast } from "@heroui/toast";

import { loginUser, registerUser } from "@/lib/hooks";
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

export default function RegisterPage() {
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
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

    if (password !== confirmPassword) {
      addToast({
        title: "Passwords do not match",
        description: "Please ensure both passwords are identical.",
        severity: "warning",
        timeout: 3500,
      });
      setIsLoading(false);

      return;
    }

    try {
      // Create account
      await registerUser(email, password, firstName, lastName);

      // Auto-login to handle tokens just like the login page
      const loginRes = await loginUser(email, password);

      if (loginRes?.AccessToken) {
        setTokens(loginRes);

        // Fetch user data and store it
        try {
          const userResponse = await api.get("/users/me");
          const userData = userResponse.data;

          login(userData, loginRes.AccessToken);
        } catch {
          // If we can't fetch user data, use the registration info
          login(
            {
              id: "unknown",
              email,
              username: email.split("@")[0],
              firstname: firstName,
              lastname: lastName,
            },
            loginRes.AccessToken
          );
        }

        addToast({
          title: "Account created",
          description: "Welcome to Cinéthos. You are now signed in.",
          severity: "success",
          timeout: 3000,
        });

        router.push("/app/discover");
      } else {
        // Fallback: if server did not return tokens, navigate to login
        addToast({
          title: "Account created",
          description: "Please log in to continue.",
          severity: "default",
          timeout: 3000,
        });
        router.push("/auth/login");
      }
    } catch (err: any) {
      const message = err?.response?.data?.message || "Registration failed";

      addToast({
        title: "Registration failed",
        description: message,
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
        Create account
      </h1>
      <p className="mt-2 text-sm text-white/85 text-center">
        Join Cinéthos and start exploring instantly.
      </p>

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

      <div className="my-6 flex items-center gap-3 text-xs text-white/70">
        <div className="h-px w-full bg-white/20" />
        <span>or</span>
        <div className="h-px w-full bg-white/20" />
      </div>

      <form className="space-y-5" onSubmit={handleSubmit}>
        <div className="grid grid-cols-2 gap-3">
          <Input
            isRequired
            classNames={{ input: "text-black" }}
            label="First Name"
            placeholder="John"
            radius="sm"
            type="text"
            value={firstName}
            variant="faded"
            onChange={(e) => setFirstName(e.target.value)}
          />
          <Input
            isRequired
            classNames={{ input: "text-black" }}
            label="Last Name"
            placeholder="Doe"
            radius="sm"
            type="text"
            value={lastName}
            variant="faded"
            onChange={(e) => setLastName(e.target.value)}
          />
        </div>
        <Input
          isRequired
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
          classNames={{ input: "text-black" }}
          label="Password"
          placeholder="••••••••"
          radius="sm"
          type="password"
          value={password}
          variant="faded"
          onChange={(e) => setPassword(e.target.value)}
        />
        <Input
          isRequired
          classNames={{ input: "text-black" }}
          label="Confirm password"
          placeholder="••••••••"
          radius="sm"
          type="password"
          value={confirmPassword}
          variant="faded"
          onChange={(e) => setConfirmPassword(e.target.value)}
        />

        <Button
          fullWidth
          className="mt-1 bg-gradient-to-r from-indigo-500 to-pink-500 text-white font-semibold shadow-lg transition hover:brightness-110"
          isLoading={isLoading}
          radius="sm"
          type="submit"
        >
          Create Account
        </Button>

        <p className="mt-4 text-center text-xs text-white/75">
          Already have an account?{" "}
          <Link
            className="text-pink-300 hover:text-pink-200 hover:underline"
            href="/auth/login"
          >
            Log in
          </Link>
        </p>
      </form>
    </div>
  );
}
