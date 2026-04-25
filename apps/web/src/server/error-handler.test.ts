import { vi, describe, it, expect, beforeEach } from 'vitest';

// Must be mocked before the module under test is imported so that transitive
// Next.js server imports don't blow up in the Node test environment.
vi.mock('next/cache', () => ({ unstable_noStore: vi.fn() }));
vi.mock('next/navigation', () => ({ redirect: vi.fn() }));

import { TRPCError } from '@trpc/server';
import { handleAPIError, HttpStatusCode } from './error-handler';
import type { APIError } from './error-handler';

// ---------------------------------------------------------------------------
// Helper – call handleAPIError and catch the thrown TRPCError.
// handleAPIError always throws, so we capture it here.
// ---------------------------------------------------------------------------
function callAndCatch(apiErr: APIError): TRPCError {
  try {
    handleAPIError(apiErr);
  } catch (e) {
    return e as TRPCError;
  }
  throw new Error('handleAPIError should have thrown but did not');
}

function makeApiError(status: number, statusText = 'Status text', body = ''): APIError {
  return { url: 'https://api.example.com/test', status, statusText, body };
}

// ---------------------------------------------------------------------------
// HttpStatusCode mapping table
// ---------------------------------------------------------------------------

describe('HttpStatusCode', () => {
  const mapping: [number, string][] = [
    [400, 'BAD_REQUEST'],
    [401, 'UNAUTHORIZED'],
    [403, 'FORBIDDEN'],
    [404, 'NOT_FOUND'],
    [405, 'METHOD_NOT_SUPPORTED'],
    [408, 'TIMEOUT'],
    [409, 'CONFLICT'],
    [412, 'PRECONDITION_FAILED'],
    [413, 'PAYLOAD_TOO_LARGE'],
    [422, 'UNPROCESSABLE_CONTENT'],
    [429, 'TOO_MANY_REQUESTS'],
    [499, 'CLIENT_CLOSED_REQUEST'],
    [500, 'INTERNAL_SERVER_ERROR'],
  ];

  it.each(mapping)('HTTP %i maps to "%s"', (status, code) => {
    expect(HttpStatusCode[status]).toBe(code);
  });

  it('covers 13 distinct HTTP status codes', () => {
    expect(mapping.length).toBe(13);
    const uniqueCodes = new Set(mapping.map(([, code]) => code));
    expect(uniqueCodes.size).toBe(13);
  });
});

// ---------------------------------------------------------------------------
// handleAPIError behaviour
// ---------------------------------------------------------------------------

describe('handleAPIError', () => {
  it('always throws', () => {
    expect(() => handleAPIError(makeApiError(404))).toThrow();
  });

  it('throws a TRPCError', () => {
    expect(() => handleAPIError(makeApiError(404))).toThrow(TRPCError);
  });

  describe('error code mapping', () => {
    it('maps HTTP 400 → BAD_REQUEST', () => {
      expect(callAndCatch(makeApiError(400)).code).toBe('BAD_REQUEST');
    });

    it('maps HTTP 401 → UNAUTHORIZED', () => {
      expect(callAndCatch(makeApiError(401)).code).toBe('UNAUTHORIZED');
    });

    it('maps HTTP 403 → FORBIDDEN', () => {
      expect(callAndCatch(makeApiError(403)).code).toBe('FORBIDDEN');
    });

    it('maps HTTP 404 → NOT_FOUND', () => {
      expect(callAndCatch(makeApiError(404)).code).toBe('NOT_FOUND');
    });

    it('maps HTTP 405 → METHOD_NOT_SUPPORTED', () => {
      expect(callAndCatch(makeApiError(405)).code).toBe('METHOD_NOT_SUPPORTED');
    });

    it('maps HTTP 408 → TIMEOUT', () => {
      expect(callAndCatch(makeApiError(408)).code).toBe('TIMEOUT');
    });

    it('maps HTTP 409 → CONFLICT', () => {
      expect(callAndCatch(makeApiError(409)).code).toBe('CONFLICT');
    });

    it('maps HTTP 412 → PRECONDITION_FAILED', () => {
      expect(callAndCatch(makeApiError(412)).code).toBe('PRECONDITION_FAILED');
    });

    it('maps HTTP 413 → PAYLOAD_TOO_LARGE', () => {
      expect(callAndCatch(makeApiError(413)).code).toBe('PAYLOAD_TOO_LARGE');
    });

    it('maps HTTP 422 → UNPROCESSABLE_CONTENT', () => {
      expect(callAndCatch(makeApiError(422)).code).toBe('UNPROCESSABLE_CONTENT');
    });

    it('maps HTTP 429 → TOO_MANY_REQUESTS', () => {
      expect(callAndCatch(makeApiError(429)).code).toBe('TOO_MANY_REQUESTS');
    });

    it('maps HTTP 499 → CLIENT_CLOSED_REQUEST', () => {
      expect(callAndCatch(makeApiError(499)).code).toBe('CLIENT_CLOSED_REQUEST');
    });

    it('maps HTTP 500 → INTERNAL_SERVER_ERROR', () => {
      expect(callAndCatch(makeApiError(500)).code).toBe('INTERNAL_SERVER_ERROR');
    });
  });

  describe('error message and cause', () => {
    it('sets the TRPCError message to the statusText of the API error', () => {
      const err = callAndCatch(makeApiError(404, 'Recipe not found'));
      expect(err.message).toBe('Recipe not found');
    });

    it('sets the TRPCError cause from the response body of the API error', () => {
      // TRPCError wraps a string cause in an Error instance
      const err = callAndCatch(makeApiError(422, 'Invalid', 'field validation failed'));
      expect(err.cause).toBeInstanceOf(Error);
      expect((err.cause as Error).message).toBe('field validation failed');
    });

    it('is an instance of both TRPCError and Error', () => {
      const err = callAndCatch(makeApiError(500));
      expect(err).toBeInstanceOf(TRPCError);
      expect(err).toBeInstanceOf(Error);
    });
  });
});
