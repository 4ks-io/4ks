export const authLoginPath = '/auth/login';
export const authLogoutPath = '/auth/logout';

export const isSSR = typeof window === 'undefined';

export function buildAuthLoginPath(returnTo?: string) {
  if (!returnTo) {
    return authLoginPath;
  }

  const params = new URLSearchParams({
    returnTo,
  });

  return `${authLoginPath}?${params.toString()}`;
}

export function navigateToAuthRoute(path: string) {
  if (typeof window !== 'undefined') {
    window.location.assign(path);
  }
}

export function navigateToLogin(returnTo?: string) {
  const currentPath =
    returnTo ??
    (typeof window !== 'undefined'
      ? `${window.location.pathname}${window.location.search}${window.location.hash}`
      : undefined);

  navigateToAuthRoute(buildAuthLoginPath(currentPath));
}

export const isMobileDevice = () => {
  if (typeof navigator === 'undefined') {
    return false;
  }

  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
    navigator.userAgent
  );
};

// tr@ck: deprecated
// https://stackoverflow.com/questions/10527983/best-way-to-detect-mac-os-x-or-windows-computers-with-javascript-or-jquery
export const isMac = () => {
  if (typeof navigator !== 'undefined') {
    // console.log('navigator.platform', navigator.platform);
    return navigator.platform.indexOf('Mac') > -1;
  }
  return false;
};

export enum Page {
  LANDING,
  REGISTER,
  // any page requiring authentication and authorization
  AUTHENTICATED,
  // any page available to anonymous users
  ANONYMOUS,
}

export type PageProps = {
  params: Promise<PagePropsParams>;
  searchParams?: Promise<PagePropsSearchParams>;
};

export type RecipePropsParams = { id: string };
export type PagePropsParams = { slug: string };
export type PagePropsSearchParams = {
  [key: string]: string | string[] | undefined;
};

export function normalizeForURL(s: string | undefined) {
  if (!s) {
    // tr@ck: i18n
    return 'recipe-title';
  }
  // Convert i18n characters like é, ü, etc. to their base
  const normalizedStr = s
    .toLowerCase()
    .trim()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '');

  // Replace spaces with hyphens, remove special characters, and convert to lowercase
  return normalizedStr.replace(/[^a-z0-9]+/gi, '-').replace(/^-+|-+$/g, '');
}
