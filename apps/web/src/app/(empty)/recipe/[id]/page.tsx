import * as React from 'react';
import { notFound } from 'next/navigation';
import { Page, PageProps } from '@/libs/navigation';
import { getRecipeBannerVariantUrl } from '@/libs/media';
import { handleUserNavigation } from '@/libs/server/navigation';
import { getRecipeIdFromPageParams } from './navigation';
import { getRecipeData, getRecipeMedia } from './data';
import log from '@/libs/logger';
import type { Metadata } from 'next';
import RecipeLayout from './recipe-layout';
import RecipeContent from '@/components/Recipe/RecipeContent';
import { Graph, Recipe, HowToStep } from 'schema-dts';

export async function generateMetadata({
  params,
}: PageProps): Promise<Metadata> {
  const { id, error } = getRecipeIdFromPageParams(await params);
  if (error) {
    return notFound();
  }

  if (id === '0') {
    return {
      title: '4ks New Recipe',
      description: '4ks',
    };
  }

  // data
  const recipeData = await getRecipeData(id);
  if (!recipeData?.data) {
    return {
      title: '4ks Recipe',
      description: '4ks',
    };
  }

  // const mediaData = recipeData?.data?.root
  //   ? (await getRecipeMedia(recipeData?.data?.root)) ?? { data: [] }
  //   : { data: [] };

  // tr@ck: add recipe as jsondl

  // const m = media?.data?.map((m) => {
  //   return m.variants?.map((v) => {
  //     return v.url;
  //   });
  // });
  // console.log(m);
  // tr@ck: add media to metatada

  return {
    title: recipeData?.data?.currentRevision?.name,
    description: '4ks',
    // openGraph: {
    //   images: m as string[],
    // },
  };
}

export default async function RecipeContentPage({
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

  //jsonLd
  const md = getRecipeBannerVariantUrl(
    recipeData?.data?.currentRevision?.banner
  );

  const jsonLd: Recipe = {
    '@type': 'Recipe',
    name: recipeData?.data?.currentRevision?.name,
    recipeIngredient: recipeData?.data?.currentRevision?.ingredients?.map(
      (i) => `${i.quantity} ${i.name}`
    ),
    recipeInstructions: recipeData?.data?.currentRevision?.instructions?.map(
      (i) =>
        ({
          '@type': 'HowToStep',
          text: i.text || '',
        } as HowToStep)
    ),
    image: md?.url,
    // description: product.description,
  };

  const graph: Graph = {
    '@context': 'https://schema.org',
    '@graph': [jsonLd],
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(graph) }}
      />
      <RecipeLayout
        recipe={recipeData.data}
        user={userData.user}
        media={mediaData.data || []}
      >
        <RecipeContent user={userData.user} recipe={recipeData.data} />
      </RecipeLayout>
    </>
  );
}
