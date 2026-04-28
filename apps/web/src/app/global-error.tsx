'use client';

import { useEffect } from 'react';
import {
  GLOBAL_ERROR_RELOAD_DELAY_MS,
  GLOBAL_ERROR_RELOAD_LIMIT,
  recordGlobalErrorReload,
  shouldScheduleGlobalErrorReload,
} from './global-error-reload';

// Only retry a narrow class of transient RSC failures, and cap retries so
// persistent server errors don't get trapped in a reload loop.
export default function GlobalError({
  error,
}: {
  error: Error & { digest?: string };
}) {
  useEffect(() => {
    const pathname = window.location.pathname;
    const storage = window.sessionStorage;
    const now = Date.now();

    if (
      !shouldScheduleGlobalErrorReload({
        error,
        now,
        pathname,
        storage,
      })
    ) {
      return;
    }

    recordGlobalErrorReload({ now, pathname, storage });

    const timeoutId = window.setTimeout(() => {
      window.location.reload();
    }, GLOBAL_ERROR_RELOAD_DELAY_MS);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [error]);

  return (
    <html>
      <body>
        <p>
          We hit a temporary loading error. Waiting at least{' '}
          {GLOBAL_ERROR_RELOAD_DELAY_MS / 1000}s before retrying.
        </p>
        <p>Automatic retries are limited to {GLOBAL_ERROR_RELOAD_LIMIT}.</p>
        <button onClick={() => window.location.reload()} type="button">
          Retry now
        </button>
      </body>
    </html>
  );
}
