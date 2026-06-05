Run `specdeck change new "$ARGUMENTS"` to create a new change record.

Then open the generated file and fill in the sections based on the current branch and any available context (commit messages, PR description, linked issues):
- **What changed**: summarise the spec changes this work introduces
- **Why**: extract motivation from commit messages, PR description, or ask the user
- **References**: include any linked tickets, PRs, or Slack threads

Print the path to the created file when done.
