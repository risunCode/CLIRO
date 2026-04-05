# Implementation Plan - App Boundary Simplification

## Phase 1 - Typed DTO boundary cleanup

- [x] 1. Replace untyped CLI sync and model catalog responses with canonical Go DTOs
  - Add typed DTO structs for model catalog items, CLI sync files, CLI sync statuses, and CLI sync results near the Wails bridge or in a dedicated boundary types file
  - Update `GetLocalModelCatalog`, `GetCLISyncStatuses`, and `SyncCLIConfig` replacement paths to return typed values instead of `map[string]any`
  - Keep JSON field names aligned with existing frontend expectations where possible to reduce churn
  - _Requirements: 1.1, 1.2, 1.4_

- [x] 2. Regenerate Wails bindings for the migrated DTOs
  - Run `wails build` after the typed DTO changes
  - Verify `frontend/wailsjs/go/main/App.d.ts` and `frontend/wailsjs/go/models.ts` expose typed models for the migrated APIs
  - Remove or update frontend code paths that still expect ambiguous map-shaped payloads
  - _Requirements: 1.3, 1.5, 8.5_

- [x] 3. Simplify frontend compat mapping for typed router/sync payloads
  - Update `frontend/src/backend/compat/router-compat.ts` and any related compat helpers to map typed DTOs instead of reconstructing shape from unknown records
  - Delete redundant coercion paths for fields that become guaranteed by the new DTO contract
  - Run `cd frontend && npm run check` to verify the new contracts compile cleanly
  - _Requirements: 1.5, 6.2, 8.5_

## Phase 2 - Router and cloudflared command migration

- [x] 4. Introduce typed router patch and runtime action input models in Go
  - Add `UpdateProxySettingsInput`, `UpdateCloudflaredSettingsInput`, `ProxyRuntimeAction`, and `CloudflaredAction`
  - Define any supporting result model such as `ProxySettingsUpdateResult`
  - Add validation helpers for supported action strings and patch fields
  - _Requirements: 2.1, 2.4, 2.5, 3.4_

- [x] 5. Implement `UpdateProxySettings` with shared runtime-sensitive sequencing
  - Replace the public Wails methods `SetProxyPort`, `SetAllowLAN`, `SetAutoStartProxy`, `SetProxyAPIKey`, `RegenerateProxyAPIKey`, `SetAuthorizationMode`, and `SetSchedulingMode` with one typed patch entry point
  - Centralize the existing restart sequencing for port and LAN changes so proxy/cloudflared behavior remains unchanged
  - Return structured result data for restart info and generated API key when applicable
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 8.2_

- [x] 6. Implement `UpdateCloudflaredSettings` and `RunCloudflaredAction`
  - Replace `SetCloudflaredConfig`, `InstallCloudflared`, `StartCloudflared`, `StopCloudflared`, and `RefreshCloudflaredStatus` public method paths with typed config patch + runtime action APIs
  - Preserve persisted enabled state, dependency checks, log messages, and status refresh behavior
  - Add validation for unsupported cloudflared actions
  - _Requirements: 2.4, 3.2, 3.3, 3.4, 3.5, 8.1_

- [x] 7. Implement `RunProxyAction` and migrate tray/runtime call sites
  - Replace `StartProxy`, `StopProxy`, and public toggle-specific action paths with one proxy runtime command API
  - Update tray-triggered behavior and emitted runtime events to keep current semantics
  - Ensure tray sync still occurs for every runtime-changing path
  - _Requirements: 3.1, 3.3, 3.4, 3.5, 8.1, 8.2_

- [x] 8. Update frontend router gateway to the new router command surface
  - Refactor `frontend/src/backend/gateways/router-gateway.ts` to expose `updateProxySettings`, `updateCloudflaredSettings`, `runProxyAction`, `runCloudflaredAction`, `runCliSync`, `getCliSyncFile`, and `saveCliSyncFile`
  - Remove direct usage of migrated raw generated binding functions from the router gateway
  - Keep endpoint tester logic intact while adapting any model catalog lookup to the new typed DTOs
  - _Requirements: 2.1, 3.1, 3.2, 5.1, 6.1, 6.2, 6.3_

## Phase 3 - Account and quota command migration

- [x] 9. Introduce typed account and quota command inputs in Go
  - Add `RunAccountActionInput`, `RunQuotaActionInput`, and supported action enums/constants
  - Validate account-scoped actions require `accountId` and reject unsupported command names
  - Keep command names aligned with frontend string unions for predictable bindings
  - _Requirements: 4.1, 4.2, 4.3, 4.5_

