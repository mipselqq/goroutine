# Developer Guidelines

This document is a quick onboarding guide and a single source of development rules for `goroutine`.

## 1) Goals and Principles

- Keep changes small and atomic.
- Prioritize readability and predictability over clever code.
- Follow Clean Architecture: dependencies flow inward (`handler -> service -> repository -> DB`).
- Treat security, observability, maintainability, and clean code as feature requirements.
- Ensure each change is testable.

## 2) Quick Start

Use the canonical Quick Start from README:

- [README Quick Start](README.md#quick-start)

## 3) Project Structure

Use the canonical project structure overview from README:

- [README Project Structure](README.md#project-structure)

## 4) Feature Development Workflow

### Step 1. Define the task

- Create an issue from template (`feature`, `bug`, `todo`).
- Do not ignore template structure.

### Step 2. Design the thinnest end-to-end path

For most features, touch layers in this order:

1. Domain: new value objects and invariants.
2. Service: use case logic and service-level errors.
3. Repository: data access and repository-level error mapping.
4. Handler: endpoint contract and HTTP status mapping.
5. Router (`internal/app/route.go`): route registration.
6. Swagger comments in handlers (spec generation is handled by hooks).

### Step 3. Implement per layer

- Do not bypass layers and do not leak DB details into handlers.
- Define interfaces at consumer side (`AuthService` in handler, `UserRepository` in service).
- Use explicit constructors (`New...`) for dependency injection.

### Step 4. Make sure non-glue code coverage is >80%

- Unit: domain, service, handler.
- Integration (`-tags=integration`): repository + database.
- E2E (`-tags=e2e`): happy path and critical user scenarios.

## 5) Code Rules

**Note:** Most formatting and style rules are enforced automatically by `git hooks` (Lefthook) and `golangci-lint`. Do not bypass them.

### 5.1 Domain

- Build every invariant-carrying entity via `New...` constructors.
- Avoid passing raw primitives across layers when a domain type exists.
- Keep validation errors short and explicit.

### 5.2 Service

- Orchestrate business logic without HTTP/JSON or SQL specifics.
- Return only service-level errors from service layer.
- Wrap with context (`fmt.Errorf("...: %w", errType)`) and map lower-layer errors to service errors.

### 5.3 Repository

- Keep SQL and infrastructure error translation in repository.
- Return only repository-level/domain-compatible errors from repository layer.
- Never leak raw driver errors outside repository.
- Pass `context.Context` through all DB operations.

### 5.4 Handler

- Handle only transport concerns: decode, validate, call service, encode response.
- Use `internal/handler/responders.go` for unified response format.
- Keep service-error to HTTP-status mapping centralized and explicit.
- Always return JSON with proper `Content-Type`.

### 5.5 Error boundaries and mapping

- Throw errors only from the corresponding layer.
- Map errors at layer boundaries:
  - DB/driver -> repository errors;
  - repository errors -> service errors;
  - service errors -> HTTP statuses/responses.

### 5.6 Observability (Logging, Metrics, Tracing)

- **Logging:** Use `slog` with module context (`logging.NewLoggerContext`). Never log secrets. Keep dev logs human-readable, prod logs structured (JSON).
- **Metrics:** Instrument critical paths with Prometheus metrics (RED method: Rate, Errors, Duration).
- **Tracing:** Propagate `context.Context` everywhere to ensure TraceIDs/SpanIDs are preserved across layers.

## 6) Testing Culture

- Tests are mandatory for any feature.
- Avoid giga PRs; large PRs reduce review quality and hide regressions.
- For code that is not glue code (wiring/bootstrapping), keep test coverage at 80%+.
- Use `t.Parallel()` by default where safe.
- Use scenario-oriented test names (`Success`, `Invalid JSON`, `User not found`).
- Use `internal/testutil` for integration/e2e setup (`SetupTestDB`, `TruncateTable`).

## 7) Migrations

- Apply schema changes only via files in `migrations/`.
- Validate migrations locally:
  - `make migrate-up`
  - `make migrate-status`
- Keep migrations idempotent-friendly and reversible where practical.

## 8) API Documentation

- Update Swagger annotations in handlers when endpoint contract changes.
- `make swag` runs on each commit via lefthook, so contract/doc drift should not happen.

## 9) Development Culture and Versioning

- **Branching:** Prefer trunk-based development with short-lived branches.
- **SemVer:** Follow Semantic Versioning.
- **Commits:** Use **Conventional Commits** (`feat`, `fix`, `chore`, `refactor`).
    - *Note:* Inside a specific PR/Branch, "work in progress" commits may be labeled as features based on technical context.
    - *However*, for the final merge/squash and Release Notes, strictly reserve the `feat` tag for changes **visible to the end-user**. Internal refactoring is `chore` or `refactor`.
- **PRs:** Must follow existing templates

## 10) Definition of Done

A task is done when:
- PR description and verification steps are clear.
- Acceptance criteria are satisfied.
- Required checks (lint, tests, swagger) pass in PR.
- Non-glue code coverage is 80%+.

## 11) Security & Configuration

- **No Hardcoded Configs:** Never hardcode ports, hosts, timeouts, or credentials. Use Environment Variables and `internal/config`.
- **Secrets:** Treat all tokens/passwords as `secrecy.SecretString`.
- **Dependencies:** Regularly check for vulnerabilities (`govulncheck`).

## 12) Common Anti-Patterns to Avoid

- Business logic inside handlers.
- Raw SQL/driver errors propagated to HTTP layer.
- Secret or token leaks in logs.
- Giga PRs with unclear scope.
- Skipping tests because change looks small.
