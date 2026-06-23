# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

New API is an AI API gateway/proxy written in Go. It aggregates upstream AI providers (OpenAI-compatible, Claude, Gemini, Azure, AWS Bedrock, etc.) behind unified APIs, with user management, billing, rate limiting, setup/admin flows, and a React dashboard.

本仓库为 QuantumNous/new-api 的本地 fork：**已彻底移除 classic 前端，只维护 `web/default/`**。所有 classic 主题相关的代码（`web/classic/` 下文件、`web/default/` 中的 `theme.frontend` 切换、`'classic'` 选项、Legacy Frontend 文案等）一律不引入；合并上游时遇到 classic 相关变更直接丢弃。

## Common Commands

### Backend / full project

- Run backend locally: `go run .`
- Run all Go tests: `go test ./...`
- Run tests in one package: `go test ./pkg/billingexpr`
- Run one Go test by name: `go test ./pkg/billingexpr -run TestName`
- Build backend binary: `go build -o new-api .`
- Build frontend and Docker image: `make build`
- Build frontend only: `make build-frontend`
- Build Docker image only: `make build-image`
- Start/stop Docker Compose stack: `make up` / `make down`
- Reset local setup wizard state: `make reset-setup`

`main.go` embeds `web/default/dist`; for a clean checkout, build the frontend before `go run`, `go build`, or tests that compile the main package.

### Frontend (`web/default/`)

Use Bun, not npm/yarn/pnpm, unless there is a specific reason.

- Install dependencies: `bun install`
- Dev server: `bun run dev` (or root `make dev-web`)
- Production build: `bun run build`
- Typecheck + production build: `bun run build:check`
- Typecheck only: `bun run typecheck`
- Lint: `bun run lint`
- Format check / write: `bun run format:check` / `bun run format`
- Sync i18n keys: `bun run i18n:sync`
- Unused code/dependency scan: `bun run knip`

After changing TypeScript or TSX, run `bun run typecheck` from `web/default/` and fix all type errors.

## Architecture

### Backend request flow

Startup is in `main.go`: it loads `.env`, initializes env/config, logger, HTTP clients, token encoders, databases, options, Redis, performance metrics, i18n, OAuth providers, and background tasks, then creates the Gin engine and calls `router.SetRouter`.

The backend follows `router -> controller -> service -> model`:

- `router/` wires HTTP routes and middleware groups for API, relay, dashboard, video, and web routes.
- `controller/` owns HTTP handlers and request/response orchestration.
- `service/` contains business logic, quota/billing helpers, HTTP client setup, token counting, async task handling, and provider-independent operations.
- `model/` owns GORM models, DB initialization/migrations, option storage, setup state, caches, and query helpers.
- `middleware/` handles auth, CORS, request logging, decompression/body cleanup, distribution, rate limiting, and route tagging.
- `setting/` loads runtime settings for model ratios, operations, system behavior, and performance.
- `common/` contains shared utilities for env, JSON, Redis, crypto, rate limits, logging, and DB mode flags.

### Relay / provider adapter flow

Relay routes are registered in `router/relay-router.go`. Requests pass through auth, performance checks, rate limiting, and `middleware.Distribute()` before reaching `controller.Relay`.

Provider selection is split across:

- `common.ChannelType2APIType` for channel/API type mapping.
- `relay.GetAdaptor` in `relay/relay_adaptor.go` for API type -> channel adaptor.
- `relay/channel/<provider>/` for provider-specific request conversion, upstream calls, response parsing, and streaming behavior.
- `relay/common.RelayInfo` for per-request metadata including token/user context, channel meta, final relay format, stream status, quota/billing state, and conversion history.

When adding or changing a provider channel, keep request DTO conversion, response usage accounting, streaming, and billing behavior aligned with existing channel adaptors.

### Frontend structure

The dashboard lives under `web/default/` and uses React 19, TypeScript, Rsbuild, TanStack Router/Query/Table/Virtual, Zustand, Base UI, Tailwind CSS, React Hook Form, Zod, and i18next.

Key frontend conventions are documented in `web/default/AGENTS.md`; follow it for component structure, i18n, route patterns, API usage, forms, state management, accessibility, and typechecking. High-level layout:

- `src/routes/` contains TanStack Router file routes using `createFileRoute`.
- `src/features/<feature>/` contains feature-scoped components, hooks, lib, API, constants, and types.
- `src/components/` contains shared UI/components.
- `src/lib/` contains shared frontend utilities and the common API client.
- `src/stores/` contains Zustand stores.
- `src/i18n/` contains i18next setup, locale files, and static keys.

