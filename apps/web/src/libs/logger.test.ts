import { vi, describe, it, expect, beforeEach } from 'vitest';

vi.mock('next/cache', () => ({ unstable_noStore: vi.fn() }));

import { caller } from './logger';

describe('caller', () => {
  it('extracts filename and lineNumber from a standard V8 stack frame', () => {
    const err = new Error();
    err.stack = [
      'Error',
      '    at Object.<anonymous> (/src/libs/logger.ts:42:10)',
      '    at Module._compile (node:internal/modules/cjs/loader:1256:14)',
    ].join('\n');

    const result = caller(err);

    expect(result).toEqual({ filename: '/src/libs/logger.ts', lineNumber: 42 });
  });

  it('returns undefined when error.stack is undefined', () => {
    const err = new Error();
    err.stack = undefined;

    const result = caller(err);

    expect(result).toBeUndefined();
  });

  it('returns undefined when no stack line matches the expected pattern', () => {
    const err = new Error();
    err.stack = 'Error: test\n    at anonymous\n    at internal';

    const result = caller(err);

    expect(result).toBeUndefined();
  });

  it('strips the rsc)/ prefix from Next.js RSC stack frames', () => {
    const err = new Error();
    // Stack frame format: (file (rsc)/path:line:col)
    // The regex captures everything after the second '(' as the filename,
    // so the captured group is "rsc)/src/server/action.ts" → strip "rsc)/" → "src/server/action.ts"
    err.stack = [
      'Error',
      '    at eval (file (rsc)/src/server/action.ts:10:15)',
    ].join('\n');

    const result = caller(err);

    expect(result?.filename).toBe('src/server/action.ts');
    expect(result?.lineNumber).toBe(10);
  });

  it('uses the first matching stack frame when multiple frames are present', () => {
    const err = new Error();
    err.stack = [
      'Error',
      '    at firstFrame (/first.ts:1:1)',
      '    at secondFrame (/second.ts:2:2)',
    ].join('\n');

    const result = caller(err);

    expect(result?.filename).toBe('/first.ts');
    expect(result?.lineNumber).toBe(1);
  });

  it('skips the Error message line (slice(1))', () => {
    const err = new Error();
    // The first line "Error: /tricky.ts:99:1" should be ignored because
    // caller() calls stack.split('\n').slice(1), removing the first line.
    err.stack = [
      'Error: /tricky.ts:99:1)',
      '    at realFrame (/real.ts:5:10)',
    ].join('\n');

    const result = caller(err);

    expect(result?.filename).toBe('/real.ts');
    expect(result?.lineNumber).toBe(5);
  });

  it('parses lineNumber as an integer', () => {
    const err = new Error();
    err.stack = 'Error\n    at fn (/file.ts:123:45)';

    const result = caller(err);

    expect(result?.lineNumber).toBe(123);
    expect(typeof result?.lineNumber).toBe('number');
  });

  it('returns the first matching stack frame even when deeper frames also match', () => {
    const err = new Error();
    // "Promise.all (index 0)" does not match (no :digit:digit pattern)
    // but the very next line does — it wins as the first match
    err.stack = [
      'Error',
      '    at async Promise.all (index 0)',
      '    at firstMatch (/first-match.ts:10:5)',
      '    at secondMatch (/second-match.ts:20:3)',
    ].join('\n');

    const result = caller(err);

    expect(result?.filename).toBe('/first-match.ts');
    expect(result?.lineNumber).toBe(10);
  });
});
