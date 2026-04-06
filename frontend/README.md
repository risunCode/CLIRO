# CLIRO Frontend

Frontend shell for the Wails desktop app. Built with Svelte + TypeScript + Vite, and wired to the Go backend through generated Wails bindings behind a clean `backend/` boundary layer.

## Stack

- Svelte 3 + TypeScript
- Vite 3
- Tailwind CSS 3
- Wails JS bindings (`frontend/wailsjs/go/main/*`)

## Architecture Summary

All backend access is centralized in `src/backend/` — feature UI and app shell never import from `wailsjs` directly.

- `src/backend/` — all backend access, split into client, models, and gateways
- `src/app/` — shell, bootstrap, routes/tabs, overlays, services, lib
- `src/features/` — per-feature UI, state, lib; each exposes a public `index.ts`
- `src/components/common/` — primitive reusable UI, domain-neutral
- `src/shared/` — cross-feature utilities, global stores, shared lib
- `src/styles/` — stylesheet entrypoint and design tokens
- `src/tabs/` — top-level tab wrappers for route composition

`src-old/` is a read-only reference of the previous structure and is **not imported** by active source.

## Key Folders

```text
src/
  App.svelte
  main.ts
  backend/
    client/       # wails-client, runtime-events, browser (low-level only)
    models/       # generated Wails type aliases used by the frontend boundary
    gateways/     # system-, logs-, accounts-, auth-, router-gateway
  app/
    bootstrap/
    modals/
    overlays/
    routes/
    services/
    shell/
    lib/
  features/
    accounts/     # index.ts public surface
    logs/         # index.ts public surface
    router/       # index.ts public surface
    settings/     # index.ts public surface
    usage/        # index.ts public surface
  components/
    common/       # Button, BaseModal, ToggleSwitch, StatusBadge, ...
  shared/
    lib/
    stores/
  styles/
    index.css
    tokens/
    base/
    primitives/
    features/
  tabs/
```

## Dependency Rules

- `app/*` → `features/<feature>/index`, `shared/*`, `backend/gateways/*`
- `features/<feature>/*` → `shared/*`, `backend/gateways/*`, internal feature only
- `backend/gateways/*` → `backend/client/*`, `backend/models/*`
- `components/common/*` → `shared/*` only
- **Forbidden**: any `src/*` → `src-old/*`, any `app/features/*` → `wailsjs/*` directly, any cross-feature internal imports

## Commands

Run commands from `frontend/` unless noted otherwise.

- Install dependencies: `npm install`
- Type check Svelte/TS: `npm run check`
- Build frontend bundle: `npm run build`

For full desktop app development and binding generation, run from repository root:

- `wails dev`
- `wails build`

## Notes

- All Wails calls go through `backend/gateways/*` — never import from `wailsjs` in feature or app code.
- Each feature exposes only its `index.ts` as a public surface; avoid deep imports into another feature.
- Keep route metadata centralized in `src/app/routes/app-routes.ts`.
- Keep feature-specific global styles in `src/styles/features/*.css`, shared primitives in `src/styles/primitives/components.css`.
