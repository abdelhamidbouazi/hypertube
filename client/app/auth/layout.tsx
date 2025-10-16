"use client";
/* eslint-disable */

import Image from "next/image";
import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { getAccessToken } from "@/lib/auth";

import Footer from "@/components/footer";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();

  useEffect(() => {
    const t = getAccessToken();
    if (t) router.replace("/app/discover");
  }, [router]);

  return (
    <div className="relative h-screen w-full overflow-hidden">
      <div className="absolute inset-0 z-0">
        <Image
          fill
          priority
          alt="Cinematic background"
          className="object-cover animate-kenburns will-change-transform"
          src="/movies-bg.jpg"
        />
      </div>

      <div className="absolute inset-0 z-10 bg-gradient-to-b from-black/65 via-black/50 to-black/70" />

      <section className="absolute inset-0 z-20 grid place-items-center px-4">
        <div className="w-[90%] max-w-md">{children}</div>
      </section>

      <div className="absolute inset-x-0 bottom-0 z-30 ">
        <Footer />
      </div>
    </div>
  );
}
