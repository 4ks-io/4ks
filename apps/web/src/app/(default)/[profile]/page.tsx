import * as React from 'react';
import Link from 'next/link';
import { serverClient } from '@/trpc/serverClient';
import { Page, PageProps } from '@/libs/navigation';
import { getUserData } from '@/libs/server/data';
import { dtos_GetRecipesByUsernameResponse, models_UserEvent } from '@4ks/api-fetch';
import { getHTTPStatusCodeFromError } from '@trpc/server/http';
import { TRPCError } from '@trpc/server';
import { notFound } from 'next/navigation';
import log from '@/libs/logger';
import AppHeader from '@/components/AppHeader';
import SearchResults from '@/components/SearchResults';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import { normalizeForURL } from '@/libs/navigation';
import NotificationsActiveIcon from '@mui/icons-material/NotificationsActive';

export type ProfilePropsParams = { profile: string };

export default async function ProfilePage({ params }: PageProps) {
  const resolvedParams = await params;
  const username = (resolvedParams as unknown as ProfilePropsParams).profile;
  const user = await getUserData();

  const isCurrent = user?.username === username;

  const data =
    (await serverClient.search.recipesByAuthor(username)) ?? undefined;

  // fetch
  // tr@ck: should below and above calls be concurrent? the profile page needs more content. redesign
  if (!data) {
    let r = {} as dtos_GetRecipesByUsernameResponse;
    try {
      r = (await serverClient.recipes.getAllByAuthor(username)) ?? {};
    } catch (e) {
      if (e instanceof TRPCError && getHTTPStatusCodeFromError(e) === 404) {
        log().Error(new Error(), [
          { k: 'msg', v: 'ProfilePage: failed to fetch recipes' },
        ]);
        return notFound();
      }
    }
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
        <Container>
          <h2>@{username}</h2>
          {isCurrent && (
            <>
              <h3>Events</h3>
              <ul style={{ listStyleType: 'none', paddingInlineStart: '0px' }}>
                {user.events?.map((e: models_UserEvent) => {
                  if (!e || !e.data) {
                    return null;
                  }
                  // tr@ck: handle name change
                  let recipeURL = `/recipe/${e.data.RecipeID}`;
                  if (e.data?.RecipeTitle) {
                    recipeURL = `${recipeURL}-${normalizeForURL(
                      e.data.RecipeTitle
                    )}`;
                  }
                  return (
                    <li key={e.id}>
                      <Link
                        prefetch={false}
                        style={{ textDecoration: 'none', color: '#000' }}
                        href={recipeURL}
                      >
                        <Typography variant="h6" noWrap>
                          {e.status == 2 && (
                            <NotificationsActiveIcon
                              fontSize="inherit"
                              color="secondary"
                              sx={{ paddingRight: 1 }}
                            />
                          )}
                          {e?.data?.RecipeTitle}
                        </Typography>
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </>
          )}
          <h3>Recipes</h3>
          {<SearchResults results={data} />}
        </Container>
      </Box>
    </>
  );
}
