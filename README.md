# Getting started

## Ingredients

- nodejs 18
- pnpm >= 6.9.0 (using 7.1.7)
- go >= 1.18
- Docker
- minikube or k3d (try minikube; fallback to k3d)
- tilt
- swaggo/swag/cmd/swag@v1.8.1 (bug in versions later than 1.8.1 as of writing this)

## Prep

1. Install and run `docker`
1. Install and run k8s: `kind` or `minikube` or `k3d`
1. Install `kubectl`
1. Install [tilt](https://docs.tilt.dev/install.html)
1. Install `nodejs`, `pnpm`, and `go`
1. Install swaggo/swag \
   `go install github.com/swaggo/swag/cmd/swag@v1.8.12`

## Instructions

1. Map `local.4ks.io` to 127.0.0.1 in your host file. Navigate to `https://local.4ks.io`
1. Run `docker compose up -d` to start the local nginx proxy.
1. Add `dev/apps/nginx/local.4ks.io.pem` to system or browser's trusted certs
1. Start kubernetes including a local container registry \
    `minikube start` or `minikube start --driver=docker` \
    or \
   `k3d cluster create 4ks --registry-create 4ks-registry`
1. Nginx is used for local development only. Build and run using docker-compose. \
   `docker-compose up -d --build`
1. Install npm deps \
   `pnpm install`
1. swaggo/swag is used to regenerate the swagger definition in apps/api/docs and the contents of libs/ts/api-fetch \
   `pnpm run swag`
1. Create the local kind cluster and registry: \
   `make up`
1. Start the Tilt-managed application resources on the existing cluster: \
   `make start`

## Local lifecycle

- `make up` creates the local Kubernetes infrastructure: the default kind
  cluster and the local registry container.
- `make down` is the destructive reset. It removes Tilt-managed app resources,
  terminates the running Tilt process, deletes the kind cluster, and removes the
  local registry container. Persistent volumes inside the kind cluster are lost.
- `make start` starts only the Tilt-managed application resources on an existing
  cluster. Use this for normal development after `make up`. If the local kind
  control-plane or registry container is not running, it exits with a `make up`
  instruction. If Tilt is already running on the default port, `make start`
  attaches by streaming `tilt logs -f` instead of trying to launch a second Tilt
  server. It prints the current Tilt resource status before attaching so failed
  pods are visible.
- `make stop` stops only the Tilt-managed application resources, then
  terminates the running Tilt process. It leaves the kind cluster, registry
  container, and persistent volumes intact.
- `make restart` runs `tilt down` followed by `tilt up --stream=true`. Use this
  when an existing Tilt session is reachable but resources are stale, stuck, or
  port-forwards keep reconnecting.

## Serving

1. Tilt syncs most code changes into the running containers on kubernmetes. libs/ts/api-fetch needs to be refreshed "manually". Simply run `pnpm swag` and tilt will work its magic.
