import { useEffect, useState } from 'react';
import { Droppable, DroppableProps } from '@hello-pangea/dnd';

// StrictModeDroppable.tsx
// Credits to https://github.com/GiovanniACamacho and https://github.com/Meligy for the TypeScript version
// Original post: https://github.com/atlassian/@hello-pangea/dnd/issues/2399#issuecomment-1175638194

export default function StrictModeDroppable({
  children,
  ...props
}: DroppableProps) {
  const [enabled, setEnabled] = useState(false);
  useEffect(() => {
    const animation = requestAnimationFrame(() => setEnabled(true));
    return () => {
      cancelAnimationFrame(animation);
      setEnabled(false);
    };
  }, []);
  if (!enabled) {
    return null;
  }
  return <Droppable {...props}>{children}</Droppable>;
}
