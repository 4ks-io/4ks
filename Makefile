
GO_MODULES := . apps/fetcher apps/media-upload utils/media-update utils/recipe-search-refresh utils/recipe-update
KIND_CLUSTER := kind
KIND_CONTROL_PLANE := $(KIND_CLUSTER)-control-plane
KIND_REGISTRY := kind-registry
TILT_PORT ?= 10350

.PHONY: up down start stop restart status logs swagger swag local tidy lint security check


start:
	@if ! docker container inspect -f '{{.State.Running}}' "$(KIND_CONTROL_PLANE)" 2>/dev/null | grep -qx true; then \
		printf "Kind control-plane container $(KIND_CONTROL_PLANE) is not running.\n"; \
		printf "Run 'make up' to create/start the local Kubernetes infrastructure, then run 'make start' again.\n"; \
		exit 1; \
	fi; \
	if ! docker container inspect -f '{{.State.Running}}' "$(KIND_REGISTRY)" 2>/dev/null | grep -qx true; then \
		printf "Local registry container $(KIND_REGISTRY) is not running.\n"; \
		printf "Run 'make up' to create/start the local Kubernetes infrastructure, then run 'make start' again.\n"; \
		exit 1; \
	fi; \
	if command -v curl >/dev/null 2>&1 && curl -fsS "http://127.0.0.1:$(TILT_PORT)/" >/dev/null 2>&1; then \
		printf "Tilt is already running on http://localhost:$(TILT_PORT)/; attaching to logs.\n"; \
		printf "Current Tilt resource status:\n"; \
		tilt get uiresources || true; \
		printf "If the session is stale or a pod is stuck, run 'make restart'.\n"; \
		tilt logs -f; \
	else \
		tilt up --stream=true; \
	fi

stop:
	@pid="$$(tilt get session Tiltfile -o jsonpath='{.status.pid}' 2>/dev/null || true)"; \
	tilt down || true; \
	if [ -n "$$pid" ]; then \
		printf "Terminating Tilt process %s.\n" "$$pid"; \
		kill "$$pid" 2>/dev/null || true; \
	elif command -v lsof >/dev/null 2>&1; then \
		pids="$$(lsof -tiTCP:$(TILT_PORT) -sTCP:LISTEN 2>/dev/null || true)"; \
		if [ -n "$$pids" ]; then \
			printf "Terminating Tilt process(es) listening on port %s: %s.\n" "$(TILT_PORT)" "$$pids"; \
			kill $$pids 2>/dev/null || true; \
		else \
			printf "No running Tilt process found after tilt down.\n"; \
		fi; \
	else \
		printf "No running Tilt process found after tilt down.\n"; \
	fi

restart: stop
	tilt up --stream=true

up:
	./scripts/kind-with-registry.sh

down: stop
	@if kind get clusters 2>/dev/null | grep -qx "$(KIND_CLUSTER)"; then \
		kind delete cluster --name $(KIND_CLUSTER); \
	else \
		printf "Kind cluster $(KIND_CLUSTER) does not exist; skipping delete.\n"; \
	fi
	@if docker container inspect "$(KIND_REGISTRY)" >/dev/null 2>&1; then \
		docker rm -f "$(KIND_REGISTRY)"; \
	else \
		printf "Docker container $(KIND_REGISTRY) does not exist; skipping remove.\n"; \
	fi

logs:
	tilt logs -f

status:
	tilt get uiresources

swagger:
	python3 -m webbrowser http://localhost:5174/swagger/index.html#/

swag:
	pnpm run swag

local:
	python3 -m webbrowser https://local.4ks.io/

tidy:
	@for module in $(GO_MODULES); do \
		if (cd "$$module" && go mod tidy); then \
			printf "\033[0;32mOK\033[0m tidy %s\n" "$$module"; \
		else \
			status=$$?; \
			printf "\033[0;31mFAIL\033[0m tidy %s\n" "$$module"; \
			exit $$status; \
		fi; \
	done
	@go work sync

lint:
	@for module in $(GO_MODULES); do \
		if (cd "$$module" && go tool revive -set_exit_status ./...); then \
			printf "\033[0;32mOK\033[0m lint %s\n" "$$module"; \
		else \
			status=$$?; \
			printf "\033[0;31mFAIL\033[0m lint %s\n" "$$module"; \
			exit $$status; \
		fi; \
	done

security:
	@for module in $(GO_MODULES); do \
		if (cd "$$module" && go tool govulncheck ./...); then \
			printf "\033[0;32mOK\033[0m security %s\n" "$$module"; \
		else \
			status=$$?; \
			printf "\033[0;31mFAIL\033[0m security %s\n" "$$module"; \
			exit $$status; \
		fi; \
	done

check: tidy lint security
