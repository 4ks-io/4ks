'use client';
import React from 'react';
import { Hit } from './types';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardMedia from '@mui/material/CardMedia';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';

function CustomHit(h: any, handleClose: () => void) {
  const id = h['id'];
  const router = useRouter();

  return (
    <Card
      key={id}
      sx={{ display: 'flex', cursor: 'pointer' }}
      onClick={() => {
        handleClose();
        router.push(`/recipe/${id}`);
      }}
    >
      <Box sx={{ display: 'flex', flexDirection: 'column', width: '100%' }}>
        <CardContent sx={{ flex: '1 0 auto' }}>
          <Typography component="div" variant="h6">
            {h['name'] as string}
          </Typography>
          <Typography
            variant="subtitle1"
            color="text.secondary"
            component="div"
          >
            {h['author'] as string}
          </Typography>
        </CardContent>
      </Box>
      <CardMedia component="img" sx={{ width: 100 }} image={h['imageUrl']} />
    </Card>
  );
}

export function FormattedHits(
  hits: any,
  title: string,
  handleClose: () => void
) {
  if (hits.length === 0) {
    return null;
  }

  return (
    <div style={{ paddingBottom: 20 }}>
      <Box
        noValidate
        component="form"
        sx={{
          display: 'flex',
          flexDirection: 'column',
          m: 'auto',
        }}
      >
        <Typography variant="h5" gutterBottom>
          {title}
        </Typography>
        <Stack spacing={1}>
          {hits.map((h: Hit) => CustomHit(h, handleClose))}
        </Stack>
      </Box>
    </div>
  );
}
