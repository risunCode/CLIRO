# Requirements Document

## Introduction

Setelah refactor auth, surface API aplikasi masih punya banyak pola yang serupa tapi diekspos sebagai method terpisah di `app.go`, generated Wails bindings, gateway frontend, dan action layer frontend. Ini paling terlihat di domain router/proxy/cloudflared, account/quota actions, CLI sync, dan system/window actions.

Tujuan refactor ini adalah menyederhanakan boundary backend-frontend di seluruh app dengan pola yang sama seperti refactor auth: kurangi method spesifik-per-case, perkenalkan command/patch API yang lebih generic dan typed, hilangkan return type `map[string]any` di Wails surface, dan pertahankan behavior runtime yang ada tanpa perubahan UX atau perubahan logika bisnis besar.

---

## Requirements

### Requirement 1 - Typed Wails boundary objects

**User Story:** As a developer, I want exported Wails methods to use typed request/response structs instead of `map[string]any`, so that backend and frontend contracts are explicit and easier to maintain.

#### Acceptance Criteria

1. WHEN backend exports CLI sync, model catalog, proxy status, or other structured data to the frontend THEN the system SHALL return typed structs or slices of typed structs instead of `map[string]any` or `[]map[string]any`.
2. WHEN a Wails method returns structured data THEN the system SHALL define canonical DTO types in Go that match the generated frontend bindings.
3. WHEN `wails build` regenerates bindings THEN the generated `frontend/wailsjs/go/main/App.d.ts` and `frontend/wailsjs/go/models.ts` SHALL expose those typed objects with no leftover `Record<string, any>`-style ambiguity for the migrated APIs.
4. IF a response shape differs across targets in the same domain THEN the system SHALL model that difference explicitly with typed optional fields or discriminated variants instead of untyped maps.
5. The frontend compat layer SHALL be simplified to field mapping only, not schema reconstruction from untyped payloads.

---

### Requirement 2 - Unified proxy and router settings patch API

**User Story:** As a developer, I want proxy and router configuration updates to go through a small set of patch-based methods, so that adding a new setting does not require a new Wails method, gateway wrapper, and controller action.

#### Acceptance Criteria

1. WHEN the frontend updates proxy configuration such as port, LAN binding, auto-start, API key, authorization mode, or scheduling mode THEN the system SHALL support that update through a typed patch request object instead of separate setter methods for each field.
2. WHEN a patch request updates multiple proxy fields in one call THEN the system SHALL apply the changes atomically in a deterministic order.
3. IF a patch changes fields that currently require proxy or cloudflared restart sequencing THEN the system SHALL preserve the existing restart behavior.
4. WHEN a patch omits a field THEN the system SHALL leave the existing stored value unchanged.
5. IF a patch contains invalid values THEN the system SHALL return a descriptive validation error without applying partial invalid changes.
6. The existing discrete methods such as `SetProxyPort`, `SetAllowLAN`, `SetAutoStartProxy`, `SetProxyAPIKey`, `SetAuthorizationMode`, and `SetSchedulingMode` SHALL be removed from the public Wails surface after migration.

---

### Requirement 3 - Unified runtime action API for proxy and cloudflared

**User Story:** As a developer, I want runtime operations to use explicit action-based commands, so that start/stop/install flows are easier to extend and less repetitive across backend and frontend.

#### Acceptance Criteria

1. WHEN the frontend requests a proxy runtime operation THEN the system SHALL expose an action-based API that supports at least `start`, `stop`, and tray toggle behavior where applicable.
2. WHEN the frontend requests a cloudflared runtime operation THEN the system SHALL expose an action-based API that supports at least `install`, `start`, `stop`, and status refresh.
3. WHEN an action is invoked THEN the system SHALL preserve all existing side effects, including logging, tray state synchronization, persisted enabled state, and dependency checks.
4. IF an unsupported action string is passed THEN the system SHALL return a descriptive error.
5. The public Wails surface SHALL no longer require separate exported methods for each migrated runtime action.

---

### Requirement 4 - Unified account and quota command API

**User Story:** As a developer, I want account and quota mutations to be modeled as generic commands, so that new account operations can be added without expanding the surface area linearly.

#### Acceptance Criteria

