'use client';
import React, { useEffect, useContext, useReducer, useState } from 'react';
import { trpc } from '@/trpc/client';
import {
  IRecipeContext,
  makeInitialState,
  initialState,
} from './recipe-context-init';
import {
  recipeContextReducer,
  RecipeContextAction,
} from './recipe-context-reducer';
import {
  models_Recipe,
  models_Ingredient,
  models_Instruction,
  models_RecipeMediaVariant,
  models_RecipeMedia,
} from '@4ks/api-fetch';
import log from '@/libs/logger';

const RecipeContext = React.createContext<IRecipeContext>(initialState);

type RecipeContextProviderProps = {
  recipe: models_Recipe;
  media: models_RecipeMedia[];
  isAuthenticated: boolean;
  children: React.ReactNode;
};

export function RecipeContextProvider({
  recipe,
  media = [],
  isAuthenticated,
  children,
}: RecipeContextProviderProps) {
  const recipeID = `${recipe.id}`;

  const recipeData = trpc.recipes.getByIDMutation.useMutation();
  const [recipeMutex, setRecipeMutex] = useState(false);

  const mediaData = trpc.recipes.getMediaByIDMutation.useMutation();
  const [mediaMutex, setMediaMutex] = useState(false);

  const [state, dispatch] = useReducer(
    recipeContextReducer,
    makeInitialState(recipe, media)
  );

  // update media
  useEffect(() => {
    // prevent infinite loop
    if (!mediaMutex) return;

    const { isPending, isError, isSuccess, data } = mediaData;
    if (isPending || isError || !isSuccess) {
      return;
    }
    setMediaMutex(false);
    if (data?.data && state.media.length != data?.data.length) {
      dispatch({ type: RecipeContextAction.SET_MEDIA, payload: data?.data });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [mediaData, state.media]);

  // update recipe
  useEffect(() => {
    // prevent infinite loop
    if (!recipeMutex) return;

    const { isPending, isError, isSuccess, data } = recipeData;
    if (isPending || isError || !isSuccess) {
      return;
    }
    setRecipeMutex(false);
    if (data?.data && state.recipe != data?.data) {
      setEditInProgress(false);
      dispatch({ type: RecipeContextAction.SET_RECIPE, payload: data.data });
    }
  }, [recipeData, state.recipe, recipeMutex]);

  function setEditInProgress(value: boolean) {
    dispatch({
      type: RecipeContextAction.SET_EDIT_IN_PROGRESS,
      payload: value,
    });
  }

  function setActionInProgress(value: boolean) {
    dispatch({
      type: RecipeContextAction.SET_ACTION_IN_PROGRESS,
      payload: value,
    });
  }

  function setBanner(banner: Array<models_RecipeMediaVariant>) {
    dispatch({
      type: RecipeContextAction.SET_BANNER,
      payload: banner,
    });
  }

  function setIngredients(ingredients: models_Ingredient[]) {
    dispatch({
      type: RecipeContextAction.SET_INGREDIENTS,
      payload: ingredients,
    });
  }

  function setInstructions(instructions: models_Instruction[]) {
    dispatch({
      type: RecipeContextAction.SET_INSTRUCTIONS,
      payload: instructions,
    });
  }

  function setTitle(title: string) {
    dispatch({
      type: RecipeContextAction.SET_TITLE,
      payload: title,
    });
  }

  async function setRecipe() {
    setRecipeMutex(true);
    recipeData.mutate(recipeID);
  }

  async function resetRecipe() {
    dispatch({
      type: RecipeContextAction.RESET_RECIPE,
    });
  }

  async function setMedia() {
    if (!recipe.root) {
      return;
    }
    setMediaMutex(true);
    mediaData.mutate(recipe.root);
  }

  useEffect(() => {
    if (!recipe?.id) {
      return;
    }

    if (isAuthenticated) {
      dispatch({
        type: RecipeContextAction.SET_CONTROLS,
        payload: {
          resetMedia: setMedia,
          resetRecipe,
          setTitle,
          setIngredients,
          setInstructions,
          setActionInProgress,
          setEditInProgress,
          setBanner,
        },
      });
      return;
    }

    dispatch({
      type: RecipeContextAction.SET_CONTROLS,
      payload: {
        resetMedia: setMedia,
        resetRecipe: setRecipe,
        setTitle: () => {},
        setIngredients: () => {},
        setInstructions: () => {},
        setActionInProgress,
        setEditInProgress,
        setBanner,
      },
    });

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [recipe]);

  return (
    <RecipeContext.Provider value={state}>{children}</RecipeContext.Provider>
  );
}

export function useRecipeContext() {
  return useContext(RecipeContext);
}
