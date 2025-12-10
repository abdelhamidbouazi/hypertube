"use client";

import React from "react";
import Link from "next/link";
import { siteConfig } from "@/config/site";
import { GithubIcon, TwitterIcon, Logo } from "@/components/icons";

export default function Footer() {
  const year = new Date().getFullYear();

  return (
    <footer className="w-full bg-transparent mt-8 md:mt-12">
      <div className="mx-auto w-full max-w-7xl px-4 sm:px-6">
        <div className="rounded-2xl border border-default-200/40 bg-content1/60 supports-[backdrop-filter]:backdrop-blur-xl shadow-lg shadow-black/5">
          <div className="grid gap-8 p-6 sm:p-8 md:grid-cols-4">
          <div className="col-span-2 md:col-span-1">
            <Link href="/" className="flex items-center gap-2 text-current">
              <Logo size={28} />
              <span className="text-base font-semibold">{siteConfig.name}</span>
            </Link>
            <p className="mt-3 max-w-sm text-sm text-default-600">
              {siteConfig.description}
            </p>
            <div className="mt-4 flex items-center gap-3 text-default-600">
              {siteConfig.links.github && (
                <Link
                  href={siteConfig.links.github}
                  aria-label="GitHub"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-default-500 hover:text-foreground transition-colors"
                >
                  <GithubIcon size={22} />
                </Link>
              )}
              {siteConfig.links.X && (
                <Link
                  href={siteConfig.links.X}
                  aria-label="X"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-default-500 hover:text-foreground transition-colors"
                >
                  <TwitterIcon size={22} />
                </Link>
              )}
              {siteConfig.links.porfolio && (
                <Link
                  href={siteConfig.links.porfolio}
                  aria-label="Portfolio"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-default-500 hover:text-foreground transition-colors"
                >
                  Portfolio
                </Link>
              )}
            </div>
          </div>
          <div>
            <h4 className="text-sm font-semibold text-foreground">Explore</h4>
            <ul className="mt-3 space-y-2 text-sm text-default-600">
              {siteConfig.navItems.map((item) => (
                <li key={item.href}>
                  <Link href={item.href} className="hover:text-foreground">
                    {item.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
          <div>
            <h4 className="text-sm font-semibold text-foreground">Resources</h4>
            <ul className="mt-3 space-y-2 text-sm text-default-600">
              <li>
                <Link href="/app/settings" className="hover:text-foreground">
                  Settings
                </Link>
              </li>
              <li>
                <Link
                  href="https://github.com/abdelhamidbouazi/hypertube"
                  className="hover:text-foreground"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  GitHub Repo
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <h4 className="text-sm font-semibold text-foreground">Contact</h4>
            <ul className="mt-3 space-y-2 text-sm text-default-600">
              {siteConfig.links.porfolio && (
                <li>
                  <Link
                    href={siteConfig.links.porfolio}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="hover:text-foreground"
                  >
                    abouazi.me
                  </Link>
                </li>
              )}
              {siteConfig.links.X && (
                <li>
                  <Link
                    href={siteConfig.links.X}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="hover:text-foreground"
                  >
                    Follow on X
                  </Link>
                </li>
              )}
            </ul>
          </div>
          </div>

          <div className="flex flex-col items-center justify-between gap-2 border-t border-default-200/40 p-4 text-xs text-default-600 md:flex-row">
            <p>© {year} {siteConfig.name}. All rights reserved.</p>
            <p className="text-default-500">
              Built by
              <span className="mx-1 text-primary">abouazi</span>
              <span className="mx-1 text-default-400">·</span>
              <span className="mx-1 text-primary">anaji-el</span>
              <span className="mx-1 text-default-400">·</span>
              <span className="mx-1 text-primary">stamim</span>
              <span className="mx-1 text-default-400">·</span>
              <span className="mx-1 text-primary">amessah</span>
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
}
