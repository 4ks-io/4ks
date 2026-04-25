import { vi, describe, it, expect, beforeEach } from 'vitest';

vi.mock('@4ks/api-fetch', () => ({}));

// Import after mock so the module sees the mocked dependency
import {
  makeInitialState,
  initialState,
  initialRecipe,
} from './recipe-context-init';

function makeRecipe(id = 'recipe-42', name = 'Sourdough Bread') {
  const revision = {
    author: { displayName: 'Bob', id: 'u2', username: 'bob' },
    createdDate: '2024-06-01',
    id: 'rev-10',
    ingredients: [{ name: 'flour', quantity: '500g' }],
    instructions: [{ text: 'Mix' }, { text: 'Bake' }],
    banner: [],
    name,
    recipeId: id,
    updatedDate: '2024-06-01',
  };
  return {
    author: { displayName: 'Bob', id: 'u2', username: 'bob' },
    contributors: [],
    createdDate: '2024-06-01',
    currentRevision: revision,
    id,
    metadata: { forks: 3, stars: 7 },
    root: '',
    branch: '',
    updatedDate: '2024-06-01',
  };
}

describe('initialRecipe', () => {
  it('has a default id of "0"', () => {
    expect(initialRecipe.id).toBe('0');
  });

  it('starts with empty contributors', () => {
    expect(initialRecipe.contributors).toEqual([]);
  });

  it('has zero metadata counts', () => {
    expect(initialRecipe.metadata?.forks).toBe(0);
    expect(initialRecipe.metadata?.stars).toBe(0);
  });

  it('has an empty currentRevision with no ingredients or instructions', () => {
    expect(initialRecipe.currentRevision?.ingredients).toEqual([]);
    expect(initialRecipe.currentRevision?.instructions).toEqual([]);
  });
});

describe('initialState', () => {
  it('has default recipeId of "-1"', () => {
    expect(initialState.recipeId).toBe('-1');
  });

  it('starts with actionInProgress false', () => {
    expect(initialState.actionInProgress).toBe(false);
  });

  it('starts with editInProgress false', () => {
    expect(initialState.editInProgress).toBe(false);
  });

  it('starts with an empty media array', () => {
    expect(initialState.media).toEqual([]);
  });

  it('has no-op handler functions', () => {
    expect(typeof initialState.resetRecipe).toBe('function');
    expect(typeof initialState.resetMedia).toBe('function');
    expect(typeof initialState.setTitle).toBe('function');
    expect(typeof initialState.setIngredients).toBe('function');
    expect(typeof initialState.setInstructions).toBe('function');
    expect(typeof initialState.setBanner).toBe('function');
    expect(typeof initialState.setActionInProgress).toBe('function');
    expect(typeof initialState.setEditInProgress).toBe('function');
  });
});

describe('makeInitialState', () => {
  it('sets recipeId from recipe.id', () => {
    const recipe = makeRecipe('recipe-99') as any;
    const state = makeInitialState(recipe, []);
    expect(state.recipeId).toBe('recipe-99');
  });

  it('sets recipe to the provided recipe', () => {
    const recipe = makeRecipe() as any;
    const state = makeInitialState(recipe, []);
    expect(state.recipe).toBe(recipe);
  });

  it('sets immutableRecipe to recipe.currentRevision', () => {
    const recipe = makeRecipe() as any;
    const state = makeInitialState(recipe, []);
    expect(state.immutableRecipe).toBe(recipe.currentRevision);
  });

  it('sets media to the provided array', () => {
    const recipe = makeRecipe() as any;
    const media = [{ id: 'm1' }, { id: 'm2' }] as any;
    const state = makeInitialState(recipe, media);
    expect(state.media).toBe(media);
  });

  it('converts numeric recipe id to a string for recipeId', () => {
    const recipe = { ...makeRecipe(), id: 7 as any };
    const state = makeInitialState(recipe as any, []);
    expect(state.recipeId).toBe('7');
  });

  it('preserves all handler functions from the module-level initialState', () => {
    const recipe = makeRecipe() as any;
    const state = makeInitialState(recipe, []);
    expect(typeof state.resetRecipe).toBe('function');
    expect(typeof state.resetMedia).toBe('function');
    expect(typeof state.setTitle).toBe('function');
  });
});
