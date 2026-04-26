import * as React from 'react';
import { notFound } from 'next/navigation';
import { Page, PageProps } from '@/libs/navigation';
import { handleUserNavigation } from '@/libs/server/navigation';
import { getRecipeIdFromPageParams } from '../navigation';
import { getRecipeData, getRecipeMedia } from '../data';
import log from '@/libs/logger';
import type { Metadata } from 'next';
import RecipeLayout from '../recipe-layout';
import RecipeForks from '@/components/Recipe/RecipeForks';

export async function generateMetadata({
  params,
}: PageProps): Promise<Metadata> {
  const { id, error } = getRecipeIdFromPageParams(await params);
  if (error) {
    return notFound();
  }

  // data
  const recipeData = await getRecipeData(id);
  if (!recipeData?.data) {
    return {
      title: '4ks Forks',
      description: '4ks',
    };
  }

  return {
    title: recipeData?.data?.currentRevision?.name,
    description: '4ks',
  };
}

export default async function RecipeForksPage({
  params,
}: PageProps) {
  const { id, error } = getRecipeIdFromPageParams(await params);
  if (error) {
    return notFound();
  }

  // data
  const [recipeData, userData] = await Promise.all([
    getRecipeData(id),
    handleUserNavigation(Page.ANONYMOUS),
  ]);

  if (!recipeData?.data) {
    log().Error(new Error(), [{ k: 'msg', v: 'failed to fetch recipe' }]);
    return notFound();
  }

  const mediaData = recipeData?.data?.root
    ? (await getRecipeMedia(recipeData?.data?.root)) ?? { data: [] }
    : { data: [] };

  return (
    <RecipeLayout
      recipe={recipeData.data}
      user={userData.user}
      media={mediaData.data || []}
    >
      <RecipeForks user={userData.user} recipe={recipeData.data} />
    </RecipeLayout>
  );
}
