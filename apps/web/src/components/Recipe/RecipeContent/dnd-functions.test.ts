import { vi, describe, it, expect } from 'vitest';
import {
  cloneList,
  handleListAdd,
  handleListDelete,
  handleListChange,
  reorder,
  handleListDragEnd,
} from './dnd-functions';

// Minimal DropResult for handleListDragEnd tests
function makeDropResult(sourceIndex: number, destinationIndex: number | null) {
  return {
    draggableId: 'item-1',
    type: 'DEFAULT',
    mode: 'FLUID' as const,
    reason: 'DROP' as const,
    combine: null,
    source: { index: sourceIndex, droppableId: 'list' },
    destination:
      destinationIndex !== null
        ? { index: destinationIndex, droppableId: 'list' }
        : null,
  };
}

describe('cloneList', () => {
  it('returns a shallow copy with the same elements', () => {
    const original = [1, 2, 3];
    const clone = cloneList(original);
    expect(clone).toEqual([1, 2, 3]);
  });

  it('the clone is a different reference', () => {
    const original = ['a', 'b'];
    expect(cloneList(original)).not.toBe(original);
  });

  it('returns an empty array when given undefined', () => {
    expect(cloneList(undefined)).toEqual([]);
  });

  it('works with objects', () => {
    const original = [{ name: 'flour', quantity: '2 cups' }];
    const clone = cloneList(original);
    expect(clone).toEqual(original);
    expect(clone).not.toBe(original);
  });
});

describe('handleListAdd', () => {
  it('appends an empty object to the list and calls the callback', () => {
    const list = [{ name: 'item1' }];
    const callback = vi.fn();
    handleListAdd(list, callback);
    expect(callback).toHaveBeenCalledOnce();
    expect(callback).toHaveBeenCalledWith([{ name: 'item1' }, {}]);
  });

  it('does not mutate the original list', () => {
    const list = [{ name: 'item1' }];
    handleListAdd(list, vi.fn());
    expect(list).toHaveLength(1);
  });

  it('works on an empty list', () => {
    const callback = vi.fn();
    handleListAdd([], callback);
    expect(callback).toHaveBeenCalledWith([{}]);
  });

  it('works when list is undefined (treats it as empty)', () => {
    const callback = vi.fn();
    handleListAdd(undefined, callback);
    expect(callback).toHaveBeenCalledWith([{}]);
  });

  it('does not throw when callback is undefined', () => {
    expect(() => handleListAdd([{ name: 'x' }], undefined)).not.toThrow();
  });

  it('does not call a null/undefined callback', () => {
    const callback = vi.fn();
    handleListAdd([{ name: 'x' }], undefined);
    expect(callback).not.toHaveBeenCalled();
  });
});

describe('handleListDelete', () => {
  it('removes the element at the specified index', () => {
    const callback = vi.fn();
    handleListDelete(1, ['a', 'b', 'c'], callback);
    expect(callback).toHaveBeenCalledWith(['a', 'c']);
  });

  it('removes the first element', () => {
    const callback = vi.fn();
    handleListDelete(0, ['a', 'b', 'c'], callback);
    expect(callback).toHaveBeenCalledWith(['b', 'c']);
  });

  it('removes the last element', () => {
    const callback = vi.fn();
    handleListDelete(2, ['a', 'b', 'c'], callback);
    expect(callback).toHaveBeenCalledWith(['a', 'b']);
  });

  it('does not mutate the original list', () => {
    const list = ['a', 'b', 'c'];
    handleListDelete(1, list, vi.fn());
    expect(list).toEqual(['a', 'b', 'c']);
  });

  it('does not throw when callback is undefined', () => {
    expect(() => handleListDelete(0, ['a'], undefined)).not.toThrow();
  });

  it('works with objects', () => {
    const list = [{ id: 1 }, { id: 2 }, { id: 3 }];
    const callback = vi.fn();
    handleListDelete(1, list, callback);
    expect(callback).toHaveBeenCalledWith([{ id: 1 }, { id: 3 }]);
  });
});

