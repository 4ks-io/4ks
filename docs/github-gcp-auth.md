# GitHub Actions — GCP Authentication

GitHub Actions authenticates to GCP using Workload Identity Federation (WIF).
Short-lived OIDC tokens minted per workflow run are exchanged for GCP
credentials to impersonate a tightly scoped service account. No long-lived JSON
keys are stored in GitHub.

## Architecture

```
GitHub Actions workflow run
  │
  ├─ mints OIDC token (id-token: write permission)
  │
  └─▶ GCP Workload Identity Pool (eng-4ks, global)
        └─▶ GitHub OIDC Provider
              └─▶ service account impersonation
                    ├─ github-build-eng@eng-4ks        (build / publish)
                    ├─ github-terraform-dev@dev-4ks    (dev Terraform)
                    └─ github-terraform-prd@prd-4ks    (prd Terraform)
```

The WIF pool and OIDC provider live in `eng-4ks`. All service accounts in all
projects are bound to this single pool.

## Workload Identity Pool and Provider

**Pool**

| Field | Value |
|---|---|
| Project | `eng-4ks` |
| Pool ID | `github` |
| Location | `global` |

**Provider**

| Field | Value |
|---|---|
| Provider ID | `github` |
| Issuer | `https://token.actions.githubusercontent.com` |
| Attribute condition | `assertion.repository == '4ks-io/4ks' && assertion.repository_owner == '4ks-io' && assertion.ref == 'refs/heads/main'` |

The `ref` condition restricts impersonation to pushes on `main`. If a PR-plan
path is reintroduced, add a second provider without the `ref` condition and bind
a less-privileged service account to it.

**Attribute mapping**

```
google.subject            = assertion.sub
attribute.actor           = assertion.actor
attribute.repository      = assertion.repository
attribute.repository_owner = assertion.repository_owner
attribute.ref             = assertion.ref
attribute.workflow        = assertion.workflow
```

## Service Accounts

### `github-build-eng` (eng-4ks)

Used by `build-container.yaml` to push container images to Artifact Registry.

| Role | Scope | Reason |
|---|---|---|
| `roles/artifactregistry.writer` | `eng-4ks` project | Push images to Artifact Registry |
| `roles/iam.workloadIdentityUser` | service account | Allow WIF pool to impersonate |

### `github-terraform-dev` (dev-4ks)

Used by `run-terraform.yaml` when `environment: dev`. Manages all `iac/app`
resources in the dev project.

| Role | Scope | Reason |
|---|---|---|
| `roles/run.admin` | `dev-4ks` project | Manage Cloud Run services |
| `roles/iam.serviceAccountUser` | `dev-4ks` project | Act as runtime service accounts |
| `roles/iam.roleAdmin` | `dev-4ks` project | Manage custom IAM roles defined in `iac/app` |
| `roles/artifactregistry.reader` | `eng-4ks` project | Validate image references when updating Cloud Run (images live in `eng-4ks`) |
| `roles/cloudfunctions.admin` | `dev-4ks` project | Manage Cloud Functions v2 (media-upload) |
| `roles/compute.loadBalancerAdmin` | `dev-4ks` project | Manage load balancer resources |
| `roles/dns.admin` | `dev-4ks` project | Manage Cloud DNS records |
| `roles/certificatemanager.editor` | `dev-4ks` project | Manage TLS certificates |
| `roles/pubsub.admin` | `dev-4ks` project | Manage Pub/Sub topics and subscriptions (fetcher) |
| `roles/resourcemanager.projectIamAdmin` | `dev-4ks` project | Set project-level IAM bindings for runtime service accounts |
| `roles/storage.admin` | `dev-4ks` project | Manage GCS buckets and objects |
| `roles/secretmanager.admin` | `dev-4ks` project | Manage secrets |
| `roles/iam.workloadIdentityUser` | service account | Allow WIF pool to impersonate |

### `github-terraform-prd` (prd-4ks)

Used by `run-terraform.yaml` when `environment: prd`. Manages all `iac/app`
resources in the production project. Same role set as dev — same Terraform
codebase, different project.

