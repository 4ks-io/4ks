# API HTTP Security Change Set

This document describes the HTTP-facing behavior introduced by the current API
hardening change set. It is intended to be permanent operational documentation,
not a one-off implementation note.

## Scope

The current change set adds or changes:

- config-driven CORS allowlisting
- explicit trusted proxy configuration for client IP resolution
- application-level route rate limiting
- tests for CORS behavior and startup config validation
- deployment wiring for the new HTTP security environment variables

## Deployment Topology

The API is deployed as:

- external HTTPS load balancer
- serverless NEG
- Cloud Run service

Relevant Terraform is in:

- `iac/app/app-api.tf`
- `iac/app/modules/http-global-load-balancer/*`

The Cloud Run service currently uses:

- `min_instance_count = 0`
- `max_instance_count = 10`

That matters directly for the current rate-limiter behavior.

## CORS

The API now reads its browser allowlist from `CORS_ALLOWED_ORIGINS`.

Behavior:

- startup fails if `CORS_ALLOWED_ORIGINS` is empty
- startup fails if `CORS_ALLOWED_ORIGINS` contains `*`
- only exact origins are allowed
- `Access-Control-Allow-Credentials: true` is returned only for allowlisted origins
- preflight requests are answered by middleware without reaching handlers
- `Vary` headers are set for origin and preflight request metadata

The current deployment wiring sets:

- local/dev manifest: `https://local.4ks.io`
- Cloud Run service: `local.web_url`, plus `https://local.4ks.io` in the dev workspace

## Trusted Proxies

Gin only trusts forwarded IP headers when the immediate upstream peer is in
`SetTrustedProxies(...)`. The API now reads that allowlist from
`TRUSTED_PROXY_CIDRS`.

Behavior:

- startup fails if `TRUSTED_PROXY_CIDRS` is empty
- startup fails if any CIDR is invalid
- `c.ClientIP()` now depends on forwarded headers only when the request came from a trusted proxy CIDR

This setting affects:

- request logging
- per-IP rate limiting
- any future security decisions that rely on `c.ClientIP()`

## Rate Limiting

### Current implementation

The current application-level limiter is implemented in Go middleware using an
in-memory store backed by `golang.org/x/time/rate`.

Properties:

- process-local
- per-instance
- non-persistent
- no shared state across Cloud Run instances
- state is lost on container restart or scale-to-zero

### What this means on Cloud Run

Because the API runs on Cloud Run and can autoscale:

- each live instance maintains its own limiter buckets
- effective total capacity increases as more instances are added
- limits are not globally coordinated across instances
- when Cloud Run scales down to zero, all limiter history is forgotten
- the first request after cold start begins with fresh empty buckets

So the current limiter is best understood as:

- a lightweight per-instance abuse dampener
- useful for reducing burst pressure on expensive endpoints
- not sufficient as a strict globally enforced quota system

It is still useful, but it is not equivalent to edge-enforced or distributed
rate limiting.

### Current route policies

Public recipe reads, keyed by resolved client IP:

- `5 QPS`
- `120 QPM`
- `2000 QPH`

Authenticated writes, keyed by authenticated user ID with IP fallback:

- `2 QPS`
- `30 QPM`
- `600 QPH`

Recipe fetch by URL:

- `1 QPS`
- `3 QPM`
- `20 QPH`

User creation:

- `1 QPS`
- `3 QPM`
- `20 QPH`

Username availability checks:

- `2 QPS`
- `20 QPM`
- `240 QPH`

Media upload initialization:

- `1 QPS`
- `6 QPM`
- `60 QPH`

### Current limitations

The current application-level limiter does not provide:

- global cross-instance coordination
- durable counters across restarts
- edge enforcement before traffic reaches the container
- tenant-wide quotas across multiple pods/instances
- central visibility of counters outside process logs

### Recommended next step if stricter guarantees are needed

If the product needs true global enforcement, move the primary limit to one of:

- Cloud Armor or load-balancer level enforcement for IP-based public traffic
- a shared backing store such as Redis/Memorystore for distributed app-level quotas
- a dedicated quota service for per-user or per-tenant write budgets

The current in-memory middleware can still remain as a secondary in-process
backstop even after edge or distributed limits are added.

## Search Endpoint Note

The API service does not currently expose a dedicated public search endpoint.
Because of that, there is no in-service search route covered by this change.

If search is later added to this API service, it should get:

- its own named policy
- public per-IP limits
- separate monitoring after rollout

## Tests Included

The change set includes:

- CORS tests for allowed origin, denied origin, preflight, and credentials
- config validation tests for invalid CORS and proxy settings
- rate-limiter tests proving multi-window enforcement can reject on a longer budget such as QPM even when QPS is still available

Relevant files:

- `apps/api/middleware/cors_test.go`
- `apps/api/middleware/rate_limit_test.go`
- `apps/api/utils/config_test.go`
