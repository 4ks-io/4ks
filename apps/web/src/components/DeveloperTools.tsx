'use client';
import React, { useState } from 'react';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Snackbar from '@mui/material/Snackbar';
import { useShiftHoldReveal } from '@/libs/use-shift-hold-reveal';

type DeveloperToolsProps = {
  t: string;
};

export default function DeveloperTools({ t }: DeveloperToolsProps) {
  const [copied, setCopied] = useState(false);
  const revealed = useShiftHoldReveal();

  function handleClick() {
    navigator.clipboard.writeText(t);
    setCopied(true);
  }

  if (!revealed) return null;

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
