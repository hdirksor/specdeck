# AGENTS.md

This file provides guidance to Agents when working with code in this repository.

## Commands

```bash
go build ./...          # compile
go run .                # run the CLI
go test ./...           # run all tests
go test ./cmd/...       # run tests in a specific package
go test -run TestName   # run a single test
go vet ./...            # lint
```

## Architecture

`main.go` is the entry point; it delegates immediately to `cmd.Execute()`.

**`cmd/`** — all Cobra commands live here. `root.go` defines the root command and `Execute()`. Each subcommand gets its own file (e.g. `cmd/build.go`). Commands that need interactivity launch a BubbleTea program inline.

### Cobra + BubbleTea pattern

Traditional commands run logic directly in their `RunE` func. Interactive commands construct a `tea.Program` and call `.Run()` from within `RunE`:

```go
func runInteractive(cmd *cobra.Command, args []string) error {
    p := tea.NewProgram(initialModel())
    _, err := p.Run()
    return err
}
```

BubbleTea models live in `internal/tui/` (to be created). Lip Gloss styles are defined alongside the models that use them.
