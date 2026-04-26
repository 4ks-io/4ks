import * as React from 'react';
import { serverClient } from '@/trpc/serverClient';
import { getUserData } from '@/libs/server/data';
import SearchResults from '@/components/SearchResults';
import Box from '@mui/material/Box';
import { redirect } from 'next/navigation';
import AppHeader from '@/components/AppHeader';

export default async function SearchPage({
  searchParams,
}: {
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  const resolvedSearchParams = searchParams ? await searchParams : undefined;

  if (
    !resolvedSearchParams ||
    !resolvedSearchParams['q'] ||
    resolvedSearchParams['q'] === ''
  ) {
    redirect('/explore');
  }

  const user = await getUserData();

  // fetch
  const data = await serverClient.search.recipes(
    resolvedSearchParams['q'] as string
  );
  if (!data) {
    return (
      <>
        <AppHeader user={user} />
        <div>no results</div>
      </>
    );
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
        <SearchResults results={data} />
      </Box>
    </>
  );
}
