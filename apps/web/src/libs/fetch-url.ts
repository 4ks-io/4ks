export class FetchURLValidationError extends Error {}

function isIPAddress(hostname: string): boolean {
  if (!hostname) {
    return false;
  }

  if (/^\d{1,3}(?:\.\d{1,3}){3}$/.test(hostname)) {
    return true;
  }

  return hostname.includes(':');
}

export function validateFetchURL(raw: string): string {
  const trimmed = raw.trim();
  if (!trimmed) {
    throw new FetchURLValidationError('Recipe URL is required');
  }

  let parsed: URL;
  try {
    parsed = new URL(trimmed);
  } catch {
    throw new FetchURLValidationError(
      'Recipe URL must be a valid absolute URL'
    );
  }

  if (parsed.protocol !== 'https:') {
    throw new FetchURLValidationError('Recipe URL must use HTTPS');
  }
  if (parsed.username || parsed.password) {
    throw new FetchURLValidationError(
      'Recipe URL must not include embedded credentials'
    );
  }

  const hostname = parsed.hostname.toLowerCase();
  if (
    !hostname ||
    hostname === 'localhost' ||
    hostname.endsWith('.localhost')
  ) {
    throw new FetchURLValidationError('Recipe URL host is not allowed');
  }
  if (isIPAddress(hostname)) {
    throw new FetchURLValidationError('Recipe URL cannot target an IP address');
  }

  return parsed.toString();
}
