import * as React from 'react';
import { notFound } from 'next/navigation';
import type { Metadata } from 'next';
import Container from '@mui/material/Container';
import AppHeader from '@/components/AppHeader';
import RecipeRevisionDetail from '@/components/Recipe/RecipeRevisionDetail';
import { Page } from '@/libs/navigation';
import { handleUserNavigation } from '@/libs/server/navigation';
import { getRecipeData, getRecipeRevision } from '../../[id]/data';

type RevisionPageProps = {
  params: Promise<{ revisionID: string }>;
};

export async function generateMetadata({
  params,
}: RevisionPageProps): Promise<Metadata> {
  const { revisionID } = await params;
  const revision = await getRecipeRevision(revisionID);

  if (!revision) {
    return {
      title: '4ks Version',
      description: '4ks',
    };
  }

  return {
    title: revision.name || '4ks Version',
    description: '4ks',
  };
}

export default async function RecipeRevisionPage({
  params,
}: RevisionPageProps) {
  const { revisionID } = await params;

  const revision = await getRecipeRevision(revisionID);
  if (!revision?.recipeId) {
    return notFound();
  }

  const [recipeData, userData] = await Promise.all([
    getRecipeData(revision.recipeId),
    handleUserNavigation(Page.ANONYMOUS),
  ]);

  if (!recipeData?.data) {
    return notFound();
  }

  return (
    <>
      <AppHeader user={userData.user} />
      <Container sx={{ marginTop: 4, marginBottom: 6 }}>
        <RecipeRevisionDetail
          user={userData.user}
          recipe={recipeData.data}
          revision={revision}
        />
      </Container>
    </>
  );
}
