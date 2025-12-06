"use client";

import Link from "next/link";
import Image from "next/image";
import { useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Input } from "@heroui/input";
import { Button } from "@heroui/button";
import { addToast } from "@heroui/toast";
import { Eye, EyeOff } from "lucide-react";

import { resetPassword } from "@/lib/hooks";
import { getErrorMessage } from "@/lib/error-utils";

function ResetPasswordForm() {
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isVisible, setIsVisible] = useState(false);
  const [isConfirmVisible, setIsConfirmVisible] = useState(false);

  const router = useRouter();
  const searchParams = useSearchParams();

  const token = searchParams.get("token");
  const email = searchParams.get("email");

  const toggleVisibility = () => setIsVisible(!isVisible);
  const toggleConfirmVisibility = () => setIsConfirmVisible(!isConfirmVisible);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!token || !email) {
      addToast({
        title: "Invalid Link",
        description: "Missing token or email. Please request a new link.",
        severity: "danger",
      });
      return;
    }

    if (password !== confirmPassword) {
      addToast({
        title: "Error",
        description: "Passwords do not match.",
        severity: "danger",
      });
      return;
    }

    setIsLoading(true);

    try {
      await resetPassword(token, email, password);

      addToast({
        title: "Success",
        description: "Password reset successfully. Please login.",
        severity: "success",
        timeout: 5000,
      });

      router.push("/auth/login");
    } catch (error) {
      const errorMessage = getErrorMessage(error);

      addToast({
        title: "Reset failed",
        description: errorMessage,
        severity: "danger",
        timeout: 4000,
      });
    } finally {
      setIsLoading(false);
    }
  };

  if (!token || !email) {
    return (
      <div className="text-center space-y-4">
        <div className="p-4 bg-red-500/20 border border-red-500/30 rounded-lg text-red-200">
          <p>Invalid or missing reset token.</p>
        </div>
        <Link
          className="block text-sm text-white/70 hover:text-white hover:underline"
          href="/forgot-password"
        >
          Request a new link
        </Link>
      </div>
    );
  }

  return (
    <form className="space-y-5 pt-6" onSubmit={handleSubmit}>
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
        label="New Password"
        placeholder="••••••••"
        radius="sm"
        type={isVisible ? "text" : "password"}
        value={password}
        variant="faded"
        onChange={(e) => setPassword(e.target.value)}
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
            onClick={toggleConfirmVisibility}
          >
            {isConfirmVisible ? (
              <EyeOff className="text-2xl text-default-400 pointer-events-none" />
            ) : (
              <Eye className="text-2xl text-default-400 pointer-events-none" />
            )}
          </button>
        }
        label="Confirm Password"
        placeholder="••••••••"
        radius="sm"
        type={isConfirmVisible ? "text" : "password"}
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
        Reset Password
      </Button>
    </form>
  );
}

export default function ResetPasswordPage() {
  return (
    <div className="relative h-screen w-full overflow-hidden">
      <div className="absolute inset-0 z-0">
        <Image
          src="/movies-bg.jpg"
          alt="Cinematic background"
          fill
          priority
          className="object-cover animate-kenburns will-change-transform"
        />
      </div>

      <div className="absolute inset-0 z-10 bg-gradient-to-b from-black/65 via-black/50 to-black/70" />

      <section className="absolute inset-0 z-20 grid place-items-center px-4">
        <div className="w-[90%] max-w-md">
          <div className="rounded-3xl border border-white/15 bg-white/10 p-8 text-white shadow-2xl backdrop-blur-md">
            <h1 className="text-4xl font-extrabold tracking-tight text-white/95 text-center">
              Set New Password
            </h1>
            <p className="mt-2 text-sm text-white/85 text-center">
              Create a strong password for your account.
            </p>

            <Suspense
              fallback={
                <div className="p-8 text-center text-white/50">Loading...</div>
              }
            >
              <ResetPasswordForm />
            </Suspense>

            <div className="mt-6 text-center">
              <Link
                className="text-sm text-white/70 hover:text-white hover:underline"
                href="/auth/login"
              >
                Back to Login
              </Link>
            </div>
          </div>
        </div>
      </section>

      <style jsx global>{`
        @keyframes kenburns {
          0% {
            transform: scale(1.05) translate3d(0, 0, 0);
          }
          100% {
            transform: scale(1.12) translate3d(2%, 2%, 0);
          }
        }
        .animate-kenburns {
          animation: kenburns 22s ease-in-out infinite alternate;
        }
      `}</style>
    </div>
  );
}
