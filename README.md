# mint

`mint` is a (very) simple command line tool to track work on a software project.

![Screenshot of the program in use, in light mode, showing several issues.](https://r1vysk5peykhs5gu.public.blob.vercel-storage.com/mint-light-F2652rXYUItzrjt638pKO.png)
![Screenshot of the program in use, in dark mode, showing several issues.](https://r1vysk5peykhs5gu.public.blob.vercel-storage.com/mint-dark-02f4fTQzr0EZRoZASYaOX.png)

## Features

- A simple, intuitive command line interface.
- Built for agents and humans.
- Issues are stored as plain text in a single YAML file.
- Track dependencies between issues (depends on, blocks).
- Track status of issues (open, ready, closed).
- See issues that are ready for work (issues with no dependencies and no blockers).

## Installation

Via [Homebrew](https://brew.sh/):
```bash
brew tap mybuddymichael/tap
brew install mybuddymichael/tap/mint
```

Manually with Go:
```
go install github.com/mybuddymichael/mint@latest
```

## Usage

```bash
→ mint create "Support closing issues"
Created issue mint-a8
```

```bash
→ mint update mint-a8 --title "Support closing issues with dependencies"
Updated mint-a8 with new title "Support closing issues with dependencies"
```

```bash
→ mint update mint-a8 --depends-on mint-j0
Updated mint-j0 "Add initial code structure"
  [blocks]
    mint-a8 "Support closing issues"
```

```bash
→ mint update mint-a8 --blocks mint-8G mint-lw
Updated mint-a8 "Support closing issues"
  [blocks]
    mint-8G "Write tests for closing issues"
    mint-lw "Update README for closing issues"
```

```bash
→ mint show mint-a8
ID: mint-a8
Title: Support closing issues
Status: open
Depends on:
  mint-j0 "Add initial code structure"
Blocks:
  mint-8G "Write tests for closing issues"
```

```bash
→ mint list
READY

   mint-j0 open Add initial code structure

BLOCKED

   mint-a8 open Support closing issues

CLOSED

   mint-8G closed Write tests for closing issues
   mint-lw closed Update README for closing issues
```

```bash
→ mint list --ready
READY

   mint-j0 open Add initial code structure
```

```bash
→ mint update mint-a8 --comment "The problem is in main.go:123."
Added a comment to issue mint-a8 with text "The problem is in main.go:123."
```

```bash
→ mint close mint-a8 --reason "Done"
Closed issue mint-a8 with reason "Done"
```

```bash
→ mint open mint-a8
Re-opened issue mint-a8
```

```bash
→ mint delete mint-a8
Deleted issue mint-a8
```

```bash
→ mint set-prefix am
Prefix set to "am" and all issues updated
```

## Backend

Issues are stored as plain text in a single YAML file (`mint-issues.yaml`), and I recommend tracking it in version control. If an issue storage file isn't found, it's created when the first issue is added.

The issue storage file is created at the top level of the project, based on the nearest .git folder. If a .git directory can't be found, the file will be created in the current directory when the command is run.

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

## Disclaimer

I made it for myself. Other tools (see below) might be a better solution for you.

The interface might change as better patterns emerge.

## Prior art

- [git-bug](https://github.com/git-bug/git-bug)
- [Radicle](https://radicle.xyz/)
- [Beads](https://github.com/steveyegge/beads)

## License

MIT