- [x] 10. Implement `RunAccountAction` in `app.go`
  - Replace the public Wails methods `RefreshAccount`, `RefreshAccountWithQuota`, `ToggleAccount`, `DeleteAccount`, and `ClearCooldown` with one account action dispatcher
  - Preserve current enable/disable health-state handling from `ToggleAccount`
  - Preserve current delete, cooldown clear, and refresh behavior while improving error messages for invalid actions
  - _Requirements: 4.1, 4.3, 4.4, 4.5, 4.6, 8.1_

- [x] 11. Implement `RunQuotaAction` in `app.go`
  - Replace the public Wails methods `RefreshQuota`, `RefreshAllQuotas`, and `ForceRefreshAllQuotas` with one quota action dispatcher
  - Reuse the existing quota service methods internally without changing quota business logic
  - Add clear validation for missing `accountId` on `refresh-one`
  - _Requirements: 4.2, 4.3, 4.5, 4.6, 8.1_

- [x] 12. Simplify frontend accounts gateway and sync wrappers
  - Refactor `frontend/src/backend/gateways/accounts-gateway.ts` to expose `runAccountAction`, `runQuotaAction`, and one generic `syncAccountAuth(accountId, target)` helper
  - Remove wrapper methods that only hardcode sync target or account mutation mode
  - Update account sync result mapping to work from the unified sync result path already introduced in the auth refactor
  - _Requirements: 4.1, 4.2, 5.2, 5.5, 6.1, 6.2, 6.4_

## Phase 4 - CLI sync command unification

- [x] 13. Replace public CLI sync file methods with typed request objects
  - Migrate `GetCLISyncFileContent` and `SaveCLISyncFileContent` to typed request-input methods such as `GetCLISyncFile` and `SaveCLISyncFile`
  - Keep the existing path validation from `cliconfig.Service.resolveConfigFile`
  - Ensure unsupported target/path errors remain descriptive
  - _Requirements: 5.1, 5.4, 6.2_

- [x] 14. Replace `SyncCLIConfig` with typed `RunCLISync`
  - Accept a typed input object containing target and optional model
  - Preserve current API key auto-generation behavior and logging side effects
  - Return typed CLI sync result DTOs instead of map-shaped payloads
  - _Requirements: 1.1, 5.1, 5.3, 5.4, 8.3_

- [x] 15. Update frontend CLI sync call sites to the new command API
  - Update router feature utilities, stores, and components that call sync status/read/write/run operations
  - Keep existing target unions and UX flows intact while removing redundant wrapper indirection
  - Run `cd frontend && npm run check` after the migration
  - _Requirements: 5.1, 5.2, 5.5, 6.2, 6.4, 8.4, 8.5_

## Phase 5 - System action and controller consolidation

- [x] 16. Add typed system action dispatcher in the Wails bridge
  - Implement `RunSystemAction` with support for `confirm-quit`, `hide-to-tray`, `restore-window`, `open-data-dir`, and optionally `clear-logs` if logs remain in scope for the migration
  - Keep `OpenExternalURL` as a dedicated typed method
  - Preserve tray/runtime side effects and current no-op safety where applicable
  - _Requirements: 6.2, 7.1, 8.1_

- [x] 17. Update frontend system/logs gateways to use the system action model
  - Refactor `frontend/src/backend/gateways/system-gateway.ts` and any related logs gateway methods affected by `RunSystemAction`
  - Remove migrated raw binding imports from gateway modules
  - Keep `getState`, `getHostName`, and `openExternalURL` behavior intact
  - _Requirements: 6.1, 6.2, 6.3, 8.1_

- [x] 18. Consolidate repetitive handlers inside `app-controller.ts`
  - Replace per-action handler duplication for router settings, runtime actions, account actions, quota actions, and auth sync targets with typed command helpers
  - Keep the controller's public API domain-oriented even if several actions share the same internal helper
  - Preserve current toast, refresh, and busy-state behavior for each flow
  - _Requirements: 6.4, 7.1, 7.2, 7.3, 7.4, 8.1_

## Phase 6 - Cleanup and validation

- [x] 19. Remove obsolete migrated Wails methods and stale frontend references
  - Delete public backend methods replaced by the new command/patch APIs once all call sites are migrated
  - Remove stale generated-binding imports and unused compat helpers in the frontend
  - Search the repo to ensure migrated specialized method names are no longer referenced
  - _Requirements: 2.6, 3.5, 4.6, 5.5, 8.4_

- [x] 20. Run full regression validation for the migrated scope
  - Run `go test . ./internal/...`
  - Run `cd frontend && npm run check`
  - Run `wails build`
  - Verify key flows still work conceptually after migration: proxy runtime, proxy settings restart sequencing, cloudflared actions, account actions, quota actions, CLI sync, and system actions
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_
