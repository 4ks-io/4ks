import { DropResult } from '@hello-pangea/dnd';

export function cloneList<T>(list: T | undefined): T {
  return Object.assign([], list) as T;
}

export function handleListAdd<T>(
  list: T[] | undefined,
  callback: ((l: T[]) => void) | undefined
) {
  const i = cloneList<T[]>(list);
  i.push({} as any);
  callback && callback(i);
}

export function handleListDelete<T>(
  index: number,
  list: T[] | undefined,
  callback: ((l: T[]) => void) | undefined
) {
  const i = cloneList<T[]>(list);
  if (i) {
    i?.splice(index, 1);
    callback && callback(i);
  }
}

export function handleListChange<T>(
  list: T[] | undefined,
  callback: ((l: T[]) => void) | undefined
) {
  return (index: number, data: T) => {
    const i = cloneList(list);
    i[index] = data;
    callback && callback(i);
  };
}

export function reorder<T>(
  list: T[],
  startIndex: number,
  endIndex: number
): T[] {
  const result = Array.from(list);
  const [removed] = result.splice(startIndex, 1);
  result.splice(endIndex, 0, removed);
  return result;
}

export function handleListDragEnd<T>(
  list: T[] | undefined,
  callback: ((l: T[]) => void) | undefined
) {
  return (result: DropResult) => {
    if (!list || !result.destination) {
      return;
    }

    if (result.destination.index === result.source.index) {
      return;
    }

    if (list) {
      const reorderedList: T[] = reorder(
        list,
        result.source.index,
        result.destination.index
      );

      callback && callback(reorderedList);
    }
  };
}
