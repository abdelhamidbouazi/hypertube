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

export default function RegisterPage() {
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [avatar, setAvatar] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [isVisible, setIsVisible] = useState(false);
  const [isConfirmVisible, setIsConfirmVisible] = useState(false);
  const router = useRouter();
  const { login } = useAuthStore();

  const toggleVisibility = () => setIsVisible(!isVisible);
  const toggleConfirmVisibility = () => setIsConfirmVisible(!isConfirmVisible);

  const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setAvatar(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setAvatarPreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

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
      await registerUser(
        username,
        email,
        password,
        firstName,
        lastName,
        avatar || undefined
      );

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
    <div className="rounded-3xl border border-white/15 bg-white/10 p-4 text-white shadow-2xl backdrop-blur-md">
      <h1 className="text-center text-2xl font-extrabold tracking-tight text-white/95">
        Create account
      </h1>
      <p className="mt-0.5 text-center text-xs text-white/85">
        Join Cinéthos and start exploring instantly.
      </p>

      <div className="mt-2 grid grid-cols-1 gap-1.5">
        <Button
          fullWidth
          size="sm"
          className="justify-center bg-white text-gray-900 font-medium hover:brightness-95 h-8"
          radius="sm"
          onPress={() => redirectToOAuth("fortytwo")}
        >
          <FortyTwoIcon />
          <span className="ml-2">Continue with Intra</span>
        </Button>
        <Button
          fullWidth
          size="sm"
          className="justify-center bg-white text-gray-900 font-medium hover:brightness-95 h-8"
          radius="sm"
          onPress={() => redirectToOAuth("google")}
        >
          <GoogleIcon size={20} className="ml-2" />
          <span className="ml-2">Continue with Google</span>
        </Button>
        <Button
          fullWidth
          size="sm"
          className="justify-center bg-white text-gray-900 font-medium hover:brightness-95 h-8"
          radius="sm"
          onPress={() => redirectToOAuth("github")}
        >
          <GithubIcon size={20} className="ml-2" />
          <span className="ml-2">Continue with Github</span>
        </Button>
      </div>

      <div className="my-2 flex items-center gap-3 text-xs text-white/70">
        <div className="h-px w-full bg-white/20" />
        <span>or</span>
        <div className="h-px w-full bg-white/20" />
      </div>

      <div className="flex flex-col items-center gap-1 mb-2">
        <div className="relative">
          <div className="w-16 h-16 rounded-full overflow-hidden bg-white/10 border-2 border-white/20 flex items-center justify-center">
            {avatarPreview ? (
              <img
                src={avatarPreview}
                alt="Avatar preview"
                className="w-full h-full object-cover"
              />
            ) : (
              <svg
                className="w-8 h-8 text-white/40"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                />
              </svg>
            )}
          </div>
          <label
            htmlFor="avatar-upload"
            className="absolute -bottom-0.5 -right-0.5 w-6 h-6 bg-gradient-to-r from-indigo-500 to-pink-500 rounded-full flex items-center justify-center cursor-pointer hover:brightness-110 transition shadow-lg"
          >
            <svg
              className="w-3 h-3 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
          </label>
          <input
            id="avatar-upload"
            type="file"
            accept="image/*"
            onChange={handleAvatarChange}
            className="hidden"
          />
        </div>
        <p className="text-xs text-white/60">
          Upload profile picture (optional)
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-2">
        <div className="grid grid-cols-2 gap-2">
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
          color="primary"
          isLoading={isLoading}
        >
          Create Account
        </Button>

        <p className="mt-4 text-center text-xs text-white/75">
          Already have an account?{" "}
          <Link
            href="/auth/login"
            className="text-primary hover:text-pink-200 hover:underline"
          >
            Log in
          </Link>
        </p>
      </form>
    </div>
  );
}
