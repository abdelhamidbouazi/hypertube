"use client";

import Link from "next/link";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Input } from "@heroui/input";
import { Button } from "@heroui/button";
import { addToast } from "@heroui/toast";
import { Eye, EyeOff } from "lucide-react";

import { loginUser, registerUser } from "@/lib/hooks";
import { setTokens } from "@/lib/auth";
import api from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { getErrorMessage } from "@/lib/error-utils";

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

export default function RegisterPage() {
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isVisible, setIsVisible] = useState(false);
  const [isConfirmVisible, setIsConfirmVisible] = useState(false);
  const router = useRouter();
  const { login } = useAuthStore();

  const toggleVisibility = () => setIsVisible(!isVisible);
  const toggleConfirmVisibility = () => setIsConfirmVisible(!isConfirmVisible);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError("");

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      setIsLoading(false);
      return;
    }

    try {
      await registerUser(username, email, password, firstName, lastName);

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
              firstname: firstName || "",
              lastname: lastName || "",
            },
            response.AccessToken
          );
        }

        addToast({
          title: "Welcome",
          description: "Account created successfully.",
          severity: "success",
          timeout: 3000,
          classNames: {
            base: "bg-green-500/10 border-green-500/20",
            title: "text-green-500",
            description: "text-green-400",
          },
        });

        router.push("/app/discover");
      } else {
        router.push("/auth/login");
      }
    } catch (err: any) {
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);

      addToast({
        title: "Registration failed",
        description: errorMessage,
        severity: "danger",
        timeout: 4000,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const redirectToOAuth = (provider: "google" | "fortytwo" | "github") => {
    const base = process.env.NEXT_PUBLIC_API_URL || "/api";
    const form = document.createElement("form");

    form.method = "POST";
    form.action = `${base}/oauth2/${provider}`;
    document.body.appendChild(form);
    form.submit();
  };

  return (
    <div className="rounded-3xl border border-white/15 bg-white/10 p-8 text-white shadow-2xl backdrop-blur-md">
      <h1 className="text-center text-4xl font-extrabold tracking-tight text-white/95">
        Create account
      </h1>
      <p className="mt-2 text-center text-sm text-white/85">
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
          onPress={() => redirectToOAuth("github")}
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

      <form onSubmit={handleSubmit} className="space-y-5">
        <div className="grid grid-cols-2 gap-3">
          <Input
            type="text"
            label="First Name"
            placeholder="John"
            variant="faded"
            radius="sm"
            isRequired
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
          />
          <Input
            type="text"
            label="Last Name"
            placeholder="Doe"
            variant="faded"
            radius="sm"
            isRequired
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
          />
        </div>
        <Input
          type="email"
          label="Email"
          placeholder="john.doe@example.com"
          variant="faded"
          radius="sm"
          isRequired
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <Input
          type="text"
          label="Username"
          placeholder="Enter your username"
          variant="faded"
          radius="sm"
          isRequired
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
        <Input
          type={isVisible ? "text" : "password"}
          label="Password"
          placeholder="••••••••"
          variant="faded"
          radius="sm"
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
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <Input
          type={isConfirmVisible ? "text" : "password"}
          label="Confirm password"
          placeholder="••••••••"
          variant="faded"
          radius="sm"
          isRequired
          endContent={
            <button
              className="focus:outline-none"
              type="button"
              onClick={toggleConfirmVisibility}
            >
              {isConfirmVisible ? (
                <EyeOff className="text-2xl text-default-400 pointer-events-none" />
              ) : (
                <Eye className="text-2xl text-default-400 pointer-events-none" />
              )}
            </button>
          }
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
        />

        <Button
          type="submit"
          radius="sm"
          fullWidth
          className="mt-1 bg-gradient-to-r from-indigo-500 to-pink-500 text-white font-semibold shadow-lg transition hover:brightness-110"
          isLoading={isLoading}
        >
          Create Account
        </Button>

        <p className="mt-4 text-center text-xs text-white/75">
          Already have an account?{" "}
          <Link
            href="/auth/login"
            className="text-pink-300 hover:text-pink-200 hover:underline"
          >
            Log in
          </Link>
        </p>
      </form>
    </div>
  );
}
