# Specdeck

[![codecov](https://codecov.io/gh/hdirksor/specdeck/graph/badge.svg)](https://codecov.io/gh/hdirksor/specdeck)

A spec-management tool for making robust specs with minimal fuss.

Specdeck is designed with large-scale projects under continuous development in mind.

## Overview

Specdeck is a command line tool written in Go to manage specs. At its core, Specdeck enforces an opinionated approach to writing specs. So in a sense, Specdeck is also that specific approach to writing specs. See more about this approach in [writing specs with SpecDeck.](#writing-specs)

## Installation:

Specdeck can be installed by running 
`go install github.com/hdirksor/specdeck@latest`

### Update notices

When a newer release is out, specdeck prints a one-line notice on stderr, at most once a day. It checks GitHub for the latest release in the background, at most once a day, so commands never wait on the network. The check is skipped in CI (when `CI` is set), when stderr is not a terminal, and for local builds. To turn it off, set `SPECDECK_NO_UPDATE_CHECK=1`.

After upgrading, run `specdeck agents` in each code repo to update its agent commands. Until you do, specdeck reminds you when it runs in that repo.

## Getting Started

A Specdeck project is designed to exist as a standalone git repository. To get started run `specdeck new` in a directory. This will create: 

- `specdeck.yml`
- `containers/`

You will also want to run `specdeck link /path/to/spec/repo` in any repositories used to build the product being spec'd. `specdeck link` records the specs repo path in a `specdeck.yml` config file in your code repo. If you use an AI coding agent, also run `specdeck agents claude` to install specdeck's commands for it.

## Writing Specs

As a tool designed to help spec continuously developed projects it is only right to treat your spec, like your code, as a continuously developed entity. As such, Specdeck is intended to be git-dependent and the expected way to introduce a change to your specs is through a pull request or new branch. 

### Images

Drop an image next to a container's YAML file, matching its filename stem, and `build` embeds it at the top of that container's section and copies it alongside the generated markdown:

```
containers/
  screen-one/
    index.yml
    index.png    # rendered at the top of screen-one/_index.md
```

Supported extensions: `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg`, `.webp`. If more than one matches the same stem (e.g. `index.png` and `index.svg`), `build` fails with an ambiguous-image error — remove one.

This only applies to a container's own file. A container embedded elsewhere via `$ref` does not carry its image into the importing container's output.

## Agentic Use

Running `specdeck agents claude` in a code repo installs Claude Code commands in `.claude/commands/specdeck/` and records `claude` under `agents` in `specdeck.yml`. Run `specdeck agents` with no names to update every recorded agent. specdeck owns that folder: it overwrites the files there and removes ones it no longer ships, so keep your own commands outside it.

- `/specdeck:specify` — point Claude at specific code or documentation and it will draft a spec document from it.

## Types

Specdeck is built on a recursive type system with `Container` being the core. Each container may have many or no child `Containers`.

## Commands

- `specdeck new [name]` — Initialise a new specdeck project in the current directory. (This is the repo where your specs will live, not your code.)
- `specdeck link <specs-repo-path>` — Link a code repository to a specdeck specs repo by recording the path in `specdeck.yml`. Re-run it to point at a different specs repo.
- `specdeck agents [agent...]` — Install or update specdeck's commands for AI coding agents (supported: `claude`). With no names, updates every agent recorded in `specdeck.yml`.
- `specdeck build` — Resolve container refs and write flat specs to `dist/`.
- `specdeck --version` — Print the installed version.

## License

MIT
