# mint

`mint` is a (very) simple command line tool to track work on a software project.

I made it for myself. Other tools (see below) might be a much better solution for you.

The interface might change as better patterns emerge.

## Features

- A simple, intuitive command line interface.
- Built for agents and humans.
- Issues are stored as plain text in a single YAML file.
- Track dependencies between issues (depends on, blocks, etc.).
- Track status of issues (open, closed, in progress, etc.).
- See issues that are ready for work (issues with no dependencies and no blockers).

## Usage

```bash
→ mint create "Support closing issues"
Created issue mint-a8

→ mint update mint-a8 --title "Support closing issues with dependencies"
Updated mint-a8 with new title "Support closing issues with dependencies"

→ mint update mint-a8 --depends-on mint-j0
Updated mint-j0 "Add initial code structure"
  [blocks]
    mint-a8 "Support closing issues"

→ mint update mint-a8 --blocks mint-8G mint-lw
Updated mint-a8 "Support closing issues"
  [blocks]
    mint-8G "Write tests for closing issues"
    mint-lw "Update README for closing issues"

→ mint show mint-a8
ID: mint-a8
Title: Support closing issues
Status: open
Depends on:
  mint-j0 "Add initial code structure"
Blocks:
  mint-8G "Write tests for closing issues"

→ mint list
All issues:
mint-8G open "Write tests for closing issues"
mint-a8 open "Support closing issues"
mint-j0 open "Add initial code structure"
mint-lw open "Update README for closing issues"

→ mint ready
Issues with no blockers:
mint-j0 "Add initial code structure"

→ mint update mint-a8 --comment "The problem is in main.go:123."
Added a comment to issue mint-a8 with text "The problem is in main.go:123."

→ mint close mint-a8 --reason "Done"
Closed issue mint-a8 with reason "Done"

→ mint open mint-a8
Re-opened issue mint-a8

→ mint delete mint-a8
Deleted issue mint-a8

→ mint set-prefix am
Prefix set to "am" and all issues updated
```

## Backend

Issues are stored as plain text in a single YAML file. If a file is not found, it's created when the first issue is added.

The file (`mint-issues.yaml`) is created at the top level of the project (based on the nearest located at mint-issues.yaml and should be tracked in version control. If a .git directory can't be found, the file will be created in the current directory when the command is run.

## Use with agents

Add something like this to your agent markdown file:

```markdown
## Issue tracking
This project uses Mint exclusively to track and manage issues. Run `mint help` to see how to use it.
```

## Stack

- Go
- urfave/cli/v3
- goccy/go-yaml

## Prior art

- [git-bug](https://github.com/git-bug/git-bug)
- [Radicle](https://radicle.xyz/)
- [Beads](https://github.com/steveyegge/beads)

## License

MIT
