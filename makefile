FRONTEND_DIR = ./web/default
BACKEND_DIR = .
COMPOSE_FILE = docker-compose.yml
POSTGRES_SERVICE = postgres
BACKEND_SERVICE = new-api
POSTGRES_DB = new-api
POSTGRES_USER = root
IMAGE = new-api:local
VERSION = v1.0.0
DEV_SQLITE_PATH ?= one-api.db

.PHONY: build build-frontend build-image up down dev-web reset-setup

# 一键构建：先打包前端，再构建 docker 镜像
build: build-frontend build-image

build-frontend:
	@echo "==> Building frontend ($(VERSION))"
	@cd $(FRONTEND_DIR) && bun install && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(VERSION) bun run build

build-image:
	@echo "==> Building docker image $(IMAGE) ($(VERSION))"
	@docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) -f Dockerfile .

# 启动 / 停止整套服务（compose 会按 Dockerfile 重新构建）
up:
	@docker compose -f $(COMPOSE_FILE) up -d --build

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
