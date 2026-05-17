'use client';
import React from 'react';
import { useRecipeContext } from '@/providers/recipe-context';
import { SectionTitle } from '../SectionTitle';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { normalizeForURL } from '@/libs/navigation';
import LinkIcon from '@mui/icons-material/Link';
import Link from 'next/link';

interface RecipeSocialProps {}

export default function RecipeSocial(props: RecipeSocialProps) {
  const rtx = useRecipeContext();

  return (
    <>
      <Stack>
        <SectionTitle value={'Contributors'} />
        {rtx?.recipe?.contributors?.map((c) => {
          return (
            <Typography key={c.username} variant="body1" gutterBottom>
              {c.username}
            </Typography>
          );
        })}
      </Stack>
      <div style={{ paddingBottom: 30 }} />

      <Stack>
        <SectionTitle value={'Source'} />
        {rtx.recipe?.branch && (
          <Link
            prefetch={false}
            href={`/recipe/${rtx.recipe?.branch}`}
            style={{ textDecoration: 'none', color: '#000' }}
          >
            <Typography variant="body1" gutterBottom>
              <LinkIcon fontSize="small" /> Inspiration
            </Typography>
          </Link>
        )}
        {rtx?.recipe.currentRevision?.link && (
          <Typography variant="body1" gutterBottom>
            <a
              href={ensureAbsoluteUrl(rtx.recipe.currentRevision?.link)}
              style={{ textDecoration: 'none', color: '#000' }}
              target="_blank"
              rel="noopener noreferrer"
            >
              <LinkIcon fontSize="small" />
              {getDomain(rtx.recipe.currentRevision?.link) || 'http'}
            </a>
          </Typography>
        )}
      </Stack>
      <div style={{ paddingBottom: 30 }} />
    </>
  );
}

function ensureAbsoluteUrl(url: string): string {
  return /^https?:\/\//i.test(url) ? url : `https://${url}`;
}

function getDomain(url: string): string {
  // Make 'subdomain' an optional parameter with a default value of false
  url = url.replace(/(https?:\/\/)?/i, ''); // Remove protocol and 'www' subdomain
  const parts = url.split('/'); // Split URL by forward slash
  return parts[0];
}