## Internationalization

### Backend (`i18n/`)

- Library: `nicksnyder/go-i18n/v2`
- Languages: `en`, `zh`

### Frontend (`web/default/src/i18n/`)

- Library: `i18next` + `react-i18next` + `i18next-browser-languagedetector`
- Languages: `en`, `zh`, `fr`, `ru`, `ja`, `vi`
- Locale files: `web/default/src/i18n/locales/{lang}.json`; strings are keyed by English source text under the `translation` namespace.
- React components must use `const { t } = useTranslation()` so language changes re-render correctly.
- User-facing text must go through `t(...)`. Dynamic keys used from constants/config must be registered in `src/i18n/static-keys.ts` or otherwise be discoverable by the sync script.
- Run `bun run i18n:sync` from `web/default/` after adding or changing frontend translation keys.

## Project Rules

### JSON package

All JSON marshal/unmarshal operations in business code must use wrappers from `common/json.go`:

- `common.Marshal(v any) ([]byte, error)`
- `common.Unmarshal(data []byte, v any) error`
- `common.UnmarshalJsonStr(data string, v any) error`
- `common.DecodeJson(reader io.Reader, v any) error`
- `common.GetJsonType(data json.RawMessage) string`

Do not call `encoding/json` marshal/unmarshal functions directly outside the wrapper. `json.RawMessage`, `json.Number`, and other `encoding/json` types may still be referenced as types.

### Database compatibility

All DB code must support SQLite, MySQL >= 5.7.8, and PostgreSQL >= 9.6.

- Prefer GORM methods (`Create`, `Find`, `Where`, `Updates`, etc.) over raw SQL.
- Let GORM handle primary key generation; do not use `AUTO_INCREMENT` or `SERIAL` directly.
- When raw SQL is unavoidable, account for dialect differences:
  - PostgreSQL uses `"column"` quoting, while MySQL/SQLite use `` `column` ``.
  - Use `commonGroupCol`, `commonKeyCol` from `model/main.go` for reserved-word columns like `group` and `key`.
  - Use `commonTrueVal`/`commonFalseVal` for boolean values.
  - Use `common.UsingMainDatabase(...)` for primary database branches and `common.UsingLogDatabase(...)` for log database branches.
- Do not use database-specific features without cross-DB fallback, including MySQL-only functions, PostgreSQL-only operators, SQLite-unsupported `ALTER COLUMN`, or database-specific JSON column types without a `TEXT` fallback.
- Migrations must work on all three databases. For SQLite, use `ALTER TABLE ... ADD COLUMN` instead of `ALTER COLUMN` (see `model/main.go` for patterns).
- Avoid GORM boolean default tags such as `gorm:"default:true"` when the default is a business rule already enforced by code. MySQL and PostgreSQL can normalize boolean defaults differently, causing GORM `AutoMigrate` to repeatedly issue `ALTER TABLE` on restart. Prefer setting these defaults in request/model normalization, hooks, constructors, or service logic; do not replace `default:true` with `default:1` unless the behavior is verified across SQLite, MySQL, and PostgreSQL.

### Upstream request DTOs

For request structs parsed from client JSON and then re-marshaled to upstream providers, optional scalar fields must preserve explicit zero/false values:

- Use pointer types with `omitempty` for optional scalars, e.g. `*int`, `*uint`, `*float64`, `*bool`.
- Absent in client JSON => `nil` => omitted upstream.
- Explicit `0`, `0.0`, or `false` => non-nil pointer => sent upstream.
- Do not use non-pointer scalar fields with `omitempty` for optional upstream request parameters.

### New channel StreamOptions

When implementing a new channel, confirm whether the upstream provider supports stream options. If supported, add the channel to `streamSupportedChannels` in `relay/common/relay_info.go`.

### Billing expression system

Before changing tiered/dynamic billing or expression-based pricing, read `pkg/billingexpr/expr.md`. Follow its expression language, variable semantics, token normalization rules, quota conversion, snapshot/versioning, pre-consume, settlement, and log display design.

### Constant package boundaries

`constant/` is only for globally reusable constants and simple types. It must not contain business logic, database operations, third-party service calls, or imports of other project packages. Keep `constant/README.md` updated when adding new constant files or categories.
