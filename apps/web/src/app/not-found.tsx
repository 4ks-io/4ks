import React from 'react';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

export default function NotFound() {
  return (
    <Box
      component="main"
      sx={{
        minHeight: '100dvh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        px: 3,
        py: 8,
        backgroundColor: 'background.default',
      }}
    >
      <Stack
        spacing={3}
        alignItems="center"
        textAlign="center"
        sx={{ width: '100%', maxWidth: 520 }}
      >
        <Box
          component="img"
          alt="4ks.io"
          src="/logo.svg"
          sx={{
            height: { xs: 72, sm: 88 },
            maxWidth: '100%',
          }}
        />
        <Stack spacing={1} alignItems="center">
          <Typography
            component="p"
            variant="overline"
            color="text.secondary"
            sx={{ lineHeight: 1.4 }}
          >
            404
          </Typography>
          <Typography component="h1" variant="h4">
            Page not found
          </Typography>
          <Typography color="text.secondary" sx={{ maxWidth: 420 }}>
            Forks could not find that page. The recipe may have moved, or the
            link may be out of date.
          </Typography>
        </Stack>
        <Stack
          direction={{ xs: 'column', sm: 'row' }}
          spacing={1.5}
          justifyContent="center"
          sx={{ width: { xs: '100%', sm: 'auto' } }}
        >
          <Button
            href="/"
            variant="contained"
            color="secondary"
            sx={{ minWidth: 132 }}
          >
            Home
          </Button>
          <Button href="/explore" variant="outlined" sx={{ minWidth: 132 }}>
            Explore
          </Button>
        </Stack>
      </Stack>
    </Box>
  );
}
