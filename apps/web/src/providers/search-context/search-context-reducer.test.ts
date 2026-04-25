import { describe, it, expect } from 'vitest';
import { searchContextReducer, SearchContextAction } from './search-context-reducer';
import type { SearchContextState } from './search-context-types';

const baseState: SearchContextState = {
  client: null,
  open: false,
  handleOpen: () => {},
  handleClose: () => {},
  results: [],
  setResults: () => {},
  clearResults: () => {},
  clear: () => {},
  value: '',
  setValue: () => {},
};

describe('searchContextReducer', () => {
  describe('INIT', () => {
    it('merges payload fields into the current state', () => {
      const mockClient = { search: () => {} };
      const result = searchContextReducer(baseState, {
        type: SearchContextAction.INIT,
        payload: { client: mockClient, open: true },
      });
      expect(result.client).toBe(mockClient);
      expect(result.open).toBe(true);
    });

    it('preserves fields not included in the payload', () => {
      const result = searchContextReducer(
        { ...baseState, value: 'tacos' },
        { type: SearchContextAction.INIT, payload: { open: true } }
      );
      expect(result.value).toBe('tacos');
    });

    it('allows overwriting an existing field', () => {
      const result = searchContextReducer(
        { ...baseState, value: 'pizza' },
        { type: SearchContextAction.INIT, payload: { value: 'pasta' } }
      );
      expect(result.value).toBe('pasta');
    });
  });

  describe('CLEAR', () => {
    it('resets value to an empty string', () => {
      const state = { ...baseState, value: 'chicken soup' };
      const result = searchContextReducer(state, { type: SearchContextAction.CLEAR });
      expect(result.value).toBe('');
    });

    it('resets results to an empty array', () => {
      const state = { ...baseState, results: [{ id: '1', name: 'Pasta' }] };
      const result = searchContextReducer(state, { type: SearchContextAction.CLEAR });
      expect(result.results).toEqual([]);
    });

    it('preserves open state and other fields', () => {
      const state = { ...baseState, open: true, value: 'pizza', results: [{}] };
      const result = searchContextReducer(state, { type: SearchContextAction.CLEAR });
      expect(result.open).toBe(true);
    });
  });

  describe('SET_VALUE', () => {
    it('updates the value field', () => {
      const result = searchContextReducer(baseState, {
        type: SearchContextAction.SET_VALUE,
        payload: 'tacos',
      });
      expect(result.value).toBe('tacos');
    });

    it('sets value to an empty string', () => {
      const state = { ...baseState, value: 'something' };
      const result = searchContextReducer(state, {
        type: SearchContextAction.SET_VALUE,
        payload: '',
      });
      expect(result.value).toBe('');
    });

    it('preserves all other state fields', () => {
      const state = { ...baseState, open: true, results: [{ id: '1' }] };
      const result = searchContextReducer(state, {
        type: SearchContextAction.SET_VALUE,
        payload: 'new query',
      });
      expect(result.open).toBe(true);
      expect(result.results).toEqual([{ id: '1' }]);
    });
  });

  describe('OPEN_DIALOG', () => {
    it('sets open to true', () => {
      const result = searchContextReducer(baseState, {
        type: SearchContextAction.OPEN_DIALOG,
      });
      expect(result.open).toBe(true);
    });

    it('is idempotent when already open', () => {
      const state = { ...baseState, open: true };
      const result = searchContextReducer(state, {
        type: SearchContextAction.OPEN_DIALOG,
      });
      expect(result.open).toBe(true);
    });

    it('preserves other fields', () => {
      const state = { ...baseState, value: 'hello', results: [{}] };
      const result = searchContextReducer(state, {
        type: SearchContextAction.OPEN_DIALOG,
      });
      expect(result.value).toBe('hello');
      expect(result.results).toEqual([{}]);
    });
  });

  describe('CLOSE_DIALOG', () => {
    it('sets open to false', () => {
      const state = { ...baseState, open: true };
      const result = searchContextReducer(state, {
        type: SearchContextAction.CLOSE_DIALOG,
      });
      expect(result.open).toBe(false);
    });

    it('is idempotent when already closed', () => {
      const result = searchContextReducer(baseState, {
        type: SearchContextAction.CLOSE_DIALOG,
      });
      expect(result.open).toBe(false);
    });

    it('preserves other fields', () => {
      const state = { ...baseState, open: true, value: 'world' };
      const result = searchContextReducer(state, {
        type: SearchContextAction.CLOSE_DIALOG,
      });
      expect(result.value).toBe('world');
    });
  });

  describe('SET_RESULTS', () => {
    it('replaces the results array', () => {
      const newResults = [
        { id: 'r1', name: 'Pasta' },
        { id: 'r2', name: 'Pizza' },
      ];
      const result = searchContextReducer(baseState, {
        type: SearchContextAction.SET_RESULTS,
        payload: newResults,
      });
      expect(result.results).toEqual(newResults);
    });

    it('can set results to an empty array', () => {
      const state = { ...baseState, results: [{ id: '1' }] };
      const result = searchContextReducer(state, {
        type: SearchContextAction.SET_RESULTS,
        payload: [],
      });
      expect(result.results).toEqual([]);
    });

    it('preserves other state fields', () => {
      const state = { ...baseState, value: 'stew', open: true };
      const result = searchContextReducer(state, {
        type: SearchContextAction.SET_RESULTS,
        payload: [{ id: 'x' }],
      });
      expect(result.value).toBe('stew');
      expect(result.open).toBe(true);
    });
  });

  describe('unknown action', () => {
    it('throws when given an unrecognised action type', () => {
      expect(() =>
        searchContextReducer(baseState, { type: 'DOES_NOT_EXIST' as any })
      ).toThrow();
    });
  });

  describe('immutability', () => {
    it('returns a new state object (does not mutate the previous state)', () => {
      const prev = { ...baseState };
      const next = searchContextReducer(prev, {
        type: SearchContextAction.SET_VALUE,
        payload: 'new',
      });
      expect(next).not.toBe(prev);
      expect(prev.value).toBe('');
    });
  });
});
