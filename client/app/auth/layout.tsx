"use client";

import Image from "next/image";
import Footer from "@/components/footer";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
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
        <div className="w-[90%] max-w-md">{children}</div>
      </section>

      {/* <div className="absolute inset-x-0 bottom-0 z-30 ">
        <Footer />
      </div> */}

      <style jsx global>{`
        @keyframes kenburns {
          0% { transform: scale(1.05) translate3d(0,0,0); }
          100% { transform: scale(1.12) translate3d(2%,2%,0); }
        }
        .animate-kenburns { animation: kenburns 22s ease-in-out infinite alternate; }
      `}</style>
    </div>
  );
}
