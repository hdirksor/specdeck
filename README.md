# specdeck

Specdeck is a tool for making your specs easier to manage in highly stateful applications. Specdeck does this by enforcing a consistent hierarchical container structure and doing basic validations.

It is highly recommended to maintain a Specdeck project lives in a git repository. It defines the possible **states** of an application (combinations of facts like language, subscription tier, or experiment enrollment) and organises **containers** — a nestable hierarchy of components — each with specs that vary per state.

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
dist/            # build output
```

## Project structure

### Facts

Facts are the atomic properties that describe a state. They live in `states/facts/` and are either boolean or enum typed.

```yaml
# states/facts/language.yml
name: language
type: enum
values: [en, es, fr, de]
```

```yaml
# states/facts/is-logged-in.yml
name: is-logged-in
type: boolean
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

#### Importing sub-containers

A container can import other containers by reference. The built output groups each import as a named section, titled from the imported container's `title` field (falling back to the filename stem).

```yaml
# containers/jot/index.yml
containers:
  - $ref: '../shared/hero.yml'
  - $ref: './noteInput.yml'
```

#### Events

Containers declare the interactions they support under an `events:` key. Each event has a trigger name (`title`), an optional `description`, and one or more typed actions nested under `actions:`.

```yaml
events:
  - title: on-press-enter
    description: user presses enter
    actions:
      navigate:
        destination: /jot/tags
```

Multiple actions per event are supported:

```yaml
events:
  - title: on-press-submit
    actions:
      navigate:
        destination: /home
      track:
        event: form_submitted
```

**Built-in action types**

| Action | Purpose | Fields |
|---|---|---|
| `navigate` | route change | `destination` (required), `transition` (optional: `push` / `modal` / `replace`) |
| `track` | analytics / logging | `event` (required), `properties` (optional map) |
| `dispatch` | trigger a background job | `job` (required), `payload` (optional map) |
| `update` | mutate local state | `target` (required), `value` (required) |
| `open` | external URL or deep link | `url` (required) |
| `dismiss` | close the current view | — |
| `prompt` | show a dialog or alert | `message` (required), `confirm_action` (optional) |

**Custom action types**

Because the action key *is* the type, any string is a valid action type — there is no allowlist. Custom types pass through as-is and render alongside built-ins:

```yaml
events:
  - title: on-press-enter
    actions:
      haptic:
        pattern: heavy
      navigate:
        destination: /home
```

Teams can freely define a shared vocabulary of custom types (e.g. `haptic`, `toast`, `permission-request`) and use them consistently across containers. The payload fields are arbitrary — specdeck does not validate them for custom types.

## Commands

### `specdeck new [name]`

Initialise a new project in the current directory (must be a git repo). Infers the project name from the directory if not provided.

