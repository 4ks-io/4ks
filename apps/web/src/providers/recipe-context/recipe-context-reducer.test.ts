import { vi, describe, it, expect } from 'vitest';

vi.mock('@4ks/api-fetch', () => ({}));

import {
  recipeContextReducer,
  RecipeContextAction,
} from './recipe-context-reducer';
import type { IRecipeContext } from './recipe-context-init';

// ---------------------------------------------------------------------------
// Minimal fixtures
// ---------------------------------------------------------------------------

function makeRevision(overrides: Record<string, any> = {}) {
  return {
    author: { displayName: 'Alice', id: 'u1', username: 'alice' },
    createdDate: '2024-01-01',
    id: 'rev-1',
    ingredients: [],
    instructions: [],
    banner: [],
    name: 'Test Recipe',
    recipeId: 'recipe-1',
    updatedDate: '2024-01-01',
    ...overrides,
  };
}

function makeRecipe(revisionOverrides: Record<string, any> = {}) {
  return {
    author: { displayName: 'Alice', id: 'u1', username: 'alice' },
    contributors: [],
    createdDate: '2024-01-01',
    currentRevision: makeRevision(revisionOverrides),
    id: 'recipe-1',
    metadata: { forks: 0, stars: 0 },
    root: '',
    branch: '',
    updatedDate: '2024-01-01',
  };
}

