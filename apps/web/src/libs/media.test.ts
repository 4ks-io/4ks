import { describe, it, expect } from 'vitest';
import { RecipeMediaSize, getRecipeBannerVariantUrl } from './media';

// Minimal stand-in for models_RecipeMediaVariant
type MockVariant = { alias: string; url: string };

const SM: MockVariant = { alias: 'sm', url: 'https://cdn.example.com/img-sm.jpg' };
const MD: MockVariant = { alias: 'md', url: 'https://cdn.example.com/img-md.jpg' };
const LG: MockVariant = { alias: 'lg', url: 'https://cdn.example.com/img-lg.jpg' };
const allVariants = [SM, MD, LG] as any;

describe('RecipeMediaSize', () => {
  it('SM has value "sm"', () => {
    expect(RecipeMediaSize.SM).toBe('sm');
  });

  it('MD has value "md"', () => {
    expect(RecipeMediaSize.MD).toBe('md');
  });

  it('LG has value "lg"', () => {
    expect(RecipeMediaSize.LG).toBe('lg');
  });

  it('all values are distinct strings', () => {
    const values = Object.values(RecipeMediaSize);
    expect(new Set(values).size).toBe(values.length);
  });
});

describe('getRecipeBannerVariantUrl', () => {
  describe('when variants is undefined', () => {
    it('returns undefined', () => {
      expect(getRecipeBannerVariantUrl(undefined)).toBeUndefined();
      expect(getRecipeBannerVariantUrl(undefined, RecipeMediaSize.SM)).toBeUndefined();
      expect(getRecipeBannerVariantUrl(undefined, RecipeMediaSize.LG)).toBeUndefined();
    });
  });

  describe('when variants is an empty array', () => {
    it('returns undefined', () => {
      expect(getRecipeBannerVariantUrl([])).toBeUndefined();
    });
  });

  describe('with a full set of variants', () => {
    it('returns the SM variant when size is SM', () => {
      expect(getRecipeBannerVariantUrl(allVariants, RecipeMediaSize.SM)).toEqual(SM);
    });

    it('returns the MD variant when size is MD', () => {
      expect(getRecipeBannerVariantUrl(allVariants, RecipeMediaSize.MD)).toEqual(MD);
    });

    it('returns the LG variant when size is LG', () => {
      expect(getRecipeBannerVariantUrl(allVariants, RecipeMediaSize.LG)).toEqual(LG);
    });

    it('defaults to MD when no size argument is provided', () => {
      expect(getRecipeBannerVariantUrl(allVariants)).toEqual(MD);
    });
  });

  describe('when the requested size is absent', () => {
    it('returns undefined when only SM and MD are available and LG is requested', () => {
      const partial = [SM, MD] as any;
      expect(getRecipeBannerVariantUrl(partial, RecipeMediaSize.LG)).toBeUndefined();
    });

    it('returns the first match when multiple variants share the same alias', () => {
      const duplicate = [
        { alias: 'md', url: 'https://cdn.example.com/first.jpg' },
        { alias: 'md', url: 'https://cdn.example.com/second.jpg' },
      ] as any;
      expect(getRecipeBannerVariantUrl(duplicate, RecipeMediaSize.MD)).toEqual({
        alias: 'md',
        url: 'https://cdn.example.com/first.jpg',
      });
    });
  });

  describe('alias matching is strict', () => {
    it('does not match variants with an alias that differs only in case', () => {
      const upperCase = [{ alias: 'MD', url: 'https://cdn.example.com/img.jpg' }] as any;
      expect(getRecipeBannerVariantUrl(upperCase, RecipeMediaSize.MD)).toBeUndefined();
    });
  });
});
