import React from 'react';
import UpdateUsername from '@/components/UpdateUsername';
import { Page, PageProps } from '@/libs/navigation';
import { handleUserNavigation } from '@/libs/server/navigation';
import { auth0 } from '@/libs/auth0';
import type { Metadata } from 'next';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import Stack from '@mui/material/Stack';
import UsernameSpecification from '@/components/UsernameSpecifications';
import DeveloperTools from '@/components/DeveloperTools';
import AppHeader from '@/components/AppHeader';
import SettingsKitchenPass from '@/components/SettingsKitchenPass';
import { serverClient } from '@/trpc/serverClient';

export const metadata: Metadata = {
  title: '4ks Settings',
  description: '4ks User Settings',
};

export default async function SettingsPage({
}: PageProps) {
  const { user } = await handleUserNavigation(Page.AUTHENTICATED);
  const session = await auth0.getSession();
  const kitchenPass = await serverClient.users.getKitchenPass();

  if (!session || !session?.user || !user || !user?.username) {
    return <div>Error</div>;
  }

  return (
    <>
      <AppHeader user={user} />
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          bgcolor: 'background.default',
          // mt: ['12px', '18px', '24px'],
          p: 3,
        }}
      >
        <Container maxWidth="sm" style={{ paddingTop: 40 }}>
          <Typography variant="h4" component="h2">
            Settings
          </Typography>
          <Stack spacing={2} sx={{ paddingTop: 5 }}>
            <SettingsKitchenPass initialKitchenPass={kitchenPass} />
          </Stack>
          <Stack spacing={2} style={{ paddingTop: 40 }}>
            <Typography variant="subtitle1" component="h2">
              Email: {user.emailAddress}
            </Typography>
            <Typography variant="subtitle1" component="h2">
              Current Username: {user.username}
            </Typography>
          </Stack>
          <Stack spacing={2} style={{ paddingTop: 40 }}>
             <Box>
              <Typography variant="h5" component="h3">
                Display Name
              </Typography>
              <Typography variant="body1" color="text.secondary" sx={{ mt: 1 }}>
                Control your public username and how it appears to other users.
              </Typography>
            </Box>
            <UpdateUsername username={user.username} />
            <UsernameSpecification />
          </Stack>
        </Container>
      </Box>
      {session?.tokenSet.accessToken && (
        <DeveloperTools t={session.tokenSet.accessToken} />
      )}
    </>
  );
}
