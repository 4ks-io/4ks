import { serverClient } from '@/trpc/serverClient';
import {
  dtos_GetRecipeMediaResponse,
  dtos_GetRecipeResponse,
  models_Recipe,
  models_RecipeRevision,
} from '@4ks/api-fetch';
import { TRPCError } from '@trpc/server';
import { getHTTPStatusCodeFromError } from '@trpc/server/http';
import { initialRecipe } from '@/providers/recipe-context/recipe-context-init';

// recipe
export async function getRecipeData(
  id: string
): Promise<dtos_GetRecipeResponse | undefined> {
  if (id == '0') {
    return { data: initialRecipe } as dtos_GetRecipeResponse;
  }

  try {
    return (await serverClient.recipes.getByID(id)) ?? undefined;
  } catch (e) {
    if (e instanceof TRPCError && getHTTPStatusCodeFromError(e) === 404) {
      return undefined;
    }
    // tr@ck: handle other errors?
  }
}

// recipe media
export async function getRecipeMedia(id: string) {
  try {
    return (
      (await serverClient.recipes.getMediaByID(id)) ??
      ({ data: [] } as dtos_GetRecipeMediaResponse)
    );
  } catch (e) {
    if (e instanceof TRPCError) {
      return { data: [] } as dtos_GetRecipeMediaResponse;
    }
  }
}

export async function getRecipeForks(id: string): Promise<models_Recipe[]> {
  try {
    return (await serverClient.recipes.getForksByID(id)) ?? [];
  } catch (e) {
    if (e instanceof TRPCError) {
      return [];
    }
  }

  return [];
}

export async function getRecipeRevisions(
  id: string
): Promise<models_RecipeRevision[]> {
  try {
    return (await serverClient.recipes.getRevisionsByID(id)) ?? [];
  } catch (e) {
    if (e instanceof TRPCError) {
      return [];
    }
  }

  return [];
}

export async function getRecipeRevision(
  id: string
): Promise<models_RecipeRevision | undefined> {
  try {
    return (await serverClient.recipes.getRevisionByID(id)) ?? undefined;
  } catch (e) {
    if (e instanceof TRPCError && getHTTPStatusCodeFromError(e) === 404) {
      return undefined;
    }
  }
}
