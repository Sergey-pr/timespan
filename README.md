# TimeSpan

Minimal desktop task tracker with per-task timers. Click the clock icon on any running task to pop open a floating always-on-top timer window.

## Stack

| Layer | Tech |
|---|---|
| App framework | [Wails v3](https://v3.wails.io) (Go + native WebView) |
| Language | Go 1.23+ |
| UI | Vue 3 (Composition API) + Pinia + Vite |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGo) |
| Query builder | `goqu` v9 |

## Prerequisites

- **Go** 1.21+
- **Node.js** 18+ and npm
- **Wails v3 CLI** — install once:
  ```sh
  go install github.com/wailsapp/wails/v3/cmd/wails3@latest
  ```
- macOS: Xcode Command Line Tools (`xcode-select --install`)

## Development

```sh
# Install frontend dependencies (first time only)
cd frontend && npm install && cd ..

# Run in dev mode — hot-reloads Go and frontend on file changes
wails3 dev -config ./build/config.yml
```

Dev mode opens the main window. The Vite dev server runs on port 9245 by default. Go code is recompiled on save; Vue files hot-reload instantly via Vite HMR.

> **Note:** `wails3` must be on your PATH. If installed via `go install`, add `$GOPATH/bin` (or `~/go/bin`) to your shell's PATH:
> ```sh
> export PATH="$HOME/go/bin:$PATH"
> ```

## Build

```sh
# Production build — outputs a native binary to bin/
wails3 build

# macOS: package into a proper .app bundle (required to launch without a Terminal window)
wails3 task darwin:package
# → bin/timespan.app  (double-click or drag to /Applications)
```

## Regenerate JS bindings

After changing exported Go methods on `App`, re-run:

```sh
wails3 generate bindings ./...
```

Generated files land in `frontend/bindings/` and are committed to the repo.

## Project layout

```
.
├── app.go               # Bound service: task CRUD, timer window control, ticker
├── db.go                # SQLite DB wrapper (goqu)
├── main.go              # App + window setup (main window + floating timer window)
├── util.go              # generateID()
├── build/               # Taskfiles, icons, platform config
│   └── config.yml       # Product name, identifier, dev-mode config
├── frontend/
│   ├── bindings/        # Auto-generated JS bindings (committed)
│   │   └── timespan/    # app.js, models.js, index.js
│   ├── src/
│   │   ├── App.vue             # Main window shell
│   │   ├── TimerWindow.vue     # Floating timer OS window
│   │   ├── components/
│   │   │   └── TaskCard.vue    # Per-task card with action buttons
│   │   ├── stores/
│   │   │   └── taskStore.js    # Pinia store
│   │   └── assets/
│   │       ├── main.css        # Dark theme for main window
│   │       └── timer.css       # Styles for timer window
│   ├── index.html       # Main window entry
│   └── timer.html       # Timer window entry
└── go.mod
```

## Data

Tasks are stored in SQLite at:

| Platform | Path |
|---|---|
| macOS | `~/Library/Application Support/TimeSpan/timespan.db` |
| Linux | `~/.config/TimeSpan/timespan.db` |
| Windows | `%APPDATA%\TimeSpan\timespan.db` |

Any task left in `running` state when the app closes is reset to `paused` on next startup.
