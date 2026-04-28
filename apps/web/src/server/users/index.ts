import { z } from 'zod';
import { publicProcedure, router } from '@/server/trpc';
import { auth0 } from '@/libs/auth0';
import { getAPIClient, handleAPIError } from '..';
import { headAuthenticatedUser } from './headAuthenticatedUser';
import { logTrpc } from '@/server/trpc';
import log from '@/libs/logger';

export const usersRouter = router({
  get: publicProcedure.input(z.string()).query(async (opts) => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.users.getApiUsers1(opts.input);
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), opts.input, s, 'users.get');
    }
  }),
  getAuthenticated: publicProcedure.query(async () => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.users.getApiUser();
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), undefined, s, 'users.getAuthenticated');
    }
  }),
  getKitchenPass: publicProcedure.query(async () => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      log().Info(new Error(), [
        { k: 'event', v: 'users_getKitchenPass_request' },
      ]);
      const result = await api.users.getApiUserKitchenPass();
      log().Info(new Error(), [
        { k: 'event', v: 'users_getKitchenPass_response' },
        { k: 'enabled', v: !!result?.enabled },
        { k: 'hasCopyText', v: !!result?.copyText },
        { k: 'hasCreatedDate', v: !!result?.createdDate },
      ]);
      return result;
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), undefined, s, 'users.getKitchenPass');
    }
  }),
  exists: publicProcedure.query(async () => {
    const s = performance.now();

    try {
      const { token } = await auth0.getAccessToken();
      return await headAuthenticatedUser(token);
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), undefined, s, 'users.exists');
    }
  }),
  create: publicProcedure
    .input(
      z.object({
        username: z.string().trim(),
        displayName: z.string().trim(),
        email: z.string().email().trim(),
      })
    )
    .mutation(async (opts) => {
      const api = await getAPIClient();
      const s = performance.now();

      try {
        return await api.users.postApiUser(opts.input);
      } catch (e) {
        handleAPIError(e);
      } finally {
        logTrpc(new Error(), opts.input, s, 'users.exists');
      }
    }),
  getUsername: publicProcedure
    .input(
      z.object({
        username: z.string().trim(),
      })
    )
    .mutation(async (opts) => {
      const api = await getAPIClient();
      const s = performance.now();

      try {
        return await api.users.postApiUsersUsername(opts.input);
      } catch (e) {
        handleAPIError(e);
      } finally {
        logTrpc(new Error(), opts.input, s, 'users.getUsername');
      }
    }),
  update: publicProcedure
    .input(
      z.object({
        username: z.string().trim(),
      })
    )
    .mutation(async (opts) => {
      const api = await getAPIClient();

      const s = performance.now();

      try {
        return await api.users.patchApiUser(opts.input);
      } catch (e) {
        handleAPIError(e);
      } finally {
      logTrpc(new Error(), opts.input, s, 'users.update');
    }
  }),
  createKitchenPass: publicProcedure.mutation(async () => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.users.postApiUserKitchenPass();
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), undefined, s, 'users.createKitchenPass');
    }
  }),
  deleteKitchenPass: publicProcedure.mutation(async () => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.users.deleteApiUserKitchenPass();
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), undefined, s, 'users.deleteKitchenPass');
    }
  }),
});

export type UsersRouter = typeof usersRouter;
