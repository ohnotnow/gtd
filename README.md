# GTD - Sysadmin Task Manager

A fast, interactive terminal tool for daily task management inspired by *Time Management for System Administrators* by Thomas Limoncelli. Plan your day with prioritised tasks, track progress, and carry incomplete work forward.

Single binary, zero dependencies, works anywhere.

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)

## Features

- **Interactive table** — navigate tasks with arrow keys, act with single keypresses
- **Priority system** — A (must do), B (should do), C (nice to do), D (delegate/defer)
- **Carry over** — push incomplete tasks to tomorrow with one keypress
- **Import tasks** — viewing an empty day? Import incomplete tasks from your most recent day
- **Portable** — single binary with embedded SQLite, no runtime dependencies
- **Cross-platform** — builds for macOS, Linux and Windows (pure Go, no CGo)

## Install

### From source

```bash
go install github.com/ohnotnow/gtd@latest
```

### Build locally

```bash
git clone https://github.com/ohnotnow/gtd.git
cd gtd
go build -o gtd .
```

### Cross-compile

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o gtd-linux .

# Windows
GOOS=windows GOARCH=amd64 go build -o gtd.exe .
```

## Usage

```bash
# Today's tasks
gtd

# Specific date
gtd 25/12/2025
```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `a` | Add a new task |
| `d` | Toggle done/not done on selected task |
| `e` / `Enter` | Edit selected task |
| `x` | Delete selected task (with confirmation) |
| `c` | Carry incomplete tasks to tomorrow |
| `i` | Import incomplete tasks from most recent day (if the current day is empty) |
| `v` | View a different day |
| `q` | Quit |
| `Esc` | Cancel current form |
| `Up` / `Down` | Navigate tasks |

### Priority levels

| Priority | Label | Meaning |
|----------|-------|---------|
| A | Must do | Critical tasks that are urgent |
| B | Should do | Important but not critical |
| C | Nice to do | Do if time permits |
| D | Delegate/defer | Hand off or postpone |

## Data storage

Tasks are stored in a SQLite database at your platform's config directory:

| Platform | Path |
|----------|------|
| macOS | `~/Library/Application Support/sysadmin-gtd/tasks.db` |
| Linux | `~/.config/sysadmin-gtd/tasks.db` |
| Windows | `%AppData%\sysadmin-gtd\tasks.db` |

## Tech stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — terminal UI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — interactive table component
- [Huh](https://github.com/charmbracelet/huh) — terminal forms and prompts
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — styled terminal output
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — pure Go SQLite driver

## Licence

[MIT](LICENCE)
