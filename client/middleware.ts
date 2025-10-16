/* eslint-disable */
import { NextRequest, NextResponse } from "next/server";

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  const token = req.cookies.get("token")?.value;

  const isAuthPage = pathname.startsWith("/auth");
  const isProtected = pathname.startsWith("/app");
  const isHome = pathname === "/";

  if (isAuthPage && token) {
    const url = req.nextUrl.clone();
    url.pathname = "/app/discover";

    return NextResponse.redirect(url);
  }

  if (isProtected && !token) {
    const url = req.nextUrl.clone();

    url.pathname = "/auth/login";
    url.searchParams.set("next", pathname);

    return NextResponse.redirect(url);
  }

  if (isHome && token) {
    const url = req.nextUrl.clone();
    url.pathname = "/app/discover";

    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/", "/auth/:path*", "/app/:path*"],
};
