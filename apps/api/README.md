# readme

Dockerfile.dev for local dev
Dockerfile.ci for github workflow
Dockerfile for production build

## HTTP Security

Permanent documentation for this change set lives in [docs/api-http-security.md](/code/4ks-io/4ks/docs/api-http-security.md).

Production topology is `external HTTPS load balancer -> serverless NEG -> Cloud Run api`.
Set `CORS_ALLOWED_ORIGINS` to explicit frontend origins only. Wildcards are rejected at startup because the API allows credentialed browser requests.
Set `TRUSTED_PROXY_CIDRS` to the proxy layer that is expected to inject `X-Forwarded-For`; do not leave this implicit.

Current application-level rate limits:

- Public recipe read routes: 5 QPS, 120 QPM, 2000 QPH per IP.
- Authenticated write routes: 2 QPS, 30 QPM, 600 QPH per user.
- Recipe fetch by URL: 1 QPS, 3 QPM, 20 QPH per user.
- User creation: 1 QPS, 3 QPM, 20 QPH per user.
- Username availability checks: 2 QPS, 20 QPM, 240 QPH per user/IP fallback.
- Media upload initialization: 1 QPS, 6 QPM, 60 QPH per user.

The API service does not currently expose a dedicated search endpoint; search traffic is handled outside this service. If one is added here later, it should get its own public per-IP limit instead of reusing the generic read bucket.

The current application-level rate limiter is in-memory and process-local. On
Cloud Run, that means limits are enforced per live instance, reset on restart,
and reset again when the service scales down to zero.
