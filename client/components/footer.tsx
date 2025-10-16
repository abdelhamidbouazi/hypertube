// src/components/Footer.tsx
"use client";

import React from "react";
import Link from "next/link";
import {
  FaGithub,
  FaLinkedin,
  FaTwitter,
  FaYoutube,
  FaInstagram,
} from "react-icons/fa";

const NAV = [
  {
    title: "Explore",
    links: [
      { label: "Home", href: "/" },
      { label: "Library", href: "/library" },
      { label: "Profile", href: "/profile" },
      { label: "About", href: "/about" },
    ],
  },
  {
    title: "Support",
    links: [
      { label: "Help Center", href: "/help" },
      { label: "Report an Issue", href: "/support" },
      { label: "Contact", href: "/contact" },
      { label: "Status", href: "/status" },
    ],
  },
  {
    title: "Legal",
    links: [
      { label: "Privacy Policy", href: "/privacy" },
      { label: "Terms of Service", href: "/terms" },
      { label: "Cookies", href: "/cookies" },
      { label: "Licenses", href: "/licenses" },
    ],
  },
] as const;

const LANGS = [
  { value: "en", label: "English" },
  { value: "fr", label: "Français" },
  { value: "es", label: "Español" },
];

function SocialIcon({ name }: { name: string }) {
  switch (name) {
    case "GitHub":
      return <FaGithub aria-hidden="true" />;
    case "Twitter":
      return <FaTwitter aria-hidden="true" />;
    case "LinkedIn":
      return <FaLinkedin aria-hidden="true" />;
    case "YouTube":
      return <FaYoutube aria-hidden="true" />;
    case "Instagram":
      return <FaInstagram aria-hidden="true" />;
    default:
      return <FaGithub aria-hidden="true" />;
  }
}

export default function Footer() {
  const year = new Date().getFullYear();
  //   const year = '2025'

  return (
    <footer className="w-full flex items-center justify-center py-3">
      <Link
        // isExternal={true}
        className="flex items-center gap-1 text-current"
        href="https://github.com/abdelhamidbouazi/hypertube"
        title="Repository of the project in github"
      >
        <span className="text-default-600">
          © {year} CINÉTHOS — built with care by the team.
        </span>
        <p className="text-primary">abouazi</p>
        <span className="text-default-600">,</span>
        <p className="text-primary">anaji-el</p>
        <span className="text-default-600">,</span>
        <p className="text-primary">stamim</p>
        <span className="text-default-600">,</span>
        <p className="text-primary">amessah</p>
      </Link>
    </footer>
  );
}
