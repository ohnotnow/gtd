# CLAUDE.md

## What this is

GTD — a single-binary terminal task manager in Go. Four source files, three test files. Pure Go SQLite (no CGo). Bubble Tea TUI.

## Files

- `main.go` — CLI entry point, arg parsing, `--print` mode
- `task.go` — Task struct, Priority (A-D) and Status (Todo/Done/InProgress) enums
- `store.go` — SQLite layer, all queries, carry-over logic
- `ui.go` — entire TUI: model, modes, key handling, views, table rendering

## Architecture

The TUI is a single Bubble Tea model with a mode enum (`modeTable`, `modeAdd`, `modeEdit`, `modeConfirmDelete`, `modeConfirmCarry`, `modeViewDate`, `modeFilter`). Forms use charmbracelet/huh. Table mode handles all action keys; form modes delegate to huh and return to table on completion or esc.

Tasks are looked up by cursor position via `visibleTasks()` which returns `filteredTasks` (if filter active) or `tasks`. Actions use task ID for store operations, so they work regardless of filtering.

## Key patterns

- `refreshTasks()` reloads from DB, re-applies filter, rebuilds table — call after any data mutation
- `rebuildTable()` reconstructs the bubbles table from `visibleTasks()`, preserving cursor
- Status is stored as an integer in `is_completed` column (0=todo, 1=done, 2=in-progress)
- Carried-over tasks are copies with `carried_from_id` pointing to the original
- Contexts partition tasks by a string key (default: "default")

## Build and test

```bash
go build -o gtd .
go test ./...
```

Tests use in-memory SQLite (`NewStoreWithPath(":memory:")`). No external dependencies needed.

## Conventions

- Date stored as `yyyy-mm-dd` internally, displayed/parsed as `dd/mm/yyyy` for user input
- Tasks ordered by `priority ASC, id ASC`
- The `#` column always shows the task's original 1-based index (even when filtered)
- Keep the app fast and simple — the user values single-keypress interactions
