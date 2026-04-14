# TimeSpan Frontend

Vue 3 + Pinia + Vite frontend for the TimeSpan desktop app, embedded in the Wails v3 binary.

## Structure

```
frontend/
├── index.html          # Main window entry point
├── timer.html          # Floating timer window entry point
├── bindings/           # Auto-generated JS bindings
│   └── timespan/       # app.js one export per Go method on App
├── src/
│   ├── App.vue             # Main window root component
│   ├── TimerWindow.vue     # Floating timer window root component
│   ├── timer.js            # Mounts TimerWindow.vue into timer.html
│   ├── main.js             # Mounts App.vue into index.html
│   ├── components/
│   │   └── TaskCard.vue    # Task card with inline edit and action buttons
│   ├── stores/
│   │   └── taskStore.js    # Pinia store task state, live elapsed, event subscriptions
│   └── assets/
│       ├── main.css        # Dark theme for the main window
│       └── timer.css       # Styles for the floating timer window
└── vite.config.js      # Multi-page build (index.html + timer.html), wails plugin
```

## Two entry points

Both windows are built as separate pages in the same Vite bundle:
- `index.html` -> `main.js` -> `App.vue` the main 400×600 task list window
- `timer.html` -> `timer.js` -> `TimerWindow.vue` the 220×100 frameless always-on-top timer

## Go and Frontend communication

- **Method calls**: via `frontend/bindings/timespan/app.js` (imported in store and components). Regenerate after changing Go methods: `wails3 generate bindings ./...` from the repo root.
- **Events** (Go and frontend, both windows): `tick` (500ms), `task:updated` (after any mutation), `timer:open` (open timer for a task id).
- **Runtime**: `@wailsio/runtime` npm package `Events.On()` for subscriptions; resolved at build time by the Wails Vite plugin.

## Live elapsed time

`taskStore.js` tracks elapsed time locally without polling Go:
- Each task stores `_baseElapsed` (committed ms) and `_segStart` (JS timestamp of segment start)
- The `tick` event fires every 500ms from Go; the store updates `elapsedMs = _baseElapsed + (now - _segStart)`
- `TimerWindow.vue` does the same independently with its own `now` ref

## Development

Run from the repo root do not run Vite standalone during development:

```sh
wails3 dev -config ./build/config.yml
```

For frontend-only work (no Go calls, mock data):

```sh
npm install   # first time
npm run dev   # Vite dev server on port 9245
```
