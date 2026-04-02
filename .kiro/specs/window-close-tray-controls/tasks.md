# Implementation Plan

- [x] 1. Add backend lifecycle guards and close-control entry points
  - Update `main.go` to disable `HideWindowOnClose` and wire `OnBeforeClose` to the app close guard.
  - Extend `App` lifecycle state in `app.go` with guarded quit flow and tray capability flags.
  - Implement backend methods for `beforeCloseGuard`, `ConfirmQuit`, `HideToTray`, `RestoreWindow`, and `ExitFromTray`.
  - _Requirements: 1.1, 1.3, 1.4, 2.2, 2.3, 5.3_

- [x] 2. Expose tray capability and close actions through app state and Wails bindings
  - Extend the backend state payload returned by `GetState()` with tray support and tray availability fields.
  - Add Wails-bound methods and frontend API adapters for `ConfirmQuit`, `HideToTray`, and `RestoreWindow`.
  - Keep backend as the only owner of quit authorization so frontend never calls raw quit directly.
  - _Requirements: 2.3, 3.6, 5.1, 5.4_

- [x] 3. Implement frontend close-confirm modal flow using existing reusable modal primitives
  - Add close-modal overlay state and controller actions in `frontend/src/app/services/app-controller.ts`.
  - Create `frontend/src/app/modals/AppCloseModal.svelte` using `BaseModal` and `ModalWindowHeader`.
  - Mount the modal in `frontend/src/app/providers/AppOverlayStack.svelte` and wire dismiss behavior so it cancels close cleanly.
  - Disable or annotate the `Minimize to Tray` action when tray is unavailable for the current session.
  - _Requirements: 1.2, 1.4, 2.1, 2.3, 2.4_

- [x] 4. Add frontend event handling for close interception, restore, and proxy refresh
  - Subscribe to backend events for `app:close-requested`, `app:window-restored`, and `app:proxy-state-changed`.
  - Open the close modal on close-request events and keep existing UI/session state intact when close is cancelled.
  - Trigger `GetState()` and `GetProxyStatus()` refresh paths after tray restore or tray-driven proxy changes.
  - _Requirements: 1.3, 5.1, 5.2, 5.4_

- [x] 5. Build a Windows-first tray controller package with safe fallback behavior
  - Add a new `internal/tray/` package with shared interface, Windows implementation, and non-Windows no-op implementation.
  - Embed a tray icon asset in the package so tray startup does not depend on runtime file paths.
  - Initialize tray startup from the backend app lifecycle, log failures, and keep the app usable if tray init fails.
  - _Requirements: 3.1, 3.6_

- [x] 6. Wire tray menu actions to backend window and proxy controls
  - Add tray menu items for `Open`, proxy enable/disable toggle, and `Exit App`.
  - Route `Open` to the shared restore-window path and `Exit App` to the explicit quit path that bypasses the close modal.
  - Route proxy toggle actions through the existing `StartProxy` and `StopProxy` backend flows rather than duplicating logic.
  - _Requirements: 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 5.3_

- [x] 7. Keep tray state synchronized with proxy runtime changes
  - Add tray state update helpers in `app.go` so tray labels reflect the latest proxy running state.
  - Refresh tray state after proxy changes triggered from either the tray menu or the main UI.
  - Emit backend events only after successful tray-driven proxy state changes and preserve prior state on failure.
  - _Requirements: 4.3, 4.4, 5.2_

- [x] 8. Add backend tests for close lifecycle and tray-state coordination
  - Create or extend Go tests covering close interception, one-shot quit authorization, explicit tray exit, and tray proxy toggle behavior.
  - Test fallback-safe behavior when tray support is unavailable or tray initialization fails.
  - Keep tests focused on deterministic app methods and helper functions rather than native UI runtime internals.
  - _Requirements: 1.1, 2.2, 3.6, 4.4, 5.3_

- [x] 9. Regenerate bindings and run repository validation for the new lifecycle flow
  - Regenerate Wails bindings as needed after adding new exported app methods.
  - Run `go test . ./internal/...` and fix any backend regressions introduced by the close/tray changes.
  - Run `cd frontend && npm run check` through the project workflow and fix any typing or UI integration issues.
  - _Requirements: 1.2, 2.1, 3.1, 4.1, 5.1_