function makeBaseState(overrides: Partial<IRecipeContext> = {}): IRecipeContext {
  const recipe = makeRecipe();
  return {
    recipeId: 'recipe-1',
    recipe,
    immutableRecipe: recipe.currentRevision,
    media: [],
    resetMedia: () => {},
    resetRecipe: () => {},
    setTitle: () => {},
    setIngredients: () => {},
    setInstructions: () => {},
    setBanner: () => {},
    actionInProgress: false,
    setActionInProgress: () => {},
    editInProgress: false,
    setEditInProgress: () => {},
    ...overrides,
  } as unknown as IRecipeContext;
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('recipeContextReducer', () => {
  describe('SET_ACTION_IN_PROGRESS', () => {
    it('sets actionInProgress to true', () => {
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_ACTION_IN_PROGRESS,
        payload: true,
      });
      expect(result.actionInProgress).toBe(true);
    });

    it('sets actionInProgress to false', () => {
      const state = makeBaseState({ actionInProgress: true } as any);
      const result = recipeContextReducer(state, {
        type: RecipeContextAction.SET_ACTION_IN_PROGRESS,
        payload: false,
      });
      expect(result.actionInProgress).toBe(false);
    });

    it('does not change other fields', () => {
      const state = makeBaseState();
      const result = recipeContextReducer(state, {
        type: RecipeContextAction.SET_ACTION_IN_PROGRESS,
        payload: true,
      });
      expect(result.recipeId).toBe(state.recipeId);
      expect(result.editInProgress).toBe(state.editInProgress);
    });
  });

  describe('SET_EDIT_IN_PROGRESS', () => {
    it('sets editInProgress to true', () => {
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_EDIT_IN_PROGRESS,
        payload: true,
      });
      expect(result.editInProgress).toBe(true);
    });

    it('sets editInProgress to false', () => {
      const state = makeBaseState({ editInProgress: true } as any);
      const result = recipeContextReducer(state, {
        type: RecipeContextAction.SET_EDIT_IN_PROGRESS,
        payload: false,
      });
      expect(result.editInProgress).toBe(false);
    });
  });

  describe('SET_ID', () => {
    it('updates recipeId', () => {
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_ID,
        payload: 'recipe-99',
      });
      expect(result.recipeId).toBe('recipe-99');
    });

    it('preserves all other state', () => {
      const state = makeBaseState();
      const result = recipeContextReducer(state, {
        type: RecipeContextAction.SET_ID,
        payload: 'new-id',
      });
      expect(result.actionInProgress).toBe(state.actionInProgress);
      expect(result.editInProgress).toBe(state.editInProgress);
    });
  });

  describe('SET_RECIPE', () => {
    it('sets recipe to the payload', () => {
      const newRecipe = makeRecipe({ name: 'New Name' });
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_RECIPE,
        payload: newRecipe,
      });
      expect(result.recipe).toBe(newRecipe);
    });

    it('sets immutableRecipe to the currentRevision of the payload', () => {
      const newRecipe = makeRecipe({ name: 'Locked In' });
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_RECIPE,
        payload: newRecipe,
      });
      expect(result.immutableRecipe).toBe(newRecipe.currentRevision);
    });
  });

  describe('RESET_RECIPE', () => {
    it('restores currentRevision from immutableRecipe', () => {
      const originalRevision = makeRevision({ name: 'Original' });
      const recipe = { ...makeRecipe(), currentRevision: makeRevision({ name: 'Edited' }) };
      const state = makeBaseState({ recipe, immutableRecipe: originalRevision } as any);

      const result = recipeContextReducer(state, {
        type: RecipeContextAction.RESET_RECIPE,
      });

      expect(result.recipe.currentRevision).toBe(originalRevision);
    });

    it('sets editInProgress to false', () => {
      const state = makeBaseState({ editInProgress: true } as any);
      const result = recipeContextReducer(state, {
        type: RecipeContextAction.RESET_RECIPE,
      });
      expect(result.editInProgress).toBe(false);
    });
  });

  describe('SET_MEDIA', () => {
    it('replaces the media array', () => {
      const newMedia = [{ id: 'm1' }, { id: 'm2' }] as any;
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_MEDIA,
        payload: newMedia,
      });
      expect(result.media).toBe(newMedia);
    });

    it('can set media to an empty array', () => {
      const state = makeBaseState({ media: [{ id: 'm1' }] } as any);
      const result = recipeContextReducer(state, {
        type: RecipeContextAction.SET_MEDIA,
        payload: [],
      });
      expect(result.media).toEqual([]);
    });
  });

  describe('SET_CONTROLS', () => {
    it('updates all handler functions from the payload', () => {
      const newHandlers = {
        resetMedia: vi.fn(),
        resetRecipe: vi.fn(),
        setEditInProgress: vi.fn(),
        setActionInProgress: vi.fn(),
        setIngredients: vi.fn(),
        setInstructions: vi.fn(),
        setTitle: vi.fn(),
        setBanner: vi.fn(),
      };

      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_CONTROLS,
        payload: newHandlers,
      });

      expect(result.resetMedia).toBe(newHandlers.resetMedia);
      expect(result.resetRecipe).toBe(newHandlers.resetRecipe);
      expect(result.setEditInProgress).toBe(newHandlers.setEditInProgress);
      expect(result.setActionInProgress).toBe(newHandlers.setActionInProgress);
      expect(result.setIngredients).toBe(newHandlers.setIngredients);
      expect(result.setInstructions).toBe(newHandlers.setInstructions);
      expect(result.setTitle).toBe(newHandlers.setTitle);
      expect(result.setBanner).toBe(newHandlers.setBanner);
    });
  });

  describe('SET_TITLE', () => {
    it('updates the recipe name in currentRevision', () => {
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_TITLE,
        payload: 'Spicy Ramen',
      });
      expect(result.recipe.currentRevision?.name).toBe('Spicy Ramen');
    });

    it('sets editInProgress to true when the title differs from the immutable revision', () => {
      const state = makeBaseState();
      const result = recipeContextReducer(state, {
        type: RecipeContextAction.SET_TITLE,
        payload: 'Different Title',
      });
      expect(result.editInProgress).toBe(true);
    });

    it('sets editInProgress to false when the title matches the immutable revision', () => {
      // immutableRecipe.name = 'Test Recipe' (from makeRevision default)
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_TITLE,
        payload: 'Test Recipe',
      });
      expect(result.editInProgress).toBe(false);
    });

    it('does not mutate the original recipe object', () => {
      const state = makeBaseState();
      const originalRevision = state.recipe.currentRevision;
      recipeContextReducer(state, {
        type: RecipeContextAction.SET_TITLE,
        payload: 'New Title',
      });
      expect(originalRevision?.name).toBe('Test Recipe');
    });
  });

  describe('SET_BANNER', () => {
    it('updates the banner in currentRevision', () => {
      const newBanner = [{ alias: 'md', url: 'https://cdn.example.com/img.jpg' }] as any;
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_BANNER,
        payload: newBanner,
      });
      expect(result.recipe.currentRevision?.banner).toBe(newBanner);
    });

    it('sets editInProgress to true when banner changes', () => {
      const newBanner = [{ alias: 'md', url: 'https://cdn.example.com/new.jpg' }] as any;
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_BANNER,
        payload: newBanner,
      });
      expect(result.editInProgress).toBe(true);
    });

    it('sets editInProgress to false when banner matches immutable revision', () => {
      // immutableRecipe.banner = [] (default)
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_BANNER,
        payload: [],
      });
      expect(result.editInProgress).toBe(false);
    });
  });

  describe('SET_INGREDIENTS', () => {
    it('updates ingredients in currentRevision', () => {
      const newIngredients = [{ name: 'flour', quantity: '2 cups' }] as any;
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_INGREDIENTS,
        payload: newIngredients,
      });
      expect(result.recipe.currentRevision?.ingredients).toBe(newIngredients);
    });

    it('sets editInProgress to true when ingredients differ from immutable revision', () => {
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_INGREDIENTS,
        payload: [{ name: 'sugar' }],
      });
      expect(result.editInProgress).toBe(true);
    });

    it('sets editInProgress to false when ingredients match immutable revision', () => {
      // immutableRecipe.ingredients = [] (default)
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_INGREDIENTS,
        payload: [],
      });
      expect(result.editInProgress).toBe(false);
    });
  });

  describe('SET_INSTRUCTIONS', () => {
    it('updates instructions in currentRevision', () => {
      const newInstructions = [{ text: 'Preheat oven to 350°F' }] as any;
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_INSTRUCTIONS,
        payload: newInstructions,
      });
      expect(result.recipe.currentRevision?.instructions).toBe(newInstructions);
    });

    it('sets editInProgress to true when instructions differ', () => {
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_INSTRUCTIONS,
        payload: [{ text: 'New step' }],
      });
      expect(result.editInProgress).toBe(true);
    });

    it('sets editInProgress to false when instructions match immutable revision', () => {
      // immutableRecipe.instructions = [] (default)
      const result = recipeContextReducer(makeBaseState(), {
        type: RecipeContextAction.SET_INSTRUCTIONS,
        payload: [],
      });
      expect(result.editInProgress).toBe(false);
    });
  });

  describe('unknown action', () => {
    it('throws on an unrecognised action type', () => {
      expect(() =>
        recipeContextReducer(makeBaseState(), { type: 'UNKNOWN' as any })
      ).toThrow();
    });
  });

  describe('immutability', () => {
    it('always returns a new state object', () => {
      const state = makeBaseState();
      const next = recipeContextReducer(state, {
        type: RecipeContextAction.SET_ID,
        payload: 'new',
      });
      expect(next).not.toBe(state);
    });

    it('does not mutate the input state', () => {
      const state = makeBaseState();
      const originalId = state.recipeId;
      recipeContextReducer(state, {
        type: RecipeContextAction.SET_ID,
        payload: 'mutated',
      });
      expect(state.recipeId).toBe(originalId);
    });
  });
});
