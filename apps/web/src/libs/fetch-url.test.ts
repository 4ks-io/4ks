import { describe, it, expect } from 'vitest';
import { validateFetchURL, FetchURLValidationError } from './fetch-url';

describe('validateFetchURL', () => {
  describe('valid URLs', () => {
    it('accepts a public https URL', () => {
      expect(validateFetchURL('https://example.com/recipe/chili?serves=4')).toBe(
        'https://example.com/recipe/chili?serves=4'
      );
    });

    it('accepts a URL with a path and query string', () => {
      expect(validateFetchURL('https://www.allrecipes.com/recipe/16354/easy-meatloaf/')).toBe(
        'https://www.allrecipes.com/recipe/16354/easy-meatloaf/'
      );
    });

    it('trims surrounding whitespace before parsing', () => {
      expect(validateFetchURL('  https://example.com/recipe  ')).toBe(
        'https://example.com/recipe'
      );
    });

    it('returns the canonical URL string from the URL parser', () => {
      const result = validateFetchURL('https://example.com/recipe');
      expect(typeof result).toBe('string');
    });
  });

  describe('invalid URLs', () => {
    it('throws FetchURLValidationError for embedded credentials', () => {
      expect(() => validateFetchURL('https://user:pass@example.com/recipe')).toThrow(
        FetchURLValidationError
      );
      expect(() => validateFetchURL('https://user:pass@example.com/recipe')).toThrow(
        /embedded credentials/
      );
    });

    it('throws FetchURLValidationError for non-https scheme', () => {
      expect(() => validateFetchURL('http://example.com/recipe')).toThrow(
        FetchURLValidationError
      );
      expect(() => validateFetchURL('http://example.com/recipe')).toThrow(
        /must use HTTPS/
      );
    });

    it('throws FetchURLValidationError for ftp scheme', () => {
      expect(() => validateFetchURL('ftp://example.com/recipe')).toThrow(
        FetchURLValidationError
      );
    });

    it('throws FetchURLValidationError for localhost', () => {
      expect(() => validateFetchURL('https://localhost/recipe')).toThrow(
        FetchURLValidationError
      );
      expect(() => validateFetchURL('https://localhost/recipe')).toThrow(
        /host is not allowed/
      );
    });

    it('throws FetchURLValidationError for *.localhost subdomains', () => {
      expect(() => validateFetchURL('https://app.localhost/recipe')).toThrow(
        FetchURLValidationError
      );
    });

    it('throws FetchURLValidationError for IPv4 addresses', () => {
      expect(() => validateFetchURL('https://127.0.0.1/recipe')).toThrow(
        FetchURLValidationError
      );
      expect(() => validateFetchURL('https://127.0.0.1/recipe')).toThrow(
        /cannot target an IP address/
      );
    });

    it('throws FetchURLValidationError for other IPv4 addresses', () => {
      expect(() => validateFetchURL('https://192.168.1.1/recipe')).toThrow(
        FetchURLValidationError
      );
    });

    it('throws FetchURLValidationError for empty string', () => {
      expect(() => validateFetchURL('')).toThrow(FetchURLValidationError);
      expect(() => validateFetchURL('')).toThrow(/required/);
    });

    it('throws FetchURLValidationError for whitespace-only input', () => {
      expect(() => validateFetchURL('   ')).toThrow(FetchURLValidationError);
    });

    it('throws FetchURLValidationError for unparseable URLs', () => {
      expect(() => validateFetchURL('not-a-url')).toThrow(FetchURLValidationError);
      expect(() => validateFetchURL('not-a-url')).toThrow(/valid absolute URL/);
    });

    it('throws an instance of Error', () => {
      expect(() => validateFetchURL('')).toThrow(Error);
    });
  });
});
