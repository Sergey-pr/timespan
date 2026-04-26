# TimeSpan

Minimal desktop task tracker with per-task timers. Start a task, track time, pause and resume freely.
Click the clock icon on any running task to pop open a floating always-on-top timer window.

## Features

- Create tasks with an optional description
- Start, pause, resume, and finish tasks
- Edit task title and description at any time
- Continue a finished task (resumes elapsed time)
- Floating always-on-top timer window per task
- Pause/resume synced between the main window and the timer window
- Persistent SQLite storage survives restarts

## Stack

| Layer         | Tech                                                  |
|---------------|-------------------------------------------------------|
| App framework | [Wails v3](https://v3.wails.io) (Go + native WebView) |
| Language      | Go 1.23+                                              |
| UI            | Vue 3 (Composition API) + Pinia + Vite                |
| Database      | SQLite via `github.com/doug-martin/goqu`              |

## Prerequisites

- **Go** 1.23+
- **Node.js** 18+ and npm
- **Wails v3 CLI** install once:
  ```sh
  go install github.com/wailsapp/wails/v3/cmd/wails3@latest
  ```

> `wails3` must be on your PATH. If installed via `go install`, add `~/go/bin` to your shell's PATH:
> ```sh
> export PATH="$HOME/go/bin:$PATH"
> ```

## Development

```sh
# Install frontend dependencies (first time only)
cd frontend && npm install && cd ..

# Run in dev mode with hot-reloads for Go and frontend on file changes
wails3 dev -config ./build/config.yml
```

Go code recompiles on save; Vue files hot-reload via Vite HMR. The Vite dev server runs on port 9245.

## Build

```sh
# Binary path: bin/timespan
wails3 build
```

## Regenerate JS bindings

After adding or changing exported methods on `App` in `app.go`:

```sh
wails3 generate bindings ./...
```

Generated files live in `frontend/bindings/` and are committed to the repo.

## Project layout

```
.
├── app.go               # App service: task CRUD, timer + error window control, tick emitter
├── db.go                # SQLite wrapper (goqu, modernc/sqlite)
├── main.go              # Three-window setup: main (400×600) + floating timer (220×100) + error (420×160)
├── utils.go             # generateID()
├── build/
│   └──  config.yml       # Product name, identifier, dev-mode config
├── frontend/
│   ├── bindings/        # Auto-generated JS bindings
│   ├── src/
│   │   ├── App.vue             # Main window
│   │   ├── TimerWindow.vue     # Floating timer OS window
│   │   ├── ErrorWindow.vue     # Error dialog window
│   │   ├── main.js             # Entry point for index.html
│   │   ├── timer.js            # Entry point for timer.html
│   │   ├── error.js            # Entry point for error.html
│   │   ├── components/
│   │   │   └── TaskCard.vue    # Task card: inline edit, status buttons
│   │   ├── stores/
│   │   │   └── taskStore.js    # Pinia store; live elapsed via segment tracking
│   │   └── assets/
│   │       ├── main.css        # Dark theme, main window
│   │       └── timer.css       # Timer window styles
│   ├── index.html       # Main window entry
│   ├── timer.html       # Timer window entry
│   └── error.html       # Error window entry
└── go.mod
```

## Data

Tasks are stored in SQLite at:

| Platform | Path                                                 |
|----------|------------------------------------------------------|
| macOS    | `~/Library/Application Support/TimeSpan/timespan.db` |
| Linux    | `~/.config/TimeSpan/timespan.db`                     |
| Windows  | `%APPDATA%\TimeSpan\timespan.db`                     |

Any task in `running` state when the app closes is reset to `paused` on next startup.

## License

MIT see [LICENSE](LICENSE).
