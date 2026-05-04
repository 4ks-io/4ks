'use client';
import React, { useState, useEffect, useRef } from 'react';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Snackbar from '@mui/material/Snackbar';

type DeveloperToolsProps = {
  t: string;
};

export default function DeveloperTools({ t }: DeveloperToolsProps) {
  const [shiftHeld, setShiftHeld] = useState(false);
  const [revealed, setRevealed] = useState(false);
  const [copied, setCopied] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Shift' && !e.repeat) {
        setShiftHeld(true);
        if (!timerRef.current) {
          timerRef.current = setTimeout(() => {
            setRevealed(true);
            timerRef.current = null;
          }, 4000);
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
  }, []);

  function handleClick() {
    navigator.clipboard.writeText(t);
    setCopied(true);
  }

  if (!shiftHeld || !revealed) return null;

  return (
    <Box
      sx={{
        position: 'fixed',
        bottom: 32,
        left: '50%',
        transform: 'translateX(-50%)',
        zIndex: 9999,
      }}
    >
      <Button
        variant="contained"
        color="warning"
        size="large"
        onClick={handleClick}
        sx={{ px: 6, py: 2, fontSize: '1.1rem', fontWeight: 700 }}
      >
        {copied ? 'Copied!' : 'Copy Access Token'}
      </Button>
      <Snackbar
        open={copied}
        autoHideDuration={2000}
        onClose={() => setCopied(false)}
        message="Access token copied to clipboard"
      />
    </Box>
  );
}
