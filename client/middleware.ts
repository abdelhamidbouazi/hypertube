import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  
  // Get token from cookies
  const token = request.cookies.get('token')?.value;

  // Protected routes that require authentication
  const isProtectedRoute = pathname.startsWith('/app');
  
  // Auth routes that should redirect to /app if already authenticated
  const isAuthRoute = pathname.startsWith('/auth');
  
  // OAuth routes should be accessible without authentication
  const isOAuthRoute = pathname.startsWith('/oauth');

  // Allow OAuth routes to pass through
  if (isOAuthRoute) {
    return NextResponse.next();
  }

  // If accessing protected route without token, redirect to login
  if (isProtectedRoute && !token) {
    const loginUrl = new URL('/auth/login', request.url);
    loginUrl.searchParams.set('redirect', pathname);
    return NextResponse.redirect(loginUrl);
  }

  // If accessing auth route with token, redirect to app
  if (isAuthRoute && token) {
    return NextResponse.redirect(new URL('/app/discover', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (images, etc.)
     */
    '/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico|css|js)$).*)',
  ],
};

