# mint

Mint (`mt`) is a (very) simple command line tool to create and track work on a software project.

I made it for myself. Other tools (see below) might be a much better solution for you.

The interface might change as bettern patterns emerge.

## Features

- A simple, intuitive command line interface.
- Built for agents.
- Issues are stored as plain text in a single YAML file.
- Track dependencies between issues (depends on, blocks, etc.).
- Track status of issues (open, closed, in progress, etc.).
- See issues that are ready for work (issues with no dependencies and no blockers).

## Usage

```bash
→ mt create "Support closing issues"
Created issue mt-a8

→ mt update mt-a8 --title "Support closing issues with dependencies"
Updated mt-a8 with new title "Support closing issues with dependencies"

→ mt update mt-a8 --depends-on mt-j0
Updated mt-j0 "Add initial code structure"
  [blocks] mt-a8 "Support closing issues"

→ mt update mt-a8 --blocks mt-8G mt-Lw
Updated mt-a8 "Support closing issues"
  [blocks]
    mt-8G "Write tests for closing issues"
    mt-Lw "Update README for closing issues"

→ mt show mt-a8
ID: mt-a8
Title: Support closing issues
Status: open
Depends on:
  mt-j0 "Add initial code structure"
Blocks:
  mt-8G "Write tests for closing issues"

→ mt ready
Issues with no blockers:
mt-j0 "Add initial code structure"

→ mt update mt-a8 --comment "The problem is in main.go:123."
Added a comment to issue mt-a8 with text "The problem is in main.go:123."

→ mt close mt-a8 --reason "Done"
Closed issue mt-a8 with reason "Done"

→ mt delete mt-a8
Deleted issue mt-a8
```

## Backend

Issues are stored as plain text in a single YAML file. If a file is not found, it's created when the first issue is added.

The file (`mint-issues.yaml`) is created at the top level of the project (based on the nearest located at mint-issues.yaml and should be tracked in version control. If a .git directory can't be found, the file will be created in the current directory when the command is run.

## Use with agents

Add something like this to your agent markdown file:

```markdown
## Issue tracking
This project uses Mint exclusively to track and manage issues. Run `mt help` to see how to use it.
```

## Stack

- Go
- urfave/cli/v3
- goccy/go-yaml

## Prior art

- [git-bug](https://github.com/git-bug/git-bug)
- [Radicle](https://radicle.xyz/)
- [Beads](https://github.com/steveyegge/beads)
