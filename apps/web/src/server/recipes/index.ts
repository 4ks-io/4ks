import { z } from 'zod';
import { publicProcedure, router, logTrpc } from '@/server/trpc';
import { getAPIClient, handleAPIError } from '..';
import { dtos_CreateRecipe, dtos_UpdateRecipe } from '@4ks/api-fetch';
import { validateFetchURL } from '@/libs/fetch-url';

export const recipesRouter = router({
  getSignedURL: publicProcedure
    .input(
      z.object({
        recipeID: z.string().trim(),
        payload: z.object({
          filename: z.string(),
        }),
      })
    )
    .mutation(async (opts) => {
      const api = await getAPIClient();
      const s = performance.now();

      try {
        return await api.recipes.postApiRecipesMedia(
          opts.input.recipeID,
          opts.input.payload
        );
      } catch (e) {
        handleAPIError(e);
      } finally {
        logTrpc(new Error(), opts.input, s, 'recipes.getSignedURL');
      }
    }),
  getAllByAuthor: publicProcedure.input(z.string()).query(async (opts) => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.recipes.getApiRecipesAuthor(opts.input);
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), opts.input, s, 'recipes.getAllByAuthor');
    }
  }),
  // tr@ck: only mutation?
  getByID: publicProcedure.input(z.string()).query(async (opts) => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.recipes.getApiRecipes1(opts.input);
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), opts.input, s, 'recipes.getByID');
    }
  }),
  getByIDMutation: publicProcedure.input(z.string()).mutation(async (opts) => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      const recipe = await api.recipes.getApiRecipes1(opts.input);
      return recipe;
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), opts.input, s, 'recipes.getByIDMutation');
    }
  }),
  getMediaByID: publicProcedure.input(z.string()).query(async (opts) => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.recipes.getApiRecipesMedia(opts.input);
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), opts.input, s, 'recipes.getMediaByIDMutation');
    }
  }),
  getMediaByIDMutation: publicProcedure
    .input(z.string())
    .mutation(async (opts) => {
      const api = await getAPIClient();
      const s = performance.now();

      try {
        return await api.recipes.getApiRecipesMedia(opts.input);
      } catch (e) {
        handleAPIError(e);
      } finally {
        logTrpc(new Error(), opts.input, s, 'recipes.getMediaByIDMutation');
      }
    }),
  fetch: publicProcedure
    .input(
      z
        .string()
        .trim()
        .superRefine((value, ctx) => {
          try {
            validateFetchURL(value);
          } catch (error) {
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: error instanceof Error ? error.message : 'Invalid URL',
            });
          }
        })
    )
    .mutation(async (opts) => {
      const api = await getAPIClient();
      const s = performance.now();

      try {
        return await api.recipes.postApiRecipesFetch({
          url: validateFetchURL(opts.input),
        });
      } catch (e) {
        handleAPIError(e);
      } finally {
        logTrpc(new Error(), opts.input, s, 'recipes.fetch');
      }
    }),
  fork: publicProcedure.input(z.string()).mutation(async (opts) => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.recipes.postApiRecipesFork(opts.input);
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), opts.input, s, 'recipes.fork');
    }
  }),
  star: publicProcedure.input(z.string()).mutation(async (opts) => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.recipes.postApiRecipesStar(opts.input);
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), opts.input, s, 'recipes.star');
    }
  }),
  delete: publicProcedure.input(z.string()).mutation(async (opts) => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.recipes.deleteApiRecipes(opts.input);
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), opts.input, s, 'recipes.delete');
    }
  }),
  create: publicProcedure.input(z.any()).mutation(async (opts) => {
    const api = await getAPIClient();
    const s = performance.now();

    try {
      return await api.recipes.postApiRecipes(opts.input as dtos_CreateRecipe);
    } catch (e) {
      handleAPIError(e);
    } finally {
      logTrpc(new Error(), opts.input, s, 'recipes.create');
    }
  }),
  update: publicProcedure
    .input(
      z.object({
        recipeID: z.string().trim(),
        payload: z.any(),
      })
    )
    .mutation(async (opts) => {
      const api = await getAPIClient();
      const s = performance.now();

      try {
        return await api.recipes.patchApiRecipes(
          opts.input.recipeID,
          opts.input.payload as dtos_UpdateRecipe
        );
      } catch (e) {
        handleAPIError(e);
      } finally {
        logTrpc(new Error(), opts.input, s, 'recipes.update');
      }
    }),
});

export type RecipesRouter = typeof recipesRouter;
