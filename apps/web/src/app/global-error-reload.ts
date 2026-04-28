export const GLOBAL_ERROR_RELOAD_STORAGE_KEY = 'global-error-reload';
export const GLOBAL_ERROR_RELOAD_LIMIT = 1;
export const GLOBAL_ERROR_RELOAD_DELAY_MS = 2_000;

export type GlobalErrorReloadState = {
  count: number;
  scheduledAt: number;
  pathname: string;
};

export type StorageLike = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>;

export function isLikelyTransientRSCError(error: Error & { digest?: string }) {
  const message = error.message.toLowerCase();

  return (
    message.includes('server components render') ||
    message.includes('failed to fetch rsc payload') ||
    message.includes('rsc payload')
  );
}

export function parseGlobalErrorReloadState(value: string | null) {
  if (!value) {
    return null;
  }

  try {
    const parsed = JSON.parse(value) as Partial<GlobalErrorReloadState>;

    if (
      typeof parsed.count !== 'number' ||
      typeof parsed.scheduledAt !== 'number' ||
      typeof parsed.pathname !== 'string'
    ) {
      return null;
    }

    return parsed satisfies GlobalErrorReloadState;
  } catch {
    return null;
  }
}

export function readGlobalErrorReloadState(storage: StorageLike) {
  return parseGlobalErrorReloadState(
    storage.getItem(GLOBAL_ERROR_RELOAD_STORAGE_KEY)
  );
}

export function writeGlobalErrorReloadState(
  storage: StorageLike,
  state: GlobalErrorReloadState
) {
  storage.setItem(GLOBAL_ERROR_RELOAD_STORAGE_KEY, JSON.stringify(state));
}

export function clearGlobalErrorReloadState(storage: StorageLike) {
  storage.removeItem(GLOBAL_ERROR_RELOAD_STORAGE_KEY);
}

export function shouldScheduleGlobalErrorReload({
  error,
  now,
  pathname,
  storage,
}: {
  error: Error & { digest?: string };
  now: number;
  pathname: string;
  storage: StorageLike;
}) {
  if (!isLikelyTransientRSCError(error)) {
    clearGlobalErrorReloadState(storage);
    return false;
  }

  const state = readGlobalErrorReloadState(storage);

  if (!state || state.pathname !== pathname) {
    return true;
  }

  if (state.count >= GLOBAL_ERROR_RELOAD_LIMIT) {
    return false;
  }

  return now - state.scheduledAt >= GLOBAL_ERROR_RELOAD_DELAY_MS;
}

export function recordGlobalErrorReload({
  now,
  pathname,
  storage,
}: {
  now: number;
  pathname: string;
  storage: StorageLike;
}) {
  const state = readGlobalErrorReloadState(storage);

  writeGlobalErrorReloadState(storage, {
    count: state?.pathname === pathname ? state.count + 1 : 1,
    scheduledAt: now,
    pathname,
  });
}
