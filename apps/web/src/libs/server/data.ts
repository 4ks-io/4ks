import { auth0 } from '@/libs/auth0';
import { serverClient } from '@/trpc/serverClient';
import { TRPCError } from '@trpc/server';
import { getHTTPStatusCodeFromError } from '@trpc/server/http';

// user
export async function getUserData() {
  const session = await auth0.getSession();

  if (!session) return undefined;

  try {
    return (await serverClient.users.getAuthenticated()) ?? undefined;
  } catch (e) {
    if (e instanceof TRPCError && getHTTPStatusCodeFromError(e) === 404) {
      return undefined;
    }
    // tr@ck: handle other errors?
  }
}