describe('handleListChange', () => {
  it('returns a function', () => {
    const onChange = handleListChange(['a', 'b'], vi.fn());
    expect(typeof onChange).toBe('function');
  });

  it('the returned function updates the element at the given index', () => {
    const callback = vi.fn();
    const onChange = handleListChange(['a', 'b', 'c'], callback);
    onChange(1, 'X');
    expect(callback).toHaveBeenCalledWith(['a', 'X', 'c']);
  });

  it('the returned function does not mutate the original list', () => {
    const list = ['a', 'b', 'c'];
    const onChange = handleListChange(list, vi.fn());
    onChange(0, 'Z');
    expect(list).toEqual(['a', 'b', 'c']);
  });

  it('works for updating the first element', () => {
    const callback = vi.fn();
    handleListChange(['a', 'b', 'c'], callback)(0, 'Z');
    expect(callback).toHaveBeenCalledWith(['Z', 'b', 'c']);
  });

  it('works for updating the last element', () => {
    const callback = vi.fn();
    handleListChange(['a', 'b', 'c'], callback)(2, 'Z');
    expect(callback).toHaveBeenCalledWith(['a', 'b', 'Z']);
  });

  it('works with objects', () => {
    const list = [{ name: 'salt' }, { name: 'pepper' }];
    const callback = vi.fn();
    handleListChange(list, callback)(0, { name: 'sugar' });
    expect(callback).toHaveBeenCalledWith([{ name: 'sugar' }, { name: 'pepper' }]);
  });

  it('does not throw when callback is undefined', () => {
    const onChange = handleListChange(['a', 'b'], undefined);
    expect(() => onChange(0, 'X')).not.toThrow();
  });
});

describe('reorder', () => {
  it('moves an element forward (lower to higher index)', () => {
    expect(reorder(['a', 'b', 'c'], 0, 2)).toEqual(['b', 'c', 'a']);
  });

  it('moves an element backward (higher to lower index)', () => {
    expect(reorder(['a', 'b', 'c'], 2, 0)).toEqual(['c', 'a', 'b']);
  });

  it('returns the same order when source equals destination', () => {
    expect(reorder(['a', 'b', 'c'], 1, 1)).toEqual(['a', 'b', 'c']);
  });

  it('handles a single-element list', () => {
    expect(reorder(['a'], 0, 0)).toEqual(['a']);
  });

  it('handles adjacent elements', () => {
    expect(reorder(['a', 'b', 'c'], 0, 1)).toEqual(['b', 'a', 'c']);
    expect(reorder(['a', 'b', 'c'], 1, 0)).toEqual(['b', 'a', 'c']);
  });

  it('does not mutate the original list', () => {
    const original = ['a', 'b', 'c'];
    reorder(original, 0, 2);
    expect(original).toEqual(['a', 'b', 'c']);
  });

  it('works with objects', () => {
    const list = [{ id: 1 }, { id: 2 }, { id: 3 }];
    expect(reorder(list, 0, 2)).toEqual([{ id: 2 }, { id: 3 }, { id: 1 }]);
  });
});

describe('handleListDragEnd', () => {
  it('returns a function', () => {
    const handler = handleListDragEnd(['a', 'b'], vi.fn());
    expect(typeof handler).toBe('function');
  });

  it('calls callback with the reordered list on a valid drag', () => {
    const callback = vi.fn();
    const handler = handleListDragEnd(['a', 'b', 'c'], callback);
    handler(makeDropResult(0, 2) as any);
    expect(callback).toHaveBeenCalledWith(['b', 'c', 'a']);
  });

  it('does nothing when destination is null', () => {
    const callback = vi.fn();
    const handler = handleListDragEnd(['a', 'b', 'c'], callback);
    handler(makeDropResult(0, null) as any);
    expect(callback).not.toHaveBeenCalled();
  });

  it('does nothing when source and destination indices are identical', () => {
    const callback = vi.fn();
    const handler = handleListDragEnd(['a', 'b', 'c'], callback);
    handler(makeDropResult(1, 1) as any);
    expect(callback).not.toHaveBeenCalled();
  });

  it('does nothing when the list is undefined', () => {
    const callback = vi.fn();
    const handler = handleListDragEnd(undefined, callback);
    handler(makeDropResult(0, 2) as any);
    expect(callback).not.toHaveBeenCalled();
  });

  it('does not throw when callback is undefined', () => {
    const handler = handleListDragEnd(['a', 'b', 'c'], undefined);
    expect(() => handler(makeDropResult(0, 2) as any)).not.toThrow();
  });

  it('does not mutate the original list', () => {
    const list = ['a', 'b', 'c'];
    const handler = handleListDragEnd(list, vi.fn());
    handler(makeDropResult(0, 2) as any);
    expect(list).toEqual(['a', 'b', 'c']);
  });

  it('moves items backward (drag from end to start)', () => {
    const callback = vi.fn();
    handleListDragEnd(['a', 'b', 'c'], callback)(makeDropResult(2, 0) as any);
    expect(callback).toHaveBeenCalledWith(['c', 'a', 'b']);
  });
});
