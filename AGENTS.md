# Repository Guidelines

`README.md` is the source of truth for installation, usage, and development. Use the links below to jump to the right sections.

- Usage & workflow: `README.md` sections on Philosophy, Basic usage, Detailed usage, and Caveats.
- Development setup: `README.md#local-development` and `#available-development-tasks` outline mise tasks; run `mise install` then `mise run dev`/`mise run test`.
- Project structure: CLI entrypoint `main.go`; commands in `cmd/`; shared logic under `internal/<domain>`; tests co-located as `_test.go`; demo asset `docs/demo.gif`; local release artifacts land in `dist/`.
- Style & linting: standard Go formatting; prefer `mise run fix` for gofmt/golangci-lint; package names lower_snakecase; add the Apache 2.0 header from `CONTRIBUTING.md` when creating files.
- Testing tips: keep table-driven tests near code; use `mise run test` for lint + race/cover; fakes live in `internal/testsupport` for Git/GitHub/filesystem interactions.
- Operational notes: authenticate `gh` (`gh auth login`, `GH_HOST` for GHE); when running turbolift across many repos, use `repos.txt` subsets and `turbolift foreach --failed/--successful` (see usage/caveats in `README.md`) to reduce blast radius.
