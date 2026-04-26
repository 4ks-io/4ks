# Local Kubernetes Secrets

Files in this directory are for local development. Keep committed manifests on
local placeholders only.

Populate `dev/secrets/local.env` with the shared local values, then bootstrap
the local Kubernetes secrets and utility config files:

```sh
cp dev/secrets/local.env.example dev/secrets/local.env
./dev/bootstrap-secrets.sh
```

Tilt invokes the same bootstrap step automatically before the `api` and `web`
resources.

Auth0 local tenant setup is also required. The Auth0 application used by
`AUTH0_CLIENT_ID` must include these exact URLs:

- Allowed Callback URLs: `https://local.4ks.io/auth/callback`
- Allowed Logout URLs: `https://local.4ks.io`

Generated secret manifests matching `dev/deploy/*secret*.yaml` are ignored.
