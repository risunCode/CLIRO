# Requirements Document

## Introduction

`auth.Manager` dan `app.go` saat ini mengekspos terlalu banyak method yang provider-prefixed dan redundan — 9 method auth terpisah (Codex vs Kiro), 3 method sync yang hampir identik, dan pola dispatch yang sudah ada di `refreshAccount` tapi tidak diterapkan konsisten di seluruh Manager.

Tujuan refactor ini: kurangi API surface, hilangkan duplikasi, dan buat kontrak yang lebih bersih antara `app.go` ↔ `auth.Manager` ↔ sub-services — tanpa mengubah fungsionalitas yang sudah berjalan.

---

## Requirements

### Requirement 1 — Provider-generic auth session API

**User Story:** As a developer, I want a single set of auth session methods that work for any provider, so that adding a new provider doesn't require new method pairs on `auth.Manager` and `app.go`.

#### Acceptance Criteria

1. WHEN `StartAuth(provider)` is called with `"codex"` THEN the system SHALL delegate to `authcodex.Service.StartAuth()` and return `*AuthStart`.
2. WHEN `StartAuth(provider)` is called with `"kiro"` THEN the system SHALL delegate to `authkiro.Service.StartAuth()` and return `*AuthStart`.
3. WHEN `StartSocialAuth(provider, socialProvider)` is called with `"kiro"` THEN the system SHALL delegate to `authkiro.Service.StartSocialAuth(socialProvider)`.
4. WHEN `StartSocialAuth(provider, socialProvider)` is called with `"codex"` THEN the system SHALL return an error (`"social auth not supported for provider codex"`).
5. WHEN `GetSession(provider, sessionID)` is called THEN the system SHALL return the session view from the matching provider service.
6. WHEN `CancelSession(provider, sessionID)` is called THEN the system SHALL cancel the session on the matching provider service.
7. WHEN `SubmitCode(provider, sessionID, code)` is called THEN the system SHALL submit the code to the matching provider service.
8. WHEN an unknown provider string is passed to any of these methods THEN the system SHALL return a descriptive error.
9. The existing `provider-prefixed` methods (`StartCodexAuth`, `StartKiroAuth`, etc.) SHALL be removed from `auth.Manager` and `app.go`.

---

### Requirement 2 — `AuthProvider` interface

**User Story:** As a developer, I want a shared interface that both `authcodex.Service` and `authkiro.Service` implement, so that `auth.Manager` can dispatch generically without type-switching.

#### Acceptance Criteria

1. The system SHALL define an `AuthProvider` interface in `internal/auth/` with methods: `StartAuth`, `StartSocialAuth`, `GetSession`, `CancelSession`, `SubmitCode`, `RefreshAccount`, `Shutdown`.
2. `authcodex.Service` SHALL satisfy `AuthProvider` (with `StartSocialAuth` returning a not-supported error).
3. `authkiro.Service` SHALL satisfy `AuthProvider` fully.
4. `auth.Manager` SHALL store providers in a `map[string]AuthProvider` keyed by lowercase provider name, replacing the current `codex *authcodex.Service` and `kiro *authkiro.Service` fields.
5. WHEN a new provider is registered at construction time THEN `auth.Manager.providerFor(name)` SHALL return it without code changes elsewhere in Manager.

---

### Requirement 3 — Unified account auth sync

**User Story:** As a developer, I want a single `SyncAccountAuth(accountID, target)` method instead of three separate sync methods, so that adding a new sync target doesn't require a new method on Manager and app.go.

#### Acceptance Criteria

1. The system SHALL define an `AuthSyncTarget` type (string) with constants `SyncTargetKilo`, `SyncTargetOpencode`, `SyncTargetCodexCLI`.
2. The system SHALL define a unified `AuthSyncResult` struct with fields sufficient for all three current result types (path written, backed-up bool, backup path, accounts count or similar).
3. WHEN `SyncAccountAuth(accountID, SyncTargetKilo)` is called THEN the system SHALL call `syncauth.SyncCodexAccountToKiloAuth` and map the result to `AuthSyncResult`.
4. WHEN `SyncAccountAuth(accountID, SyncTargetOpencode)` is called THEN the system SHALL call `syncauth.SyncCodexAccountToOpencodeAuth` and map the result.
5. WHEN `SyncAccountAuth(accountID, SyncTargetCodexCLI)` is called THEN the system SHALL call `syncauth.SyncCodexAccountToCodexCLI` and map the result.
6. WHEN an unknown target is passed THEN the system SHALL return a descriptive error.
7. The existing `SyncCodexAccountToKiloAuth`, `SyncCodexAccountToOpencodeAuth`, `SyncCodexAccountToCodexCLI` methods SHALL be removed from `auth.Manager` and `app.go`.

---

### Requirement 4 — Simplified `app.go` Wails surface

**User Story:** As a developer, I want `app.go` to expose the minimum set of auth-related Wails methods, so that the frontend binding layer is smaller and easier to maintain.

#### Acceptance Criteria

1. `app.go` SHALL expose exactly these auth session methods: `StartAuth(provider)`, `StartSocialAuth(provider, social)`, `GetAuthSession(provider, sessionID)`, `CancelAuth(provider, sessionID)`, `SubmitAuthCode(provider, sessionID, code)`.
2. `app.go` SHALL expose exactly one sync method: `SyncAccountAuth(accountID, target)`.
3. All removed `app.go` methods SHALL have their frontend gateway counterparts updated in `frontend/src/backend/gateways/auth-gateway.ts`.
4. WHEN `wails build` is run after the refactor THEN the generated `frontend/wailsjs/go/main/App.js` and `App.d.ts` SHALL reflect the new signatures with no leftover old method names.
5. The frontend auth flow (connect modals, auth polling) SHALL continue to work correctly after the gateway update.

---

### Requirement 5 — `RefreshAccount` consolidation (optional / phase 2)

**User Story:** As a developer, I want a single `RefreshAccount` entry point with options instead of three separate refresh methods, so that the refresh surface is easier to reason about.

#### Acceptance Criteria

1. The system SHALL define a `RefreshOptions` struct with boolean fields `Force` and `IncludeQuota`.
2. `RefreshAccount(accountID string, opts RefreshOptions)` on `auth.Manager` SHALL replace `RefreshAccount` (force=true), `EnsureFreshAccount` (force=false), and the quota flag routes through the quota service when `IncludeQuota=true`.
3. `app.go` SHALL expose `RefreshAccount(accountID, force, includeQuota bool)` coercing to `RefreshOptions`.
4. IF `IncludeQuota` is true THEN the system SHALL trigger a quota refresh after token refresh completes.
5. The existing `RefreshAccount`, `RefreshAccountWithQuota`, `EnsureFreshAccount` method variants SHALL be removed or made private after migration.

> **Note:** Requirement 5 is scoped as phase 2. Phase 1 covers requirements 1–4 only.

---

## Out of Scope

- Changes to `authcodex.Service` or `authkiro.Service` internal logic.
- Changes to `internal/sync/authtoken/` implementation functions.
- Changes to `internal/provider/`, `internal/gateway/`, or `internal/route/`.
- Any UI/UX changes in the frontend beyond gateway method renames.
