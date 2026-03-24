# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2026-03-24
### Added
- Search/filter tasks by pressing `/` (case-insensitive substring match on task name)
- Jump to a task by pressing `1`-`9`
- CLAUDE.md and TECHNICAL_OVERVIEW.md project documentation

## [1.0.6] - 2026-03-12
### Added
- In-progress status for tasks — press `s` to toggle
- Status symbols in the table: `▶` for in-progress, `✓` for done

### Changed
- Carry-over and import now include in-progress tasks (previously only todo)
- In-progress status is preserved when carrying over or importing tasks

## [1.0.5] - 2026-02-16
### Changed
- Minor internal cleanup

## [1.0.4] - 2026-02-13
### Added
- Screenshot in README

## [1.0.3] - 2026-02-13
### Added
- `--context` flag to keep separate task lists (e.g. work, personal)
- Contexts are fully isolated — each has its own tasks, carry-over, and import

## [1.0.2] - 2026-02-13
### Added
- `--print` flag for non-interactive output (useful for scripting and piping)

## [1.0.1] - 2026-02-13
### Changed
- README updates

## [1.0.0] - 2026-02-13
### Added
- Initial release — interactive terminal task manager
- Priority system (A–D)
- Carry-over incomplete tasks to tomorrow
- Import tasks from most recent day
- SQLite storage with automatic setup
- GitHub Actions CI for cross-platform builds

[Unreleased]: https://github.com/ohnotnow/gtd/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/ohnotnow/gtd/compare/v1.0.6...v1.1.0
[1.0.6]: https://github.com/ohnotnow/gtd/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/ohnotnow/gtd/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/ohnotnow/gtd/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/ohnotnow/gtd/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/ohnotnow/gtd/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/ohnotnow/gtd/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/ohnotnow/gtd/releases/tag/v1.0.0
