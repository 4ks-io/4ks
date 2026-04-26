# Secret Rotation

All runtime secrets are stored in Google Secret Manager and injected natively
into Cloud Run and Cloud Functions. No secret values pass through Terraform.

## How versioning works

Each secret has `version_destroy_ttl = 86400s` (24 h). The version lifecycle is:

```
ENABLED → DISABLED → DESTROYED (auto, after 24 h)
```

Adding a new version does **not** automatically disable the old one — that is a
manual step. Disabling starts the 24 h countdown. This keeps active versions at
1 per secret at steady state, staying within the Secret Manager free tier.

## Secrets inventory

| Secret Manager ID          | Project   | Consumers                          |
|----------------------------|-----------|------------------------------------|
| `auth0-client-secret`      | prd-4ks   | Cloud Run `web`                    |
| `auth0-session-secret`     | prd-4ks   | Cloud Run `web`                    |
| `typesense-api-key`        | prd-4ks   | Cloud Run `api`                    |
| `typesense-search-api-key` | prd-4ks   | Cloud Run `web`                    |
| `api-fetcher-psk`          | prd-4ks   | Cloud Run `api`, Cloud Function `fetcher` |

## General rotation procedure

For any secret:

**1. Add the new version**

```sh
gcloud secrets versions add <secret-id> \
  --data-file=<(echo -n "$NEW_VALUE") \
  --project=<project>
```

**2. Verify the new version is accessible**

```sh
gcloud secrets versions access latest \
  --secret=<secret-id> \
  --project=<project>
```

**3. Force consumers to redeploy** (Cloud Run fetches secrets at cold start)

```sh
# Cloud Run service:
gcloud run services update <service-name> \
  --region=us-east4 --project=<project> \
  --update-env-vars=_REDEPLOY=$(date +%s)

# Cloud Function (redeploy triggers new instance):
gcloud functions deploy <function-name> \
  --region=us-east4 --project=<project> \
  --gen2
```

**4. Verify the consumer is healthy**

Check Cloud Run revision health and test the affected feature.

**5. Disable the old version**

```sh
# List versions to find the previous version number:
gcloud secrets versions list <secret-id> --project=<project>

# Disable it (starts the 24 h destroy countdown):
gcloud secrets versions disable <version-number> \
  --secret=<secret-id> \
  --project=<project>
```

**6. Clean up the temporary env var**

```sh
gcloud run services update <service-name> \
  --region=us-east4 --project=<project> \
  --remove-env-vars=_REDEPLOY
```

---

## Secret-specific notes

### `auth0-client-secret`

Rotate in the Auth0 dashboard: **Applications → [app] → Settings → Rotate**.
Auth0 generates a new secret automatically. Copy the new value before leaving
the page.

Consumer: `web`. Test by completing a login flow end-to-end after redeployment.

### `auth0-session-secret`

A random string used to sign session cookies. Rotating **immediately invalidates
all active user sessions** — everyone is logged out. Schedule during low-traffic
hours.

Generate a new value:

```sh
openssl rand -hex 32
```

Consumer: `web`. After redeployment, verify login works. Expect users to be
signed out.

### `typesense-api-key`

Admin key with full Typesense access. Rotate in the Typesense Cloud dashboard
or via the Typesense API, then revoke the old key in Typesense after disabling
the Secret Manager version.

Consumer: `api`. Test by triggering a search or index operation.

### `typesense-search-api-key`

Search-only key scoped to read operations. Rotate in the Typesense Cloud
dashboard and revoke the old key after the rotation is confirmed.

Consumer: `web`. Test by using search in the UI.

### `api-fetcher-psk`

Shared secret between the `api` service and the `fetcher` function. Both must
pick up the new value before the old version is disabled — otherwise fetcher
requests to the API will be rejected with 401.

Generate a new value:

```sh
openssl rand -hex 32
```

Redeploy **both** consumers before disabling the old version:

```sh
gcloud run services update api \
  --region=us-east4 --project=prd-4ks \
  --update-env-vars=_REDEPLOY=$(date +%s)

gcloud functions deploy fetcher \
  --region=us-east4 --project=prd-4ks \
  --gen2
```

Verify by triggering a recipe fetch and checking Cloud Logging for PSK errors.

---

## Verifying active version count

To confirm you are within the 6-version free tier:

```sh
PROJECT=prd-4ks

for s in auth0-client-secret auth0-session-secret typesense-api-key typesense-search-api-key api-fetcher-psk; do
  count=$(gcloud secrets versions list $s --project=$PROJECT \
    --filter="state=ENABLED" --format="value(name)" | wc -l)
  echo "$s: $count active version(s)"
done
```

Each secret should show `1 active version(s)` at steady state.
