import React from 'react';
import LandingPageSearchBox from '@/components/LandingPageSearchBox';
import AppHeaderAvatarAuthenticated from '@/components/AppHeaderAvatarAuthenticated';
import AppHeaderAvatarUnauthenticated from '@/components/AppHeaderAvatarUnauthenticated';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import Tooltip from '@mui/material/Tooltip';
import { handleUserNavigation } from '@/libs/server/navigation';
import { Page, PageProps } from '@/libs/navigation';
import NewRecipeButton from '@/components/NewRecipeButton';
import FetchRecipeButton from '@/components/FetchRecipeButton';
import Link from 'next/link';

export default async function DefaultPage({}: PageProps) {
  const { user } = await handleUserNavigation(Page.ANONYMOUS);

  return (
    <Box height="92vh" display="flex" flexDirection="column">
      <FetchRecipeButton />
      <NewRecipeButton />
      <Toolbar variant="dense">
        <Box sx={{ flexGrow: 1 }} />
        {user ? (
          <AppHeaderAvatarAuthenticated username={`${user.username}`} />
        ) : (
          <AppHeaderAvatarUnauthenticated />
        )}
      </Toolbar>

      <Stack
        height={'100%'}
        direction="column"
        justifyContent="center"
        alignItems="center"
        spacing={2}
      >
        <Link
          prefetch={false}
          href="/explore"
          style={{ textDecoration: 'none', color: '#000' }}
        >
          <Tooltip title="Explore">
            <Box
              component="img"
              sx={{ height: 96, paddingRight: 1 }}
              alt="4ks.io"
              src={'/logo.svg'}
            />
          </Tooltip>
        </Link>
        <LandingPageSearchBox />
      </Stack>
    </Box>
  );
}
