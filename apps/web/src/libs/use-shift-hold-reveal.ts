'use client';
import { useEffect, useRef, useState } from 'react';

export function useShiftHoldReveal(delayMs = 4000) {
  const [shiftHeld, setShiftHeld] = useState(false);
  const [revealed, setRevealed] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Shift' && !e.repeat) {
        setShiftHeld(true);
        if (!timerRef.current) {
          timerRef.current = setTimeout(() => {
            setRevealed(true);
            timerRef.current = null;
          }, delayMs);
        }
      }
    }

    function handleKeyUp(e: KeyboardEvent) {
      if (e.key === 'Shift') {
        setShiftHeld(false);
        setRevealed(false);
        if (timerRef.current) {
          clearTimeout(timerRef.current);
          timerRef.current = null;
        }
      }
    }

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [delayMs]);

  return shiftHeld && revealed;
}