| Role | Scope | Reason |
|---|---|---|
| `roles/run.admin` | `prd-4ks` project | Manage Cloud Run services |
| `roles/iam.serviceAccountUser` | `prd-4ks` project | Act as runtime service accounts |
| `roles/iam.roleAdmin` | `prd-4ks` project | Manage custom IAM roles defined in `iac/app` |
| `roles/artifactregistry.reader` | `eng-4ks` project | Validate image references when updating Cloud Run (images live in `eng-4ks`) |
| `roles/cloudfunctions.admin` | `prd-4ks` project | Manage Cloud Functions v2 (media-upload) |
| `roles/compute.loadBalancerAdmin` | `prd-4ks` project | Manage load balancer resources |
| `roles/dns.admin` | `prd-4ks` project | Manage Cloud DNS records |
| `roles/certificatemanager.editor` | `prd-4ks` project | Manage TLS certificates |
| `roles/pubsub.admin` | `prd-4ks` project | Manage Pub/Sub topics and subscriptions (fetcher) |
| `roles/resourcemanager.projectIamAdmin` | `prd-4ks` project | Set project-level IAM bindings for runtime service accounts |
| `roles/storage.admin` | `prd-4ks` project | Manage GCS buckets and objects |
| `roles/secretmanager.admin` | `prd-4ks` project | Manage secrets |
| `roles/iam.workloadIdentityUser` | service account | Allow WIF pool to impersonate |

## Required GCP APIs

The following APIs must be enabled on each project before WIF impersonation works.

| API | Project(s) | Reason |
|---|---|---|
| `iamcredentials.googleapis.com` | `eng-4ks`, `dev-4ks`, `prd-4ks` | Required for service account impersonation via WIF |

```sh
gcloud services enable iamcredentials.googleapis.com --project="eng-4ks"
gcloud services enable iamcredentials.googleapis.com --project="dev-4ks"
gcloud services enable iamcredentials.googleapis.com --project="prd-4ks"
```

## GitHub Repository Variables

Stored as repository-level variables (not environment-specific) under
**Settings → Secrets and variables → Actions → Variables**.

| Variable | Value |
|---|---|
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | `projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/github/providers/github` |
| `GCP_BUILD_SERVICE_ACCOUNT` | `github-build-eng@eng-4ks.iam.gserviceaccount.com` |
| `GCP_DEV_TERRAFORM_SERVICE_ACCOUNT` | `github-terraform-dev@dev-4ks.iam.gserviceaccount.com` |
| `GCP_PRD_TERRAFORM_SERVICE_ACCOUNT` | `github-terraform-prd@prd-4ks.iam.gserviceaccount.com` |

`GCP_BUILD_SERVICE_ACCOUNT` and `GCP_WORKLOAD_IDENTITY_PROVIDER` are shared
across all jobs and must stay at repo level. The service account variables will
move to GitHub environment scope if story 030 is implemented.

## Workflow Permissions

Any workflow job that calls `google-github-actions/auth@v3` must declare:

```yaml
permissions:
  contents: read
  id-token: write
```

The `id-token: write` permission allows the runner to request an OIDC token
from GitHub. Without it the WIF exchange fails before reaching GCP.

Reusable workflows that perform GCP auth (`build-container.yaml`,
`run-terraform.yaml`) declare this permission at the job level. The caller
`main.yaml` also declares it at the workflow level to propagate it correctly
to reusable workflow calls.

## Deferred Service Account

`github-terraform-eng@eng-4ks.iam.gserviceaccount.com` is not yet created.
It will be used when `iac/eng` Terraform workflows are introduced to manage
shared build infrastructure in `eng-4ks`.

## Review Guidance

- Do not grant `roles/owner` or `roles/editor` to any GitHub service account.
- `roles/resourcemanager.projectIamAdmin` is broad — it allows setting any
  binding on the project. Prefer it over `roles/owner` but treat it as
  sensitive. Review any Terraform resource that calls `google_project_iam_member`
  or `google_project_iam_binding` to ensure it is not escalating privileges
  beyond what `iac/app` already manages.
- The WIF provider condition restricts impersonation to `refs/heads/main`. A
  compromised PR branch cannot impersonate production service accounts.
- If a PR Terraform plan path is reintroduced, bind it to a read-only service
  account with no IAM write permissions.
