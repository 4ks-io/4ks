import test from 'node:test';
import assert from 'node:assert/strict';

import { validateFetchURL } from './fetch-url';

test('validateFetchURL accepts a public https recipe url', () => {
  assert.equal(
    validateFetchURL('https://example.com/recipe/chili?serves=4'),
    'https://example.com/recipe/chili?serves=4'
  );
});

test('validateFetchURL rejects embedded credentials', () => {
  assert.throws(
    () => validateFetchURL('https://user:pass@example.com/recipe'),
    /embedded credentials/
  );
});

test('validateFetchURL rejects non-https urls', () => {
  assert.throws(
    () => validateFetchURL('http://example.com/recipe'),
    /must use HTTPS/
  );
});

test('validateFetchURL rejects localhost and ip literal targets', () => {
  assert.throws(
    () => validateFetchURL('https://localhost/recipe'),
    /host is not allowed/
  );
  assert.throws(
    () => validateFetchURL('https://127.0.0.1/recipe'),
    /cannot target an IP address/
  );
});
