import { IRecipeContext } from './recipe-context-init';
import { models_Recipe, models_RecipeRevision } from '@4ks/api-fetch';
interface IAction {
  type: RecipeContextAction;
  payload?: any;
}

export enum RecipeContextAction {
  SET_ID = 'setRecipeId',
  SET_RECIPE = 'setRecipe',
  RESET_RECIPE = 'resetRecipeContent',
  SET_CONTROLS = 'setRecipeControls',
  SET_INGREDIENTS = 'setRecipeIngredients',
  SET_INSTRUCTIONS = 'setRecipeInstructions',
  SET_TITLE = 'setRecipeTitle',
  SET_BANNER = 'setBanner',
  SET_MEDIA = 'setMedia',
  SET_ACTION_IN_PROGRESS = 'setActionInProgress',
  SET_EDIT_IN_PROGRESS = 'setEditInProgress',
}

function getIsEditing(
  current: models_Recipe | undefined,
  recipe: models_Recipe | undefined
) {
  return JSON.stringify(current) !== JSON.stringify(recipe);
}

export function recipeContextReducer(
  state: IRecipeContext,
  action: IAction
): IRecipeContext {
  let recipe: models_Recipe;

  switch (action.type) {
    //
    case RecipeContextAction.SET_ACTION_IN_PROGRESS:
      return { ...state, actionInProgress: action.payload };
    //
    case RecipeContextAction.SET_EDIT_IN_PROGRESS:
      return { ...state, editInProgress: action.payload };
    //
    case RecipeContextAction.SET_ID:
      return { ...state, recipeId: action.payload };
    //
    case RecipeContextAction.SET_RECIPE:
      return {
        ...state,
        recipe: action.payload,
        immutableRecipe: (action.payload as models_Recipe)
          .currentRevision as models_RecipeRevision,
      };
    //
    case RecipeContextAction.RESET_RECIPE:
      return {
        ...state,
        recipe: {
          ...state.recipe,
          currentRevision: state.immutableRecipe,
        },
        editInProgress: false,
      };
    //
    case RecipeContextAction.SET_MEDIA:
      return { ...state, media: action.payload };
    //
    case RecipeContextAction.SET_CONTROLS:
      const {
        resetMedia,
        resetRecipe,
        setEditInProgress,
        setActionInProgress,
        setIngredients,
        setInstructions,
        setTitle,
        setBanner,
      } = action.payload;

      return {
        ...state,
        resetMedia,
        resetRecipe,
        setEditInProgress,
        setActionInProgress,
        setIngredients,
        setInstructions,
        setTitle,
        setBanner,
      };

    //
    case RecipeContextAction.SET_BANNER:
      recipe = {
        ...state.recipe,
        currentRevision: {
          ...state.recipe.currentRevision,
          banner: action.payload,
        },
      };
      return {
        ...state,
        editInProgress: getIsEditing(
          state.immutableRecipe,
          recipe?.currentRevision
        ),
        recipe,
      };
    //
    case RecipeContextAction.SET_TITLE:
      recipe = {
        ...state.recipe,
        currentRevision: {
          ...state.recipe.currentRevision,
          name: action.payload,
        },
      };
      return {
        ...state,
        editInProgress: getIsEditing(
          state.immutableRecipe,
          recipe?.currentRevision
        ),
        recipe,
      };
    //
    case RecipeContextAction.SET_INSTRUCTIONS:
      recipe = {
        ...state.recipe,
        currentRevision: {
          ...state.recipe.currentRevision,
          instructions: action.payload,
        },
      };
      return {
        ...state,
        editInProgress: getIsEditing(
          state.immutableRecipe,
          recipe?.currentRevision
        ),
        recipe,
      };
    //
    case RecipeContextAction.SET_INGREDIENTS:
      recipe = {
        ...state.recipe,
        currentRevision: {
          ...state.recipe.currentRevision,
          ingredients: action.payload,
        },
      };
      return {
        ...state,
        editInProgress: getIsEditing(
          state.immutableRecipe,
          recipe?.currentRevision
        ),
        recipe,
      };
    //
    default:
      throw new Error();
  }
}