1. WHEN the frontend performs an account mutation such as enable, disable, delete, clear cooldown, refresh account, or refresh account with quota THEN the system SHALL support that through a typed account action API.
2. WHEN the frontend performs a quota action such as refresh one account, refresh all, or force refresh all THEN the system SHALL support that through a typed quota action API.
3. IF an action requires an `accountID` THEN the system SHALL validate that the identifier is present before execution.
4. WHEN an action succeeds THEN the observable app state and persisted data SHALL match the behavior of the current specialized method.
5. IF an unknown or unsupported action is passed THEN the system SHALL return a descriptive error.
6. The existing public methods `RefreshAccount`, `RefreshAccountWithQuota`, `RefreshQuota`, `RefreshAllQuotas`, `ForceRefreshAllQuotas`, `ToggleAccount`, `DeleteAccount`, and `ClearCooldown` SHALL be removed from the migrated Wails surface.

---

### Requirement 5 - Unified sync command APIs

**User Story:** As a developer, I want both CLI config sync and account auth sync flows to follow a consistent command pattern, so that target-specific behavior is handled by typed targets instead of new exported methods.

#### Acceptance Criteria

1. WHEN the frontend requests CLI config sync work THEN the system SHALL expose typed target-based commands for status lookup, sync execution, file read, and file save.
2. WHEN the frontend requests account auth sync work THEN the system SHALL expose a typed target-based command that handles all supported targets through one API.
3. WHEN sync results differ by target THEN the system SHALL represent those differences through typed result variants or explicit optional fields.
4. IF a sync target or file path is unsupported THEN the system SHALL return a descriptive error.
5. The frontend SHALL no longer maintain separate wrapper methods whose only difference is a hardcoded sync target.

---

### Requirement 6 - Simplified frontend backend access layer

**User Story:** As a developer, I want the frontend backend access layer to be organized by domain with stable typed entrypoints, so that feature code no longer depends on dozens of raw generated binding exports.

#### Acceptance Criteria

1. WHEN frontend feature code calls backend functionality THEN it SHALL do so through domain-oriented gateway modules, not by importing raw generated bindings directly.
2. WHEN a backend domain is migrated to the new command model THEN the corresponding frontend gateway SHALL expose a small set of generic methods that mirror the typed backend contract.
3. The file `frontend/src/backend/client/wails-client.ts` SHALL stop acting as a long flat re-export list for migrated domains and instead support stable grouped access through the gateway layer.
4. The frontend controller and feature action layers SHALL replace target-specific wrapper methods with typed command helpers where behavior is shared.
5. Existing user-visible flows SHALL continue to work with no required UI redesign.

---

### Requirement 7 - App controller action consolidation

**User Story:** As a developer, I want the app controller to use reusable command helpers instead of repetitive per-action handlers, so that the state management layer is easier to understand and safer to extend.

#### Acceptance Criteria

1. WHEN multiple controller actions differ only by backend command name, toast text, or refresh behavior THEN the system SHALL consolidate them behind shared helper functions.
2. WHEN account auth sync or router actions are triggered from the UI THEN the controller SHALL use typed descriptors or command helpers instead of separate nearly identical functions per target.
3. IF a controller action has unique side effects beyond the shared helper behavior THEN the system SHALL keep only the minimal custom logic necessary for that case.
4. The resulting controller API exposed to Svelte components SHALL remain clear and domain-oriented even if the internal implementation becomes more generic.

---

### Requirement 8 - Backward-safe migration and behavioral parity

**User Story:** As a developer, I want the simplification refactor to preserve current behavior, so that the app becomes easier to maintain without regressions in proxy runtime, sync flows, or desktop lifecycle behavior.

#### Acceptance Criteria

1. WHEN the migrated APIs replace existing specialized methods THEN all existing flows for accounts, router, cloudflared, logs, settings, and window actions SHALL continue to function correctly.
2. WHEN proxy settings that previously caused controlled restart sequences are updated through the new API THEN the observable runtime behavior SHALL match the current implementation.
3. WHEN CLI sync and account auth sync are executed through the new API THEN the generated config files and auth files SHALL remain functionally equivalent to the current implementation.
4. WHEN the migration is complete THEN no frontend call sites for migrated domains SHALL reference removed specialized Wails method names.
5. The system SHALL pass `go test . ./internal/...`, `cd frontend && npm run check`, and `wails build` after the migrated scope is complete.

---

## Out of Scope

- Rewriting internal provider, gateway, routing, or protocol behavior.
- UI redesigns or visual changes.
- Changing persisted config formats unless required to preserve existing behavior.
- Replacing Wails itself or changing the desktop shell architecture.
- Large business-logic rewrites inside `internal/sync/`, `internal/cloudflared/`, or provider auth implementations beyond interface adaptation.
