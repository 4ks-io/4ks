import type { SessionData } from '@auth0/nextjs-auth0/types';
import { headers } from 'next/headers';
import { notFound, redirect } from 'next/navigation';
import { auth0 } from '@/libs/auth0';
import { serverClient } from '@/trpc/serverClient';
import log from '@/libs/logger';
import { models_User } from '@4ks/api-fetch';
import { buildAuthLoginPath, Page } from '../navigation';

export type UserSession = {
  user: models_User | undefined;
  session: SessionData | undefined;
  isAuthenticated: boolean;
  isRegistered: boolean;
};

export async function handleUserNavigation(page: Page): Promise<UserSession> {
  // Route protection remains explicit in server code. Middleware keeps the
  // session fresh and mounts /auth/*, while authenticated pages still redirect
  // here when a session is required.
  const session = await auth0.getSession();
  const requestHeaders = await headers();
  const currentPath = requestHeaders.get('x-url-pathname') ?? undefined;
  log().Debug(new Error(), [
    { k: 'page', v: page },
    { k: 'session', v: !!session },
  ]);

  if (!session) {
    // unauthenticated
    if ([Page.REGISTER, Page.AUTHENTICATED].includes(page)) {
      redirect(buildAuthLoginPath(currentPath));
    }

    // anonymous
    return {
      user: undefined,
      session: undefined,
      isAuthenticated: false,
      isRegistered: false,
    };
  }

  // check user exists
  const data = await serverClient.users.exists();

  // handle error / trpc should have crash before this
  if (!data?.Status || ![200, 204].includes(data?.Status)) {
    return notFound();
    // tr@ck: retry or return unexpected error page
  }

  // authenticatd but not registered
  if (data?.Status == 204) {
    if (page == Page.REGISTER) {
      return {
        user: undefined,
        session: session,
        isAuthenticated: true,
        isRegistered: false,
      };
    }
    redirect('/register');
  }

  // authenticated and registered
  if (data?.Status == 200 && page == Page.REGISTER) {
    redirect('/');
  }

  // authenticated and registered
  return {
    user: await serverClient.users.getAuthenticated(),
    session: session,
    isAuthenticated: true,
    isRegistered: true,
  };
}
