"use client";

import Link from "next/link";
import { Input } from "@heroui/input";
import { Button } from "@heroui/button";

function GoogleIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5" aria-hidden>
      <path fill="#EA4335" d="M12 10.2v3.9h5.5c-.24 1.4-1.66 4.1-5.5 4.1a6.4 6.4 0 1 1 0-12.8 5.7 5.7 0 0 1 4.03 1.58l2.74-2.74A9.6 9.6 0 1 0 12 21.6c5.52 0 9.17-3.87 9.17-9.33 0-.62-.07-1.06-.16-1.53H12z"/>
    </svg>
  );
}
function GithubIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-5 w-5 fill-current" aria-hidden>
      <path d="M12 .5a11.5 11.5 0 0 0-3.64 22.43c.58.11.79-.25.79-.56v-2.18c-3.22.7-3.9-1.4-3.9-1.4-.53-1.34-1.29-1.7-1.29-1.7-1.06-.72.08-.7.08-.7 1.18.09 1.8 1.22 1.8 1.22 1.04 1.78 2.72 1.27 3.38.97.11-.75.41-1.27.75-1.56-2.57-.29-5.27-1.28-5.27-5.7 0-1.26.45-2.28 1.2-3.08-.12-.3-.52-1.52.11-3.16 0 0 .98-.31 3.2 1.18a11.1 11.1 0 0 1 5.83 0c2.22-1.49 3.2-1.18 3.2-1.18.63 1.64.23 2.86.11 3.16.75.8 1.2 1.82 1.2 3.08 0 4.43-2.71 5.4-5.29 5.69.42.36.8 1.07.8 2.16v3.21c0 .31.2.68.8.56A11.5 11.5 0 0 0 12 .5Z"/>
    </svg>
  );
}
function FortyTwoIcon() {
  return (
    <span className="inline-flex h-5 w-5 items-center justify-center rounded bg-black text-white text-[10px] font-bold">
      42
    </span>
  );
}

export default function LoginPage() {
  return (
    <div className="rounded-3xl border border-white/15 bg-white/10 p-8 text-white shadow-2xl backdrop-blur-md">
      <h1 className="text-4xl font-extrabold tracking-tight text-white/95">CINÉTHOS</h1>
      <p className="mt-2 text-sm text-white/85">Welcome back — your cinematic universe awaits.</p>

      

      <form className="space-y-5 pt-4">
        <Input
          type="email"
          label="Email"
          placeholder="you@example.com"
          variant="faded"
          radius="sm"
          isRequired
        />
        <Input
          type="password"
          label="Password"
          placeholder="••••••••"
          variant="faded"
          radius="sm"
          isRequired
        />

        <div className="flex items-center justify-between text-sm text-white/75">
          <Link href="/forgot-password" className="hover:text-white hover:underline">Forgot password?</Link>
        </div>

        <Button type="submit" radius="sm" fullWidth className="mt-1 bg-gradient-to-r from-indigo-500 to-pink-500 text-white font-semibold shadow-lg transition hover:brightness-110">
          Log In
        </Button>

        <p className="mt-4 text-center text-xs text-white/75">
          Don’t have an account?{" "}
          <Link href="/auth/register" className="text-pink-300 hover:text-pink-200 hover:underline">Sign up</Link>
        </p>
      </form>
      <div className="my-6 flex items-center gap-3 text-xs text-white/70">
        <div className="h-px w-full bg-white/20" />
        <span>or</span>
        <div className="h-px w-full bg-white/20" />
      </div>
      <div className="mt-6 grid grid-cols-1 gap-2">
        <Button as={Link} href="/oauth/42" radius="sm" fullWidth className="justify-center bg-white text-gray-900 font-medium hover:brightness-95">
          <FortyTwoIcon />
          <span className="ml-2">Continue with 42</span>
        </Button>
        <Button as={Link} href="/oauth/google" radius="sm" fullWidth className="justify-center bg-white text-gray-900 font-medium hover:brightness-95">
          <GoogleIcon />
          <span className="ml-2">Continue with Google</span>
        </Button>
        <Button as={Link} href="/oauth/github" radius="sm" fullWidth className="justify-center bg-white text-gray-900 font-medium hover:brightness-95">
          <GithubIcon />
          <span className="ml-2">Continue with GitHub</span>
        </Button>
      </div>

    </div>
  );
}
