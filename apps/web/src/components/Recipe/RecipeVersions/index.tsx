'use client';
import * as React from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import Link from 'next/link';
import {
  models_Recipe,
  models_RecipeRevision,
  models_User,
} from '@4ks/api-fetch';

type RecipeVersionsProps = {
  recipe: models_Recipe;
  revisions: models_RecipeRevision[];
  user: models_User | undefined;
};

export default function RecipeVersions({
  recipe,
  revisions,
  user,
}: RecipeVersionsProps) {
  if (revisions.length === 0) {
    return (
      <Box sx={{ display: 'flex' }}>
        <Typography color="text.secondary">
          No version history is available for this recipe.
        </Typography>
      </Box>
    );
  }

  return (
    <Stack spacing={2}>
      {revisions.map((revision) => {
        const isCurrent = recipe.currentRevision?.id === revision.id;

        return (
          <Card key={revision.id} variant="outlined">
            <CardContent>
              <Stack spacing={1.5}>
                <Stack
                  direction={{ xs: 'column', sm: 'row' }}
                  spacing={1}
                  justifyContent="space-between"
                  alignItems={{ xs: 'flex-start', sm: 'center' }}
                >
                  <Box>
                    <Typography variant="h6">
                      {revision.name || 'Untitled version'}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      By @{revision.author?.username || 'chef'}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Saved {formatDate(revision.createdDate)}
                    </Typography>
                  </Box>
                  <Stack direction="row" spacing={1}>
                    {isCurrent ? (
                      <Chip label="Current version" color="primary" size="small" />
                    ) : (
                      <Chip label="Immutable snapshot" size="small" />
                    )}
                  </Stack>
                </Stack>

                <Stack direction="row" spacing={1}>
                  <Button
                    component={Link}
                    href={`/recipe/revisions/${revision.id}`}
                    prefetch={false}
                    variant="outlined"
                    size="small"
                  >
                    View
                  </Button>
                  {!isCurrent && (
                    <Button
                      component={Link}
                      href={`/recipe/revisions/${revision.id}`}
                      prefetch={false}
                      variant="contained"
                      size="small"
                    >
                      {user?.id ? 'Fork this version' : 'View to fork'}
                    </Button>
                  )}
                </Stack>
              </Stack>
            </CardContent>
          </Card>
        );
      })}
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
