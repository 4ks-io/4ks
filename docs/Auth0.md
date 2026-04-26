# Auth0 Notes

This file captures the Auth0 knowledge that mattered during the `019-c`
migration to `@auth0/nextjs-auth0` v4.

## Current model

- The web app uses Auth0's built-in v4 routes mounted by
  `apps/web/src/middleware.ts`.
- Login route: `/auth/login`
- Callback route: `/auth/callback`
- Logout route: `/auth/logout`
- Server auth client: `apps/web/src/libs/auth0.ts`

The old v3-style `/app/auth/*` routes are obsolete in this repo.

## Environment mapping

From Terraform and local manifests:

- Local dev host: `https://local.4ks.io`
- Dev deployed host: `https://dev.4ks.io`
- Production host: `https://www.4ks.io`

Tenant split:

- Dev tenant/domain: `dev-4ks.us.auth0.com`
  Used by `local.4ks.io` and `dev.4ks.io`
- Production tenant/domain: `4ks.us.auth0.com`
  Used by `www.4ks.io`

## Correct Auth0 Application URIs

Auth0 treats these fields differently:

- Allowed Callback URLs: multi-value list
- Allowed Logout URLs: multi-value list
- Application Login URI: conceptually a single login entrypoint, though some
  Auth0 UIs accept multiple comma-separated values

If the tenant UI forces Application Login URI to a single value, prefer the
deployed host for that application and do not treat local login as depending on
that field during normal `/auth/login` flows.

### Dev tenant application

Use these for the Auth0 application behind:

- `AUTH0_DOMAIN=dev-4ks.us.auth0.com`
- `AUTH0_CLIENT_ID=eM5Zyyfp6coLg3zORMFsZEnQmpqxBjHd`

Application Login URI:

- `https://local.4ks.io/auth/login`
- `https://dev.4ks.io/auth/login`

If only one value is allowed, prefer:

- `https://dev.4ks.io/auth/login`

Allowed Callback URLs:

- `https://local.4ks.io/auth/callback`
- `https://dev.4ks.io/auth/callback`

Allowed Logout URLs:

- `https://local.4ks.io`
- `https://dev.4ks.io`

### Production tenant application

Use these for the Auth0 application behind:

- `AUTH0_DOMAIN=4ks.us.auth0.com`
- `AUTH0_CLIENT_ID=vAUr50Saqug9Mf3Yu4cFvAaT2nsgNLIN`

Application Login URI:

- `https://www.4ks.io/auth/login`

Allowed Callback URLs:

- `https://www.4ks.io/auth/callback`

Allowed Logout URLs:

- `https://www.4ks.io`

## What should be removed

These are stale for the v4 setup and should not remain as the canonical values:

- `/app/auth/callback`
- `/app/auth/logout`
- `/login` as a callback URL
- `/` as a callback URL
- `/authback` logout URLs unless some external legacy flow still uses them

## Why `/auth/login` is the right Application Login URI

Auth0's Application Login URI must point to a route in the app that starts the
authentication flow. In this repo, that route is the SDK-mounted `/auth/login`
handler, not a custom `/login` page.

`/auth/login` is also where the app sends `returnTo`, for example:

```txt
/auth/login?returnTo=/recipe/<id>-slug
```

That is what allows login from a recipe page to return the user to the same
recipe after callback.

## Important repo-specific behavior

### Post-auth return to current page

The repo does not rely on Auth0 defaults for this. The current implementation:

- builds `/auth/login?returnTo=...` on client-triggered logins
- preserves the current request path in middleware via `x-url-pathname`
- uses that path for server-side redirects on protected pages

Relevant files:

- [apps/web/src/libs/navigation.ts](/code/4ks-io/4ks/apps/web/src/libs/navigation.ts)
- [apps/web/src/libs/server/navigation.ts](/code/4ks-io/4ks/apps/web/src/libs/server/navigation.ts)
- [apps/web/src/middleware.ts](/code/4ks-io/4ks/apps/web/src/middleware.ts)

### APP_BASE_URL and AUTH0_DOMAIN

Auth0 v4 uses:

- `APP_BASE_URL`
- `AUTH0_DOMAIN`
- `AUTH0_CLIENT_ID`
- `AUTH0_CLIENT_SECRET`
- `AUTH0_SECRET`

This repo still uses `AUTH0_AUDIENCE` as an app env var, but it is injected
into `authorizationParameters` in `apps/web/src/libs/auth0.ts`. It is no longer
an Auth0 SDK top-level environment variable.

### Local validation gotcha

Local login on `https://local.4ks.io` will fail with:

- `unauthorized_client: Callback URL mismatch`

unless the dev tenant app explicitly includes:

- `https://local.4ks.io/auth/callback`

## Known warnings

`@auth0/nextjs-auth0@4.19.0` emits webpack warnings from `dpopUtils.js`:

- `Critical dependency: the request of a dependency is an expression`

In this repo that warning has been non-blocking.

## Local dev stability note

After dependency changes, long-lived Tilt web containers can retain stale
`.next` output and produce missing vendor chunk errors. The local Tilt flow now
clears `apps/web/.next` before reinstalling on package-manifest changes.
