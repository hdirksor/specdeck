# Specdeck

A spec-management tool for making robust specs with minimal fuss.

Specdeck is designed with large-scale projects under continuous development in mind.

## Overview

Spececk is a command line tool written in Go to manage specs. At its core, Specdeck enforces an opinionated approach to writing specs. So in a sense, Specdeck is also that specific approach to writing specs. See more about this approach in [writing specs with SpecDeck.](#writing-specs)

## Installation:

Specdeck can be installed by running 
`go get hdirksor/specdeck`

## Getting Started

A Specdeck project is designed to exist as a standalone git repository. To get started run `specdeck new` in a directory. This will create: 

- `specdeck.toml`
- `containers/`

You will also want to run `specdeck link /path/to/spec/repo` in any repositories used to build the product being spec'd. `specdeck link` creates a toml config file in your code repo and adds some agentic commands if you are using Claude.

## Writing Specs

As a tool designed to help spec continuously developed projects it is only right to treat your spec, like your code, as a continuously developed entity. As such, Specdeck is intended to be git-dependent and the expected way to introduce a change to your specs is through a pull request or new branch. 

## Agentic Use

If you are using Claude, running `specdeck link` will create two new commands in your claude configuration. 

- `/specdeck spec` can be used to point Specdeck to specific code or documentation and begin generating specs from it. 

- `/specdeck build` can be used to point Claude to specific parts of your spec to begin building.

## Types

Specdeck is built on a recursive type system with `Container` being the core. Each container may have many or no child `Containers`.<D-s>


## Commands

- `specdeck new [name]` — Initialise a new specdeck project in the current directory.
- `specdeck link [specs-repo-path]` — Link a code repository to a specdeck specs repo. Creates `specdeck.yml` and installs Claude Code skills in the current directory.
- `specdeck build` — Resolve container refs and write flat specs to `dist/`.
- `specdeck validate` — Validate cross-references in the specdeck project.
- `specdeck change new <title>` — Create a new change record in `changes/`.
- `specdeck sync` — Re-write all Claude Code skill files without changing `specdeck.yml` configuration.


