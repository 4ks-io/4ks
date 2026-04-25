# GCS Buckets and IAM

This document records which buckets are intentionally public, which remain
private, and which service identities are expected to access them.

## Public Buckets

`static.<env>.4ks.io`

- Purpose: site-owned static assets and fallback media served to browsers.
- Public access: `roles/storage.objectViewer` for `allUsers`.
- Notes: this bucket is intentionally public and should stay documented as such
  in Terraform reviews.

`media-read.<env>.4ks.io`

- Purpose: processed recipe media variants served publicly after validation.
- Public access: `roles/storage.objectViewer` for `allUsers`.
- Notes: this bucket is intentionally public because processed media is served
  directly to browsers.

## Private Buckets

`media-write.<env>.4ks.io`

- Purpose: temporary browser uploads before the media-upload function validates,
  transforms, and removes the source object.
- Public access: none.
- Notes: this bucket should remain private even though browsers upload to it
  via signed URLs.

`<org>-<stage>-media-upload-deploy`

- Purpose: source bundle bucket for Cloud Function deployment artifacts.
- Public access: none.
- Notes: operational bucket, not a user-facing media path.

## Service Identities

Cloud Run API service account

- Needs enough access to mint signed upload URLs and publish fetcher jobs.
- Should trend toward least-privilege grants instead of broad project-wide
  editor roles.

media-upload service account

- Needs access to read from the upload bucket and write processed variants to
  the distribution bucket.
- Should use the narrowest bucket-scoped permissions that still support the
  upload pipeline.

## Review Guidance

- Treat any new `allUsers` grant as a deliberate public exposure that requires
  documentation.
- Prefer bucket-scoped IAM over project-wide storage roles when practical.
- Treat upload-path changes separately from read-only bucket changes so rollout
  and rollback stay simple.
