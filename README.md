# specdeck

A spec-management tool for making robust specs with minimal fuss.

Specdeck is designed for use with large-scale, highly stateful projects under rapid development.

## Core principles

- Git-Native


## Getting Started

A SpecDeck project is designed to exist as a standalone repository. Running `specdeck new` in a directory will create the basic directory structure a specdeck project relies on. Specdeck enforces a schema to make 

Initialise a new project in an empty git repository:

```sh
specdeck new [name]
```

If no name is provided, the directory name is used. This creates:

```
specdeck.toml       # project configuration
states/             # state and fact definitions
containers/         # component hierarchy
dist/            # build output
```
## Git-native

Specdeck is meant to be used as a git project, following paradigms you likely use with your code:

- When defining the specs for a new feature, it is beneficial to introduce those changes as a PR. Keeping a meaningful commit history results in a clear history of what work was completed when. This can later be turned into release notes or other documentation.  creates a This makes it possible to use the diff as a prompting tool when instructing an agent. And havin g
## Command Line Use

## Agentic Use

## Schema

At the core of SpecDeck is a schema built around what are calling `containers`. A `container` is simply the foundational model of the SpecDeck schema. 

### Containers

Containers are a nestable hierarchy that mirrors your application's component structure — app, tabs, screens, widgets. They live in `containers/` as a directory tree.

- **Leaf containers** are `.yml` files and hold specs.
- **Non-leaf containers** that also carry specs use an `index.yml` file alongside their subdirectories.

```
containers/
  app/
    index.yml               # specs for the app container
    home-tab/
      feed-screen/
        post-card.yml       # specs for the post card component
```
## Commands

### `specdeck new [name]`

Initialise a new project in the current directory (must be a git repo). Infers the project name from the directory if not provided.

