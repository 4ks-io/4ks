import { Auth0Client } from '@auth0/nextjs-auth0/server';

// Auth0 v4 reads tenant/base settings from APP_BASE_URL and AUTH0_DOMAIN.
// AUTH0_AUDIENCE remains an app-level env that we inject into the SDK config.
const authorizationParameters: {
  audience?: string;
  scope: string;
} = {
  scope: 'openid profile email offline_access',
};

if (process.env.AUTH0_AUDIENCE) {
  authorizationParameters.audience = process.env.AUTH0_AUDIENCE;
}

export const auth0 = new Auth0Client({
  authorizationParameters,
});
