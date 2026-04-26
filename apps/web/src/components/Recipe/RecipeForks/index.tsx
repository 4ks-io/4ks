'use client';
import * as React from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import Link from 'next/link';
import { models_Recipe } from '@4ks/api-fetch';
import { normalizeForURL } from '@/libs/navigation';

type RecipeForksProps = {
  forks: models_Recipe[];
  parentRecipe?: models_Recipe;
};

export default function RecipeForks({
  forks,
  parentRecipe,
}: RecipeForksProps) {
  if (forks.length === 0 && !parentRecipe) {
    return (
      <Box sx={{ display: 'flex' }}>
        <Typography color="text.secondary">
          No forks yet. The first fork will appear here.
        </Typography>
      </Box>
    );
  }

  return (
    <Stack spacing={2}>
      {parentRecipe && (
        <ForkCard
          recipe={parentRecipe}
          authorLabel="Original recipe"
          relationshipLabel="Parent"
        />
      )}
      {forks.map((fork) => {
        return (
          <ForkCard key={fork.id} recipe={fork} authorLabel="Forked by" />
        );
      })}
    </Stack>
  );
}

type ForkCardProps = {
  recipe: models_Recipe;
  authorLabel: string;
  relationshipLabel?: string;
};

function ForkCard({
  recipe,
  authorLabel,
  relationshipLabel,
}: ForkCardProps) {
  const title = recipe.currentRevision?.name || 'Untitled recipe';
  const href = `/recipe/${recipe.id}-${normalizeForURL(title)}`;

  return (
    <Card key={recipe.id} variant="outlined">
      <CardContent>
        <Stack
          direction={{ xs: 'column', sm: 'row' }}
          spacing={1}
          justifyContent="space-between"
          alignItems={{ xs: 'flex-start', sm: 'center' }}
        >
          <Box>
            <Typography
              component={Link}
              href={href}
              prefetch={false}
              sx={{ textDecoration: 'none', color: 'inherit' }}
              variant="h6"
            >
              {title}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {authorLabel} @{recipe.author?.username || 'chef'}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Updated {formatDate(recipe.updatedDate)}
            </Typography>
          </Box>
          <Stack direction="row" spacing={1}>
            {relationshipLabel && (
              <Chip label={relationshipLabel} size="small" color="primary" />
            )}
            <Chip label={`${recipe.metadata?.forks || 0} forks`} size="small" />
            <Chip label={`${recipe.metadata?.stars || 0} stars`} size="small" />
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
}

function formatDate(value?: string) {
  if (!value) {
    return 'unknown date';
  }

  return new Intl.DateTimeFormat('en', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}
