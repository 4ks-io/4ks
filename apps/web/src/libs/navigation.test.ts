import { describe, it, expect } from 'vitest';
import {
  authLoginPath,
  authLogoutPath,
  isSSR,
  Page,
  normalizeForURL,
} from './navigation';

describe('auth path constants', () => {
  it('authLoginPath points to the Auth0 login handler', () => {
    expect(authLoginPath).toBe('/app/auth/login');
  });

  it('authLogoutPath points to the Auth0 logout handler', () => {
    expect(authLogoutPath).toBe('/app/auth/logout');
  });
});

describe('isSSR', () => {
  it('is true in a Node.js (non-browser) test environment', () => {
    // window is undefined in Vitest's node environment
    expect(isSSR).toBe(true);
  });
});

describe('Page enum', () => {
  it('defines LANDING, REGISTER, AUTHENTICATED, and ANONYMOUS', () => {
    expect(Page.LANDING).toBeDefined();
    expect(Page.REGISTER).toBeDefined();
    expect(Page.AUTHENTICATED).toBeDefined();
    expect(Page.ANONYMOUS).toBeDefined();
  });

  it('all four values are distinct', () => {
    const values = [Page.LANDING, Page.REGISTER, Page.AUTHENTICATED, Page.ANONYMOUS];
    expect(new Set(values).size).toBe(4);
  });

  it('values are numbers (numeric enum)', () => {
    expect(typeof Page.LANDING).toBe('number');
    expect(typeof Page.REGISTER).toBe('number');
    expect(typeof Page.AUTHENTICATED).toBe('number');
    expect(typeof Page.ANONYMOUS).toBe('number');
  });
});

describe('normalizeForURL', () => {
  describe('falsy input', () => {
    it('returns "recipe-title" for undefined', () => {
      expect(normalizeForURL(undefined)).toBe('recipe-title');
    });

    it('returns "recipe-title" for an empty string', () => {
      expect(normalizeForURL('')).toBe('recipe-title');
    });
  });

  describe('basic normalization', () => {
    it('lowercases the input', () => {
      expect(normalizeForURL('HELLO WORLD')).toBe('hello-world');
    });

    it('trims leading and trailing whitespace', () => {
      expect(normalizeForURL('  hello  ')).toBe('hello');
    });

    it('replaces spaces with hyphens', () => {
      expect(normalizeForURL('hello world')).toBe('hello-world');
    });

    it('passes through a simple ascii slug unchanged', () => {
      expect(normalizeForURL('chicken-tikka-masala')).toBe('chicken-tikka-masala');
    });

    it('preserves digits', () => {
      expect(normalizeForURL('recipe 42')).toBe('recipe-42');
    });
  });

  describe('special character removal', () => {
    it('replaces punctuation with hyphens', () => {
      // Apostrophes and other punctuation become hyphens, then get collapsed/stripped
      expect(normalizeForURL("chili!")).toBe('chili');
      // Apostrophe is a non-alphanumeric character → becomes a hyphen separator
      expect(normalizeForURL("mom's chili")).toBe('mom-s-chili');
    });

    it('collapses consecutive non-alphanumeric characters into a single hyphen', () => {
      expect(normalizeForURL('hello   world')).toBe('hello-world');
      expect(normalizeForURL('hello -- world')).toBe('hello-world');
      expect(normalizeForURL('a!@#b')).toBe('a-b');
    });

    it('strips leading hyphens that result from leading special characters', () => {
      expect(normalizeForURL('!hello')).toBe('hello');
      expect(normalizeForURL('---hello')).toBe('hello');
    });

    it('strips trailing hyphens that result from trailing special characters', () => {
      expect(normalizeForURL('hello!')).toBe('hello');
      expect(normalizeForURL('hello---')).toBe('hello');
    });

    it('strips both leading and trailing hyphens', () => {
      expect(normalizeForURL('---hello---')).toBe('hello');
    });
  });

  describe('i18n character normalization', () => {
    it('converts é to e', () => {
      expect(normalizeForURL('café')).toBe('cafe');
    });

    it('converts ü to u', () => {
      expect(normalizeForURL('über')).toBe('uber');
    });

    it('converts ï to i', () => {
      expect(normalizeForURL('naïve')).toBe('naive');
    });

    it('converts ñ to n', () => {
      expect(normalizeForURL('jalapeño')).toBe('jalapeno');
    });

    it('converts a mix of accented characters', () => {
      expect(normalizeForURL('Crème brûlée')).toBe('creme-brulee');
    });

    it('handles fully accented words', () => {
      expect(normalizeForURL('Ñoño')).toBe('nono');
    });
  });
});
