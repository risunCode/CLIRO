# CLIro-Go Frontend

Frontend shell for the Wails desktop app. Built with Svelte + TypeScript + Vite, and wired to the Go backend through generated Wails bindings plus shared frontend gateway adapters.

## Stack

- Svelte 3 + TypeScript
- Vite 3
- Tailwind CSS 3
- Wails JS bindings (`frontend/wailsjs/go/main/*`)

## Architecture Summary

- App orchestration and shell live under `src/app/`.
- Feature modules live under `src/features/` and own their UI/state helpers.
- Cross-feature APIs, adapters, components, and utilities live under `src/shared/`.
- Top-level tab wrappers remain in `src/tabs/` for transitional route composition.
- Styles are organized in `src/styles/` with `src/styles/index.css` as the single stylesheet entrypoint.

## Key Folders

```text
src/
  App.svelte
  main.ts
  app/
    api/
    bootstrap/
    modals/
    overlays/
    routes/
    services/
    shell/
  features/
    accounts/
    logs/
    router/
    settings/
    usage/
  shared/
    api/
    components/
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

## Commands

Run commands from `frontend/` unless noted otherwise.

- Install dependencies: `npm install`
- Local frontend dev server: `npm run dev`
- Type check Svelte/TS: `npm run check`
- Build frontend bundle: `npm run build`
- Preview frontend build: `npm run preview`

For full desktop app development and binding generation, run from repository root:

- `wails dev`
- `wails build`

## Notes

- Prefer using `src/shared/api/wails/gateway.ts` over direct calls to generated bindings in feature UI.
- Keep route metadata centralized in `src/app/routes/app-routes.ts`.
- Keep feature-specific global styles in `src/styles/features/*.css`, and shared primitives in `src/styles/primitives/components.css`.
