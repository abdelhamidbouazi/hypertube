"use client";

import Link from "next/link";
import Image from "next/image";
import { useState } from "react";
import { Input } from "@heroui/input";
import { Button } from "@heroui/button";
import { addToast } from "@heroui/toast";

import { forgotPassword } from "@/lib/hooks";
import { getErrorMessage } from "@/lib/error-utils";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isSent, setIsSent] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      await forgotPassword(email);
      setIsSent(true);
      addToast({
        title: "Email sent",
        description: "Check your inbox for the reset link.",
        severity: "success",
        timeout: 5000,
      });
    } catch (error) {
      const errorMessage = getErrorMessage(error);

      addToast({
        title: "Request failed",
        description: errorMessage,
        severity: "danger",
        timeout: 4000,
      });
    } finally {
      setIsLoading(false);
    }
  };

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
              Reset Password
            </h1>
            <p className="mt-2 text-sm text-white/85 text-center">
              Enter your email to receive a reset link.
            </p>

            {isSent ? (
              <div className="mt-6 text-center space-y-4">
                <div className="p-4 bg-green-500/20 border border-green-500/30 rounded-lg text-green-200">
                  <p>
                    We have sent a password reset link to{" "}
                    <strong>{email}</strong>.
                  </p>
                </div>
                <p className="text-sm text-white/70">
                  Didn&apos;t receive it? Check your spam folder or{" "}
                  <button
                    onClick={() => setIsSent(false)}
                    className="text-pink-300 hover:text-pink-200 hover:underline"
                  >
                    try again
                  </button>
                  .
                </p>
              </div>
            ) : (
              <form className="space-y-5 pt-6" onSubmit={handleSubmit}>
                <Input
                  isRequired
                  classNames={{
                    input: "!text-white",
                    label: "!text-white/90",
                  }}
                  label="Email"
                  placeholder="you@example.com"
                  radius="sm"
                  type="email"
                  value={email}
                  variant="faded"
                  onChange={(e) => setEmail(e.target.value)}
                />

                <Button
                  fullWidth
                  className="mt-1 bg-gradient-to-r from-indigo-500 to-pink-500 text-white font-semibold shadow-lg transition hover:brightness-110"
                  isLoading={isLoading}
                  radius="sm"
                  type="submit"
                >
                  Send Reset Link
                </Button>
              </form>
            )}

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
