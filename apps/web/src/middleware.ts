import type { NextRequest } from 'next/server';
import { NextResponse } from 'next/server';
import { auth0 } from '@/libs/auth0';

export async function middleware(request: NextRequest) {
  const requestHeaders = new Headers(request.headers);
  requestHeaders.set(
    'x-url-pathname',
    `${request.nextUrl.pathname}${request.nextUrl.search}`
  );

  const authResponse = await auth0.middleware(request);
  const pathname = request.nextUrl.pathname;
  const authRoute =
    pathname.startsWith('/auth') ||
    pathname.startsWith('/me/') ||
    pathname.startsWith('/my-org/') ||
    authResponse.headers.has('location');

  if (authRoute) {
    return authResponse;
  }

  const response = NextResponse.next({
    request: {
      headers: requestHeaders,
    },
  });

  authResponse.headers.forEach((value, key) => {
    response.headers.set(key, value);
  });

  for (const cookie of authResponse.cookies.getAll()) {
    response.cookies.set(cookie);
  }

  return response;
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|sitemap.xml|robots.txt|static|.*\\..*).*)',
  ],
};
