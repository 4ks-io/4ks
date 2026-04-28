import { describe, expect, it } from 'vitest';
import {
  clearGlobalErrorReloadState,
  GLOBAL_ERROR_RELOAD_DELAY_MS,
  GLOBAL_ERROR_RELOAD_LIMIT,
  GLOBAL_ERROR_RELOAD_STORAGE_KEY,
  isLikelyTransientRSCError,
  parseGlobalErrorReloadState,
  readGlobalErrorReloadState,
  recordGlobalErrorReload,
  shouldScheduleGlobalErrorReload,
} from './global-error-reload';

function createStorage() {
  const data = new Map<string, string>();

  return {
    getItem(key: string) {
      return data.get(key) ?? null;
    },
    setItem(key: string, value: string) {
      data.set(key, value);
    },
    removeItem(key: string) {
      data.delete(key);
    },
  };
}

function makeError(message: string) {
  return new Error(message) as Error & { digest?: string };
}

describe('isLikelyTransientRSCError', () => {
  it('matches known server component render failures', () => {
    expect(
      isLikelyTransientRSCError(
        makeError('An error occurred in the Server Components render')
      )
    ).toBe(true);
  });

  it('matches failed RSC payload fetches', () => {
    expect(
      isLikelyTransientRSCError(makeError('Failed to fetch RSC payload'))
    ).toBe(true);
  });

  it('does not match unrelated errors', () => {
    expect(isLikelyTransientRSCError(makeError('Database unavailable'))).toBe(
      false
    );
  });
});

describe('parseGlobalErrorReloadState', () => {
  it('returns null for invalid JSON', () => {
    expect(parseGlobalErrorReloadState('{')).toBeNull();
  });

  it('returns null for malformed payloads', () => {
    expect(
      parseGlobalErrorReloadState(
        JSON.stringify({ count: '1', scheduledAt: 1, pathname: '/' })
      )
    ).toBeNull();
  });

  it('returns parsed state for valid payloads', () => {
    expect(
      parseGlobalErrorReloadState(
        JSON.stringify({ count: 1, scheduledAt: 123, pathname: '/' })
      )
    ).toEqual({ count: 1, scheduledAt: 123, pathname: '/' });
  });
});

describe('global error reload policy', () => {
  it('schedules the first transient RSC retry', () => {
    const storage = createStorage();

    expect(
      shouldScheduleGlobalErrorReload({
        error: makeError('An error occurred in the Server Components render'),
        now: 1_000,
        pathname: '/',
        storage,
      })
    ).toBe(true);
  });

  it('records retries against the current pathname', () => {
    const storage = createStorage();

    recordGlobalErrorReload({ now: 1_000, pathname: '/', storage });

    expect(readGlobalErrorReloadState(storage)).toEqual({
      count: 1,
      scheduledAt: 1_000,
      pathname: '/',
    });
  });

  it('does not schedule a second auto-reload once the limit is reached', () => {
    const storage = createStorage();

    recordGlobalErrorReload({ now: 1_000, pathname: '/', storage });

    expect(GLOBAL_ERROR_RELOAD_LIMIT).toBe(1);
    expect(
      shouldScheduleGlobalErrorReload({
        error: makeError('Failed to fetch RSC payload'),
        now: 1_000 + GLOBAL_ERROR_RELOAD_DELAY_MS,
        pathname: '/',
        storage,
      })
    ).toBe(false);
  });

  it('resets the budget when the pathname changes', () => {
    const storage = createStorage();

    recordGlobalErrorReload({ now: 1_000, pathname: '/', storage });

    expect(
      shouldScheduleGlobalErrorReload({
        error: makeError('Failed to fetch RSC payload'),
        now: 1_000 + GLOBAL_ERROR_RELOAD_DELAY_MS,
        pathname: '/explore',
        storage,
      })
    ).toBe(true);
  });

  it('clears stale retry state for unrelated errors', () => {
    const storage = createStorage();

    recordGlobalErrorReload({ now: 1_000, pathname: '/', storage });

    expect(
      shouldScheduleGlobalErrorReload({
        error: makeError('Database unavailable'),
        now: 2_000,
        pathname: '/',
        storage,
      })
    ).toBe(false);
    expect(storage.getItem(GLOBAL_ERROR_RELOAD_STORAGE_KEY)).toBeNull();
  });

  it('can clear stored state explicitly', () => {
    const storage = createStorage();

    recordGlobalErrorReload({ now: 1_000, pathname: '/', storage });
    clearGlobalErrorReloadState(storage);

    expect(readGlobalErrorReloadState(storage)).toBeNull();
  });
});
