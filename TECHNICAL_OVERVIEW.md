# Technical Overview

Last updated: 2026-03-24

## What This Is

A fast, interactive terminal tool for daily task management inspired by *Time Management for System Administrators* by Thomas Limoncelli. Single binary, zero dependencies, works anywhere.

## Stack

- Go 1.24 / no framework
- TUI: charmbracelet suite (bubbletea, bubbles, huh, lipgloss)
- Database: modernc.org/sqlite (pure Go, no CGo)

## Directory Structure

```
.
├── main.go          CLI entry point, arg parsing, print mode
├── task.go          Domain model: Task, Priority, Status enums
├── store.go         SQLite persistence layer
├── ui.go            Bubble Tea TUI (model, update, view, all modes)
├── main_test.go     CLI arg parsing + print mode tests
├── task_test.go     Domain model unit tests
└── store_test.go    Database layer tests (in-memory SQLite)
```

Four source files, three test files. That's the whole app.

## Domain Model

```
Task
├── ID              int64       (auto-increment PK)
├── Date            string      (yyyy-mm-dd)
├── Description     string
├── Priority        A|B|C|D
├── TimeEstimate    string      (free text: "30m", "2h", "1d")
├── Status          Todo(0) | Done(1) | InProgress(2)
└── CarriedFromID   *int64      (self-referencing FK for carry-over lineage)
```

### Priority Levels

| Code | Label | Colour |
|------|-------|--------|
| A | Must do | Red `#ef4444` |
| B | Should do | Orange `#f97316` |
| C | Nice to do | Sky `#0ea5e9` |
| D | Delegate/defer | Zinc `#a1a1aa` |

### Status Values

Stored as integer in `is_completed` column (naming is historical):

| Value | Meaning | Symbol |
|-------|---------|--------|
| 0 | Todo | (empty) |
| 1 | Done | checkmark |
| 2 | In progress | play |

## Database

Single table, SQLite, stored at platform config dir (`~/Library/Application Support/sysadmin-gtd/tasks.db` on macOS).

```sql
CREATE TABLE tasks (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    date            TEXT NOT NULL,
    description     TEXT NOT NULL,
    priority        TEXT NOT NULL DEFAULT 'B',
    time_estimate   TEXT NOT NULL DEFAULT '',
    is_completed    INTEGER NOT NULL DEFAULT 0,
    carried_from_id INTEGER REFERENCES tasks(id),
    context         TEXT NOT NULL DEFAULT 'default'
);
```

Tasks are ordered by `priority ASC, id ASC` when queried.

## CLI Interface

```
gtd [dd/mm/yyyy] [--print] [--context <name>]
```

- No args: today's tasks, interactive TUI
- `--print`: non-interactive tabular output to stdout
- `--context`: partition tasks into named lists (default: "default")
- All flags are order-independent

## TUI Architecture

Single Bubble Tea model (`ui.go`) with a mode-based state machine:

```
modeTable ──┬── a ──→ modeAdd
            ├── e/enter ──→ modeEdit
            ├── x ──→ modeConfirmDelete
            ├── c ──→ modeConfirmCarry
            ├── v ──→ modeViewDate
            └── / ──→ modeFilter

All form modes ── esc ──→ modeTable
All form modes ── complete ──→ modeTable + refreshTasks()
```

Forms use charmbracelet/huh. The table uses charmbracelet/bubbles/table.

### Key Bindings (table mode)

| Key | Action |
|-----|--------|
| `a` | Add task |
| `s` | Toggle in-progress |
| `d` | Toggle done |
| `e`/`enter` | Edit task |
| `x` | Delete (with confirm) |
| `c` | Carry incomplete to tomorrow |
| `i` | Import from most recent day |
| `v` | View different date |
| `/` | Search/filter by name |
| `1`-`9` | Jump to task by number |
| `q` | Quit |

### Filter

Case-insensitive substring match on task description. Maintains a `filteredTasks` slice separate from `tasks`. All actions work on the visible (filtered) set via task ID. Original row numbers are preserved in the `#` column.

## Key Store Operations

| Method | Purpose |
|--------|---------|
| `GetTasksForDate` | Load tasks for a date+context |
| `AddTask` / `UpdateTask` / `DeleteTask` | CRUD |
| `MarkComplete` / `MarkIncomplete` / `MarkInProgress` | Status transitions |
| `GetCarryOverCandidates` | Incomplete tasks not already carried to target date |
| `CarryOverTasks` | Copy tasks to tomorrow with `carried_from_id` link |
| `CopyIncompleteTasks` | Import tasks without carry lineage |
| `GetLatestDateWithIncompleteTasks` | Find most recent date for import prompt |

## Testing

- Framework: Go standard `testing` package
- Pattern: table-driven tests, in-memory SQLite via `NewStoreWithPath(":memory:")`
- Run: `go test ./...`
- Coverage: CLI parsing, print output, all store CRUD/carry operations, domain model enums

## Local Development

```bash
go build -o gtd .    # build
go test ./...        # test
./gtd                # run
```
