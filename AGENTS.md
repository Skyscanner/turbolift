# Repository Guidelines

`README.md` is the source of truth for installation, usage, and development. Use these anchors: Philosophy, Basic usage, Detailed usage, Caveats, Local development, Available Development Tasks.

- Tooling: pinned via `.mise.toml` (Go 1.23, golangci-lint v2.1.5, goreleaser v2.12.7). Run `mise install`; `mise run dev` prepares deps + lint; `mise run test` runs lint + race/cover; `mise run fix` applies gofmt/golangci-lint fixes.
- Project structure: CLI entrypoint `main.go`; commands in `cmd/`; shared logic under `internal/<domain>`; tests co-located as `_test.go`; demo asset `docs/demo.gif`; release artifacts in `dist/` (clean with `mise run clean`); fakes for git/GitHub/filesystem interactions in `internal/testsupport`.
- Style: standard Go formatting; package names lower_snakecase; include the Apache 2.0 header from `CONTRIBUTING.md` when creating files.
- Usage workflow: follow README sections noted above; when operating turbolift across many repos, use `repos.txt` subsets and `turbolift foreach --failed/--successful` to reduce blast radius (see Caveats).
- Operational notes: authenticate GitHub CLI (`gh auth login`; set `GH_HOST` for GHE). Use `mise run build`/`mise run install` for local binaries; `mise run release` drives GoReleaser via the managed binary.
