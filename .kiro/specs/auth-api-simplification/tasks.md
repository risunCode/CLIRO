# Implementation Plan — Auth API Simplification

## Phase 1 (Requirements 1–4)

---

- [x] 1. Definisikan unified types di root `auth` package
  - Tambah `AuthStart`, `AuthSessionView` sebagai canonical struct di `internal/auth/types.go` (bukan alias ke sub-package)
  - Tambah `AuthSyncTarget` type + konstanta `SyncTargetKilo`, `SyncTargetOpencode`, `SyncTargetCodexCLI`
  - Tambah `AuthSyncResult` unified struct dengan field: `Target`, `Path`, `BackedUp`, `BackupPath`, `Created`
  - Hapus type alias lama `CodexAuthStart`, `KiroAuthStart`, `CodexAuthSessionView`, `KiroAuthSessionView`
  - _Requirements: 1.1, 2.1, 3.1, 3.2_

- [x] 2. Buat `AuthProvider` interface dan compile-time assertions
  - [x] 2.1 Buat file `internal/auth/provider.go` dengan interface `AuthProvider`
    - Method: `StartAuth() (*AuthStart, error)`
    - Method: `StartSocialAuth(socialProvider string) (*AuthStart, error)`
    - Method: `GetSession(sessionID string) AuthSessionView`
    - Method: `CancelSession(sessionID string)`
    - Method: `SubmitCode(sessionID, code string) error`
    - Method: `RefreshAccount(account config.Account, force bool) (config.Account, error)`
    - Method: `Shutdown(ctx context.Context) error`
    - _Requirements: 2.1_

  - [x] 2.2 Tambah compile-time assertions di `provider.go`
    - `var _ AuthProvider = (*authcodex.Service)(nil)`
    - `var _ AuthProvider = (*authkiro.Service)(nil)`
    - Jalankan `go build ./internal/auth/...` — akan fail dulu, ini expected
    - _Requirements: 2.2, 2.3_

- [x] 3. Adapter `authcodex.Service` agar implement `AuthProvider`
  - Tambah `GetSession(sessionID string) auth.AuthSessionView` — wrap `GetAuthSession`, map ke root type
  - Tambah `CancelSession(sessionID string)` — delegate ke `CancelAuth`
  - Tambah `SubmitCode(sessionID, code string) error` — delegate ke `SubmitAuthCode`
  - Tambah `StartSocialAuth(_ string) (*auth.AuthStart, error)` — return error "social auth not supported for provider codex"
  - Update `StartAuth()` return type dari `*authcodex.AuthStart` → `*auth.AuthStart`
  - Jalankan `go build ./internal/auth/...` — compile-time assertion codex harus hijau
  - _Requirements: 2.2_

- [x] 4. Adapter `authkiro.Service` agar implement `AuthProvider`
  - Tambah `GetSession(sessionID string) auth.AuthSessionView` — wrap `GetAuthSession`, map ke root type
  - Tambah `CancelSession(sessionID string)` — delegate ke `CancelAuth`
  - Tambah `SubmitCode(sessionID, code string) error` — delegate ke `SubmitAuthCode`
  - Update `StartAuth()` dan `StartSocialAuth()` return type ke `*auth.AuthStart`
  - Jalankan `go build ./internal/auth/...` — compile-time assertion kiro harus hijau
  - _Requirements: 2.3_

- [x] 5. Refactor `auth.Manager` — ganti field ke `providers map[string]AuthProvider`
  - [x] 5.1 Update struct dan constructor `NewManager`
    - Ganti field `codex *authcodex.Service` + `kiro *authkiro.Service` → `providers map[string]AuthProvider`
    - Isi map di `NewManager`: `"codex": authcodex.NewService(...)`, `"kiro": authkiro.NewService(...)`
    - Tambah helper `providerFor(name string) (AuthProvider, error)`
    - _Requirements: 2.4, 2.5_

  - [x] 5.2 Tambah generic auth session methods
    - `StartAuth(provider string) (*AuthStart, error)`
    - `StartSocialAuth(provider, socialProvider string) (*AuthStart, error)`
    - `GetSession(provider, sessionID string) AuthSessionView`
    - `CancelSession(provider, sessionID string)`
    - `SubmitCode(provider, sessionID, code string) error`
    - Semua delegate ke `providerFor(provider)` lalu panggil method interface
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8_

  - [x] 5.3 Hapus semua provider-prefixed methods lama dari Manager
    - Hapus: `StartCodexAuth`, `GetCodexAuthSession`, `CancelCodexAuth`, `SubmitCodexAuthCode`
    - Hapus: `StartKiroAuth`, `StartKiroSocialAuth`, `GetKiroAuthSession`, `CancelKiroAuth`, `SubmitKiroAuthCode`
    - Update `Shutdown` agar iterasi `m.providers` map bukan codex+kiro literal
    - _Requirements: 1.9_

