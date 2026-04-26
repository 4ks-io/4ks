'use client';
import React, { useEffect } from 'react';
import {
  models_Recipe,
  models_RecipeRevision,
  models_User,
} from '@4ks/api-fetch';
import { trpc } from '@/trpc/client';
import { normalizeForURL, navigateToLogin } from '@/libs/navigation';
import { useRouter } from 'next/navigation';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import Link from 'next/link';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

type RecipeRevisionDetailProps = {
  user: models_User | undefined;
  recipe: models_Recipe;
  revision: models_RecipeRevision;
};

export default function RecipeRevisionDetail({
  user,
  recipe,
  revision,
}: RecipeRevisionDetailProps) {
  const router = useRouter();
  const forkRevision = trpc.recipes.forkRevision.useMutation();
  const isAuthenticated = !!user?.id;
  const isCurrent = recipe.currentRevision?.id === revision.id;

  useEffect(() => {
    if (forkRevision.isSuccess) {
      router.push(
        `/recipe/${forkRevision.data?.id}-${normalizeForURL(
          forkRevision.data?.currentRevision?.name
        )}`
      );
    }
  }, [forkRevision.data, forkRevision.isSuccess, router]);

  function handleFork() {
    if (!isAuthenticated) {
      navigateToLogin(`/recipe/revisions/${revision.id}`);
      return;
    }

    forkRevision.mutate(`${revision.id}`);
  }

  return (
    <Stack spacing={3}>
      <Stack spacing={1}>
        <Typography variant="h4">
          {revision.name || 'Untitled version'}
        </Typography>
        <Stack direction="row" spacing={1} flexWrap="wrap">
          <Chip
            label={isCurrent ? 'Current version' : 'Immutable snapshot'}
            color={isCurrent ? 'primary' : 'default'}
            size="small"
          />
          <Chip
            label={`Saved ${formatDate(revision.createdDate)}`}
            size="small"
          />
          <Chip label={`By @${revision.author?.username || 'chef'}`} size="small" />
        </Stack>
      </Stack>

      {!isCurrent && (
        <Alert severity="info">
          This is a read-only historical version. It cannot be edited in place.
        </Alert>
      )}

      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
        <Button
          component={Link}
          href={`/recipe/${recipe.id}-${normalizeForURL(recipe.currentRevision?.name)}/versions`}
          prefetch={false}
          variant="outlined"
        >
          Back to versions
        </Button>
        <Button
          component={Link}
          href={`/recipe/${recipe.id}-${normalizeForURL(recipe.currentRevision?.name)}`}
          prefetch={false}
          variant="outlined"
        >
          Open current recipe
        </Button>
        {!isCurrent && (
          <Button
            onClick={handleFork}
            variant="contained"
            disabled={forkRevision.isPending}
          >
            Fork this version
          </Button>
        )}
      </Stack>

      <Card variant="outlined">
        <CardContent>
          <Stack spacing={2}>
            <Box>
              <Typography variant="h5" gutterBottom>
                Ingredients
              </Typography>
              <Stack spacing={1}>
                {(revision.ingredients || []).map((ingredient, index) => (
                  <Typography key={`${ingredient.id}-${index}`} variant="body1">
                    {ingredient.quantity
                      ? `${ingredient.quantity} ${ingredient.name}`
                      : ingredient.name}
                  </Typography>
                ))}
              </Stack>
            </Box>

            <Divider />

            <Box>
              <Typography variant="h5" gutterBottom>
                Instructions
              </Typography>
              <Stack spacing={1.5}>
                {(revision.instructions || []).map((instruction, index) => (
                  <Box key={`${instruction.id}-${index}`}>
                    <Typography variant="subtitle2" color="text.secondary">
                      Step {index + 1}
                    </Typography>
                    <Typography variant="body1">
                      {instruction.text || instruction.name}
                    </Typography>
                  </Box>
                ))}
              </Stack>
            </Box>
          </Stack>
        </CardContent>
      </Card>
    </Stack>
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
