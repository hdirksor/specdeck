# Implementation Plan: update notice

Tell interactive users when a newer specdeck release exists, without slowing
commands down or self-updating.

Decisions:
- Repo is public; use the unauthenticated GitHub "latest release" API (stable releases only).
- The notice is read from a cache; a detached `specdeck __update-check` child
  refreshes the cache at most once a day, so commands never wait on the network.
- The notice is shown at most once a day, on stderr, only when stderr is a terminal.
- No check when `CI` or `SPECDECK_NO_UPDATE_CHECK` is set, for local builds, or for
  completion, `help`, and `--version`.
- Upgrade hint depends on install method: `go install` → `go install …@latest`;
  release download → release link only. Homebrew comes later with a tap.

## Stage 1: Version detection and --version
Goal: specdeck knows its own version and how it was installed.
Success Criteria:
- goreleaser stamps the release tag into the binary.
- A clean `vX.Y.Z` from Go build info is treated as a `go install`; pseudo-versions,
  `+dirty`, and `(devel)` are local builds.
- `specdeck --version` prints the version.
Status: Complete

## Stage 2: Cache and notice decision
Goal: decide from the cached state whether to print a notice, and what it says.
Success Criteria:
- Notice only when the cached latest version is newer than the current one.
- Notice at most once per 24h; showing it records the time.
- Message includes the release link, plus the upgrade command for `go install`.
- Missing or corrupt cache means no notice and no error.
Status: Complete

## Stage 3: Refresh from GitHub
Goal: fetch the latest release into the cache.
Success Criteria:
- Reads `tag_name` and `html_url` from the latest-release API (tested with httptest).
- 2s timeout; failures are silent and leave the previous latest version intact.
- Refresh is due only when the last attempt is older than 24h.
Status: Complete

## Stage 4: CLI wiring
Goal: run the notice and the background refresh from `cmd.Execute`.
Success Criteria:
- Gating by env, terminal, build, and command as listed above.
- Hidden `__update-check` command performs the refresh.
- The parent records the attempt before spawning the child so runs don't pile up.
Status: Complete

## Stage 5: Docs
Goal: document the notice and opt-out; fix the stale command list in `main.go`.
Success Criteria: README and `main.go` match the real commands.
Status: Complete

## Stage 6: Stale project commands
Goal: remind users to run `specdeck link` when installed commands are older than the binary.
Success Criteria:
- A repo is linked when `.claude/commands/specdeck/` exists in the current directory.
- A skill is stale when it is missing or its content differs from the embedded copy; extra files are ignored.
- Notice shows on every run while stale, under the same terminal/CI/opt-out rules, and never for `link`.
Status: Complete

## Stage 7: Code-repo config package
Goal: one place that reads and writes the code repo's `specdeck.yml` (`specs_repo`, `agents`).
Success Criteria:
- Missing file loads as empty; malformed file is an error naming specdeck.yml.
- Unknown old fields (e.g. `skills_version`) are ignored on load.
Status: Complete

## Stage 8: `link` only links
Goal: `specdeck link <path>` sets `specs_repo` and nothing else.
Success Criteria:
- Requires exactly one path; re-running updates `specs_repo` and keeps `agents`.
- Does not touch `.claude/`.
- Prints `Linked to <path>.` and, when no agents are recorded, a hint to run `specdeck agents claude`.
Status: Complete

## Stage 9: `specdeck agents`
Goal: install and update agent commands separately from linking.
Success Criteria:
- `agents claude` installs/updates `.claude/commands/specdeck/` and records `claude` in `agents`.
- `agents` with no names updates every recorded agent; errors with a hint when none are recorded.
- Unknown names fail with the supported list.
- Files in an agent's folder that specdeck no longer ships are removed.
- No `--check` flag for now (decided 2026-10-23).
Status: Complete

## Stage 10: Notice and docs
Goal: the out-of-date notice points at `specdeck agents`; docs and memory match.
Success Criteria:
- Notice checks recorded agents (including extra files) and is skipped for `agents`.
- README, `main.go`, and memory notes describe `link` and `agents` separately.
Status: Complete
