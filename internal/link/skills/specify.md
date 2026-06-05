Before doing anything else: read `specdeck.yml`. If the `skills_version` field is missing or less than 2, stop and tell the user: "Your specdeck skills are out of date — run `specdeck sync` to update them, then retry."

Read `specdeck.yml` in the project root to find the `specs_repo` path.

Look at $ARGUMENTS in the context of the current codebase — read relevant files, understand the feature's scope, boundaries, and how it fits with existing behaviour.

Then draft a spec document and write it to the specs repo under a kebab-case filename matching the feature (e.g. `user-login.md`).

The spec should cover:
- What the feature does (behaviour, not implementation)
- Who uses it and when
- Key states and transitions
- Edge cases and constraints

Do not invent details that are not evident from the codebase or the feature description. Ask if anything is ambiguous.