- [x] 6. Tambah `SyncAccountAuth` dan hapus 3 method sync lama di Manager
  - Tambah `SyncAccountAuth(accountID string, target AuthSyncTarget) (AuthSyncResult, error)`
  - Dispatch internal via `switch target` ke `syncauth.SyncCodexAccountToXxx` dan map ke `AuthSyncResult`
  - Hapus: `SyncCodexAccountToKiloAuth`, `SyncCodexAccountToOpencodeAuth`, `SyncCodexAccountToCodexCLI`
  - Jalankan `go build ./internal/auth/...` — harus bersih
  - _Requirements: 3.3, 3.4, 3.5, 3.6, 3.7_

- [x] 7. Fix `handleKiroProtocolURL` di `app.go` (internal deep-link handler)
  - `app.go:1196` memanggil `a.auth.GetAllKiroAuthSessions()` yang akan hilang
  - Tambah method internal `(m *Manager) allKiroSessions() []AuthSessionView` yang type-assert provider "kiro" ke `*authkiro.Service`
  - Atau: expose sebagai `GetAllSessions(provider string) []AuthSessionView` di Manager
  - Update call site di `app.go`
  - _Requirements: 1.9 (tidak boleh break existing functionality)_

- [x] 8. Refactor `app.go` Wails methods — auth session
  - Hapus: `StartCodexAuth`, `GetCodexAuthSession`, `CancelCodexAuth`, `SubmitCodexAuthCode`
  - Hapus: `StartKiroAuth`, `StartKiroSocialAuth`, `GetKiroAuthSession`, `CancelKiroAuth`, `SubmitKiroAuthCode`
  - Tambah: `StartAuth(provider string) (*auth.AuthStart, error)`
  - Tambah: `StartSocialAuth(provider, socialProvider string) (*auth.AuthStart, error)`
  - Tambah: `GetAuthSession(provider, sessionID string) auth.AuthSessionView`
  - Tambah: `CancelAuth(provider, sessionID string)`
  - Tambah: `SubmitAuthCode(provider, sessionID, code string) error`
  - Jalankan `go build .` — harus bersih
  - _Requirements: 4.1_

- [x] 9. Refactor `app.go` Wails methods — sync
  - Hapus: `SyncCodexAccountToKiloAuth`, `SyncCodexAccountToOpencodeAuth`, `SyncCodexAccountToCodexCLI`
  - Tambah: `SyncAccountAuth(accountID, target string) (auth.AuthSyncResult, error)`
  - Jalankan `go build .` — harus bersih
  - _Requirements: 4.2_

- [x] 10. Regenerate Wails JS/TS bindings
  - Jalankan `wails build` atau `wails dev` untuk regenerate `frontend/wailsjs/go/main/App.js` dan `App.d.ts`
  - Verify method lama tidak ada di generated file
  - Verify method baru (`StartAuth`, `GetAuthSession`, `SyncAccountAuth`, dll) muncul dengan signature yang benar
  - _Requirements: 4.4_

- [x] 11. Update frontend auth gateway
  - Edit `frontend/src/backend/gateways/auth-gateway.ts`
  - Ganti 9 method auth lama → 5 method baru dengan parameter `provider` string
  - Ganti 3 method sync lama → 1 method `syncAccountAuth(accountId, target)`
  - Update semua call site di `accounts-actions.ts`, `auth-api.ts`, dan connect modals
  - _Requirements: 4.3, 4.5_

- [x] 12. Update frontend connect flow untuk API baru
  - `KiroConnectModal.svelte` — update call dari `startKiroAuth()` / `startKiroSocialAuth(p)` → `startAuth("kiro")` / `startSocialAuth("kiro", p)`
  - `AccountSyncModal.svelte` — update 3 sync call terpisah → `syncAccountAuth(id, target)`
  - Poll `getAuthSession` — update signature baru `(provider, sessionId)`
  - `cancelAuth` — update signature baru `(provider, sessionId)`
  - _Requirements: 4.5_

- [x] 13. Jalankan full test suite dan type check
  - `go test ./internal/...` — 22 packages, 0 failures
  - `cd frontend && npm run check` — 0 type errors
  - Pastikan tidak ada referensi ke method lama (`StartCodexAuth`, `StartKiroAuth`, `SyncCodexAccountToKiloAuth`, dll) tersisa
  - _Requirements: 1.9, 3.7, 4.4_
