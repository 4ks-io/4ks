# Local Secrets

This directory stores local-only inputs for the development Kubernetes
bootstrap. Do not commit real secrets here.

## Setup

Create your local env file from the example:

```sh
cp dev/secrets/local.env.example dev/secrets/local.env
```

Fill in the values that are shared out-of-band, especially:

- `GOOGLE_APPLICATION_CREDENTIALS`
- `AUTH0_CLIENT_SECRET`
- `AUTH0_SECRET`

`dev/bootstrap-secrets.sh` requires `GOOGLE_APPLICATION_CREDENTIALS` to point at
a local Google service account JSON file. The default path is:

```text
dev/secrets/sbx-4ks-google-app-creds.json
```

## Bootstrap

Run:

```sh
./dev/bootstrap-secrets.sh
```

Tilt runs the same script automatically before the `api` and `web` resources.

The script creates or updates these Kubernetes secrets:

- `web-local-secrets`
- `api-google-app-creds`
- `api-local-secrets`

It also generates utility config files:

- `utils/media-update/local.env`
- `utils/recipe-update/local.env`
- `utils/recipe-search-refresh/local.env`

## Optional Values

These values can stay as `replace-me` for normal local development:

- `API_FETCHER_PSK`
- `PAT_DIGEST_SECRET`
- `PAT_ENCRYPTION_SECRET`

When left unset or set to `replace-me`, the bootstrap script generates
ephemeral local values. Set explicit values only if you need stable local
Kitchen Pass or fetcher credentials across bootstrap runs.

## Git Hygiene

`.gitignore` excludes local secret files under `dev/secrets`. The committed
files in this directory should stay limited to:

- `local.env.example`
- `README.md`
