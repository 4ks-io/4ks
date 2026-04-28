'use client';

import { useEffect } from 'react';

// When a client-side RSC navigation fails (e.g. cold-start stream error), force a
// full-page reload. The browser retries as a plain HTML request, which always succeeds.
export default function GlobalError({}: { error: Error & { digest?: string } }) {
  useEffect(() => {
    window.location.reload();
  }, []);

  return (
    <html>
      <body />
    </html>
  );
}
