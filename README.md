# specdeck

Specdeck is a cli with a few simple commands for managing and efficiently generating specs.


A Specdeck project lives in a git repository. It defines the possible **states** of an application (combinations of facts like language, subscription tier, or experiment enrollment) and organises **containers** — a nestable hierarchy of components — each with specs that vary per state.

## Getting started

Initialise a new project in an empty git repository:

```sh
specdeck new [name]
```

If no name is provided, the directory name is used. This creates:

```
specdeck.toml       # project configuration
states/             # state and fact definitions
containers/         # component hierarchy
exports/            # build output
```

## Project structure

### Facts

Facts are the atomic properties that describe a state. They live in `states/facts/` and are either boolean or enum typed. Each fact has a `scope` that controls how it participates in state generation:

- `cross` — multiplied fully against all other `cross` facts
- `isolated` — varied one at a time against the default values of all other facts
- `manual` — never auto-generated (default if omitted)

```yaml
# states/facts/language.yml
name: language
type: enum
values: [en, es, fr, de]
scope: isolated
```

```yaml
# states/facts/is-logged-in.yml
name: is-logged-in
type: boolean
scope: cross
```

### States

States are defined in YAML files under `states/`, grouped by domain or feature area. Each file contains a list of states that reference facts.

```yaml
# states/checkout.yml
- name: default
  summary: Standard logged-in English user going through checkout
  facts:
    language: en
    is-logged-in: true

- name: spanish-premium-checkout
  summary: Spanish-speaking premium user in the new checkout experiment
  facts:
    language: es
    subscription: premium
    experiment: checkout-v2
```

State names are unique across all state files. The state named `default` is the baseline state unless overridden (see below).

States can also be generated automatically from fact scopes — see `specdeck generate states` below. Generated states land in `states/generated.yml`. Any manually defined state with the same name as a generated one takes precedence.

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

#### Defining specs

All specs must be defined within a state reference. There is no implicit inheritance between states — each state's spec set is its complete, self-contained truth.

```yaml
# containers/app/home-tab/feed-screen/post-card.yml
states:
  - ref: default
    specs:
      background-color: "#FFFFFF"
      title-text-color: "#000000"
      title-text: "Latest posts"

  - ref: spanish-premium-checkout
    specs:
      background-color: "#FFFFFF"
      title-text-color: "#000000"
      title-text: "Últimas publicaciones"
```

Specs can be defined in shorthand (string value) or verbose form (with an optional description):

```yaml
specs:
  background-color: "#FFFFFF"
  title-text:
    value: "Latest posts"
    description: Heading shown at the top of the feed
```

#### Default state

By default, the state named `default` is the baseline. To use a differently named state as the baseline, declare it at the top of the container file:

```yaml
default: baseline
states:
  - ref: baseline
    specs:
      background-color: "#FFFFFF"
  - ref: dark-mode
    specs:
      background-color: "#1A1A1A"
```

Top-level `specs:` (not nested under a state) are shorthand for the default state:

```yaml
# these specs implicitly belong to the 'default' state
specs:
  background-color: "#FFFFFF"
  title-text: "Hello"
```

## Commands

### `specdeck new [name]`

Initialise a new project in the current directory (must be a git repo). Infers the project name from the directory if not provided.

### `specdeck generate states`

Generates `states/generated.yml` from fact scopes. `cross` facts are fully cartesian-producted; `isolated` facts each contribute one state per non-default value; `manual` facts are skipped. The all-defaults combination is always named `default`. State names are built from non-default values joined with `-`, e.g. `logged-in-premium-dark`.

Re-running overwrites `generated.yml` entirely. To override a generated state, define a state with the same name in any other file under `states/`.

### `specdeck add state <name>`

Propagates a named state into all existing leaf containers that don't already have it, copying each container's default specs as a starting point. The state must already be defined in `states/` before running this command.

```sh
specdeck add state dark-mode
```
