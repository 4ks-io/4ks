import { describe, expect, it } from 'vitest';
import {
  isSettledMutation,
  shouldHandleSettledMutation,
} from './mutation-state';

describe('isSettledMutation', () => {
  it('returns false while a mutation is still pending', () => {
    expect(
      isSettledMutation({
        isPending: true,
        isError: false,
        isSuccess: false,
      })
    ).toBe(false);
  });

  it('returns true when a mutation succeeds', () => {
    expect(
      isSettledMutation({
        isPending: false,
        isError: false,
        isSuccess: true,
      })
    ).toBe(true);
  });

  it('returns true when a mutation errors', () => {
    expect(
      isSettledMutation({
        isPending: false,
        isError: true,
        isSuccess: false,
      })
    ).toBe(true);
  });
});

describe('shouldHandleSettledMutation', () => {
  it('returns false when the caller has not enabled handling yet', () => {
    expect(
      shouldHandleSettledMutation(false, {
        isPending: false,
        isError: false,
        isSuccess: true,
      })
    ).toBe(false);
  });

  it('returns true when handling is enabled and the mutation is settled', () => {
    expect(
      shouldHandleSettledMutation(true, {
        isPending: false,
        isError: false,
        isSuccess: true,
      })
    ).toBe(true);
  });
});
