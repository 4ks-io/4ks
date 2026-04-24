# colly docs

https://benjamincongdon.me/blog/2018/03/01/Scraping-the-Web-in-Golang-with-Colly-and-Goquery/

## goquery

https://github.com/PuerkitoBio/goquery

```
./cook -d=true -t=https://minimalistbaker.com/creamy-vegan-tofu-cauliflower-korma-curry/

make; ./cook -d=false -t=https://www.bbcgoodfood.com/recipes/best-ever-chocolate-brownies-recipe
```

## testing

https://dave.cheney.net/2019/05/07/prefer-table-driven-tests

```
go tool cover -func=c.out
```

# testing

https://ieftimov.com/posts/testing-in-go-go-test/

## fetcher auth

The fetcher callback to `/api/_fetcher/recipes` now uses HMAC-SHA256 instead of
encrypted timestamps.

- Signed fields: HTTP method, request host, request path, SHA-256 body hash,
  timestamp, and nonce.
- Replay protection: the API accepts only timestamps within a 2 minute skew
  window and stores seen nonces in memory for 5 minutes per instance.
- Secret storage: `API_FETCHER_PSK` should come from Secret Manager in deployed
  environments.
- Rotation: deploy the API and fetcher from the same secret version. Rotate by
  writing a new secret version, updating both services to consume it, then
  disabling the old version after rollout completes.

The nonce cache is process-local, so replay protection is enforced per API
instance. If this endpoint is scaled across multiple instances and stronger
cross-instance replay guarantees are required, move nonce storage to a shared
store with TTL semantics.
