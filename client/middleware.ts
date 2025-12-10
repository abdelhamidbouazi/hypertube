import type { NextRequest } from "next/server";

import { NextResponse } from "next/server";

export function middleware(request: NextRequest) {
  const token = request.cookies.get("token")?.value;
  const { pathname } = request.nextUrl;

  const publicRoutes = [
    "/",
    "/auth/login",
    "/auth/register",
    "/forgot-password",
    "/reset_password",
  ];

  const isPublicRoute = publicRoutes.some((route) => pathname === route);
  const isPublicAppRoute = pathname === "/app/discover" || 
    (pathname.startsWith("/app/movie/") && !pathname.includes("/watch"));
  const isWatchRoute = pathname.startsWith("/app/movie/") && pathname.includes("/watch");
  const isProtectedRoute = pathname.startsWith("/app") && !isPublicAppRoute;

  if (isPublicRoute && token) {
    return NextResponse.redirect(new URL("/app/discover", request.url));
  }

  if (isWatchRoute && !token) {
    return NextResponse.redirect(new URL("/auth/login", request.url));
  }

  if (isProtectedRoute && !token) {
    return NextResponse.redirect(new URL("/auth/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico|.*\\..*).*)"],
};
