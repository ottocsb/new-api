FRONTEND_DIR = ./web/default
BACKEND_DIR = .
COMPOSE_FILE = docker-compose.yml
POSTGRES_SERVICE = postgres
BACKEND_SERVICE = new-api
POSTGRES_DB = new-api
POSTGRES_USER = root
IMAGE = new-api:local
VERSION_FILE = VERSION
VERSION ?= $(shell if [ -f $(VERSION_FILE) ]; then cat $(VERSION_FILE); else echo v1.0.2; fi)
DEV_SQLITE_PATH ?= one-api.db
RELEASE_TEST_CONTAINER ?= new-api-build-check

# Docker Hub 推送配置：使用前先 `docker login`
# 可通过命令行覆盖，例如：make push DOCKER_USER=youruser VERSION=v1.0.1
DOCKER_USER ?=
DOCKER_IMAGE_NAME ?= new-api
REMOTE_IMAGE = $(DOCKER_USER)/$(DOCKER_IMAGE_NAME)

.PHONY: build bump-version build-frontend build-image test-image release up down dev-web reset-setup push

# 一键构建：先打包前端，再构建 docker 镜像
build: bump-version
	@$(MAKE) build-frontend build-image VERSION="$$(cat $(VERSION_FILE))"

bump-version:
	@current="$$(if [ -f "$(VERSION_FILE)" ]; then cat "$(VERSION_FILE)"; else echo "$(VERSION)"; fi)"; \
	next="$$(printf '%s\n' "$$current" | awk 'BEGIN{FS=OFS="."} /^v?[0-9]+(\.[0-9]+)*$$/ { prefix=""; if (substr($$1,1,1)=="v") { prefix="v"; $$1=substr($$1,2) } $$NF=$$NF+1; print prefix $$0; ok=1 } END{ if (!ok) exit 1 }')" || { \
		echo "ERROR: 无法递增版本号：$$current"; \
		exit 1; \
	}; \
	printf '%s\n' "$$next" > "$(VERSION_FILE)"; \
	echo "==> Version bumped: $$current -> $$next"

build-frontend:
	@echo "==> Building frontend ($(VERSION))"
	@if command -v bun >/dev/null 2>&1; then \
		cd $(FRONTEND_DIR) && bun install && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(VERSION) bun run build; \
	elif [ -d "$(FRONTEND_DIR)/node_modules" ]; then \
		cd $(FRONTEND_DIR) && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(VERSION) npm run build; \
	else \
		cd $(FRONTEND_DIR) && npm install --no-package-lock && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(VERSION) npm run build; \
	fi

build-image:
	@echo "==> Building docker image $(IMAGE) ($(VERSION))"
	@docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) -f Dockerfile .

test-image:
	@echo "==> Testing docker image $(IMAGE) ($(VERSION))"
	@container="$(RELEASE_TEST_CONTAINER)-$$(printf '%s' "$(VERSION)" | tr '.' '-')"; \
	docker rm -f "$$container" >/dev/null 2>&1 || true; \
	docker run -d --name "$$container" -e SQLITE_PATH=/data/test.db "$(IMAGE)" --log-dir /tmp/logs >/dev/null; \
	trap 'docker rm -f "$$container" >/dev/null 2>&1 || true' EXIT; \
	for i in $$(seq 1 30); do \
		if docker exec "$$container" wget -q -O - http://localhost:3000/api/status | grep -q '"success"[[:space:]]*:[[:space:]]*true'; then \
			echo "==> Image runtime check passed"; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "ERROR: image runtime check failed"; \
	docker logs "$$container"; \
	exit 1

release: build
	@$(MAKE) test-image push VERSION="$$(cat $(VERSION_FILE))"

# 推送到 Docker Hub：先 `docker login`，然后 `make push DOCKER_USER=youruser`
push:
	@if [ -z "$(DOCKER_USER)" ]; then \
		echo "ERROR: DOCKER_USER 未设置，请使用 'make push DOCKER_USER=youruser'"; \
		exit 1; \
	fi
	@echo "==> Tagging $(IMAGE) -> $(REMOTE_IMAGE):$(VERSION) and :latest"
	@docker tag $(IMAGE) $(REMOTE_IMAGE):$(VERSION)
	@docker tag $(IMAGE) $(REMOTE_IMAGE):latest
	@echo "==> Pushing $(REMOTE_IMAGE):$(VERSION)"
	@docker push $(REMOTE_IMAGE):$(VERSION)
	@echo "==> Pushing $(REMOTE_IMAGE):latest"
	@docker push $(REMOTE_IMAGE):latest

# 启动 / 停止整套服务（直接复用 make build 产出的镜像，不再重新构建）
up:
	@docker compose -f $(COMPOSE_FILE) up -d

down:
	@docker compose -f $(COMPOSE_FILE) down

# 仅启动前端开发服务器（后端需另行启动）
dev-web:
	@echo "Starting frontend dev server..."
	@cd $(FRONTEND_DIR) && bun install && bun run dev

reset-setup:
	@echo "Resetting local setup wizard state..."
	@if docker compose -f $(COMPOSE_FILE) ps --services --status running | grep -qx "$(POSTGRES_SERVICE)"; then \
		echo "Detected running docker PostgreSQL. Removing setup record and root users..."; \
		docker compose -f $(COMPOSE_FILE) exec -T $(POSTGRES_SERVICE) \
			psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) \
			-c 'DELETE FROM setups;' \
			-c 'DELETE FROM users WHERE role = 100;' \
			-c "DELETE FROM options WHERE key IN ('SelfUseModeEnabled', 'DemoSiteEnabled');"; \
		echo "Restarting docker backend so setup status is recalculated..."; \
		docker compose -f $(COMPOSE_FILE) restart $(BACKEND_SERVICE); \
	elif db_path="$${SQLITE_PATH:-$(DEV_SQLITE_PATH)}"; db_path="$${db_path%%\?*}"; [ -f "$$db_path" ]; then \
		db_path="$${SQLITE_PATH:-$(DEV_SQLITE_PATH)}"; \
		db_path="$${db_path%%\?*}"; \
		echo "Detected local SQLite database: $$db_path"; \
		sqlite3 "$$db_path" \
			"DELETE FROM setups; DELETE FROM users WHERE role = 100; DELETE FROM options WHERE key IN ('SelfUseModeEnabled', 'DemoSiteEnabled');"; \
		echo "SQLite setup state reset. Restart the local backend process before testing the setup wizard."; \
	else \
		echo "No running docker PostgreSQL or local SQLite database found."; \
		echo "Start the stack with 'make up', or set SQLITE_PATH/DEV_SQLITE_PATH to your local SQLite database."; \
		exit 1; \
	fi
