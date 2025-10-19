"use client";

import Footer from "@/components/footer";
import Image from "next/image";
import Link from "next/link";

export default function LandingPage() {
  return (
    <div className="relative h-screen w-full overflow-hidden">
      <section className="absolute inset-0 flex items-center justify-center">
        {/* BG */}
        <div className="absolute inset-0 z-0">
          <Image
            src="/movies-bg.jpg"
            alt="Cinematic background"
            fill
            priority
            className="object-cover animate-kenburns will-change-transform"
          />
        </div>

        {/* Overlay */}
        <div className="absolute inset-0 z-10 bg-gradient-to-b from-black/70 via-black/50 to-black/70" />

        {/* Card */}
        <div className="relative z-20 mx-auto w-[90%] max-w-xl rounded-3xl border border-white/15 bg-white/10 p-8 text-center text-white backdrop-blur-md shadow-2xl">
          <h1 className="text-4xl font-extrabold tracking-tight">CINÉTHOS</h1>
          <p className="mt-3 text-sm text-white/80">
            Your next cinematic experience awaits — discover, stream, and enjoy instantly.
          </p>

          <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link
              href="/auth/login"
              className="w-40 inline-flex items-center justify-center rounded-xl bg-white px-6 py-3 text-sm font-semibold text-gray-900 shadow-md transition hover:brightness-95"
            >
              Log In
            </Link>
            <Link
              href="/app/discover"
              className="w-40 inline-flex items-center justify-center rounded-xl border border-white/30 bg-white/10 px-6 py-3 text-sm font-semibold text-white/90 backdrop-blur-sm transition hover:bg-white/20"
            >
              Discover
            </Link>
          </div>
        </div>
      </section>

      <div className="absolute inset-x-0 bottom-0 z-30">
          <Footer />
        
      </div>

      {/* Keyframes */}
      <style jsx global>{`
        @keyframes kenburns {
          0% { transform: scale(1.05) translate3d(0, 0, 0); }
          100% { transform: scale(1.12) translate3d(2%, 2%, 0); }
        }
        .animate-kenburns {
          animation: kenburns 22s ease-in-out infinite alternate;
        }
      `}</style>
    </div>
  );
}
