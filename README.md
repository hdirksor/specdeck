# specdeck

Specdeck is a tool for managing UI specs as structured YAML. It organises specs into a nestable hierarchy of **containers** — screens, components, widgets — each with behaviour notes, key-value specs, states, and events. The `build` command resolves cross-file references and writes the full spec tree to markdown.

## Project structure

A specdeck project is a directory (ideally a git repo) with:

```
containers/     # YAML spec files
dist/           # markdown build output (generated)
```

## Containers

Containers are YAML files in `containers/`. They mirror your application's component hierarchy and can be nested arbitrarily deep via directory structure or inline definitions.

```
containers/
  screen-one/
    index.yml           # screen-level container
    entity.yml
    text-fiel.yml
  shared/
    nav-bar.yml
    card.yml
```

### Schema

```yaml
title: List
description: Optional — one sentence summary.

behavior:
  - scrollable vertically
  - shows zero state when empty

specs:
  border: 2px white
  content-padding: 16px
  scroll-direction: vertical

states:
  - title: Loading
    description: Shown while the initial fetch is in progress.
    specs:
      content-padding: 0dp   # overrides base spec for this state

events:
  - title: submit
    description: Emitted by the submit button when input is not blank.
    effects:
      - title: save note
        description: Persist the trimmed input text.
      - title: clear input

containers:
  - $ref: ../shared/nav-bar.yml
  - $ref: ./note-card.yml
    overrides:
      border: none
  - title: Inline sub-container
    specs:
      label: inline
```

**`specs`** are flat key-value pairs. Values are strings.

**`states`** each carry their own `specs` that override the container's base specs for that state. A state with no specs of its own inherits the base specs unchanged.

**`events`** describe interactions. Each event has a `title`, optional `description`, and a list of `effects` (each with `title` and optional `description`).

**`containers`** can be `$ref` imports, inline definitions, or a mix. `$ref` entries accept an optional `overrides:` map to patch individual specs in the imported container.

### Cross-file references

Any container can import another with `$ref`:

```yaml
containers:
  - $ref: ../shared/nav-bar.yml
```

Paths are relative to the current file. `build` resolves all refs recursively before writing output.

### `index.yml`

A directory that also carries specs of its own uses an `index.yml` file. In the built output this becomes `_index.md` (Hugo section page).

## Command Line and Agentic Use

### `specdeck new [name]`

Initialise a new specdeck project in the current directory. Infers the project name from the directory if not provided.

### `specdeck build`

Resolve all `$ref` entries and write each container to `dist/` as markdown. Leaf containers become `<name>.md`; directories with `index.yml` become `_index.md`.

```
🟢  dist/screen-one/index.md
🟢  dist/screen-one/list.md
🟢  dist/shared/nav-bar.md
```

### `specdeck link [specs-repo-path]`

Run this in a **code repository** to link it to a specdeck specs repo. Creates `specdeck.yml` and installs Claude Code skills under `.claude/commands/` so Claude has context about the spec structure.

```sh
specdeck link ../my-app-specs
```

### `specdeck sync`

Re-writes the Claude Code skill files to the current specdeck version without changing `specdeck.yml`. Run this after upgrading specdeck.
