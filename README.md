# notgit

`notgit` is a minimal reimplementation of Git written in Go i.e git but worse.
It exists to understand how version control actually works under the hood.
Simple, inspectable, and intentionally minimal.

---

## Table of Contents

* [Overview](#overview)
* [Features](#features)
* [Libraries Used](#libraries-used)
* [Project Structure](#project-structure)
* [Build and Prerequisites](#build-and-prerequisites)
* [Usage](#usage)
* [What I Learned](#what-i-learned)
* [License](#license)

---

## Overview

Git is everywhere. But few developers understand what happens when you type `git add` or `git commit`.
`notgit` rebuilds those internals piece by piece using Go.
The goal is not to replace Git, but to demystify it.

It focuses on the core concepts: repository, index, blob, tree, commit, and branch.

---

## Features

* `init` - Initialize a new notgit repository
* `add` - Add file contents to the index
* `commit` - Record changes to the repository
* `branch` - List, create, or delete branches
* `switch` - Move between branches
* `merge` - Merge branch histories
* `cat-file` - Inspect raw object data
* `config` - Manage repository settings
* `log` - View commit history
* `status` - Show current working tree state

Use `notgit [command] --help` for more information about a command.

---

## Libraries Used

### spf13/cobra

[**spf13/cobra**](https://github.com/spf13/cobra) is a modern library for building command-line interfaces in Go.
It provides a clean and structured way to define commands, subcommands, and flags.
`notgit` uses Cobra to manage commands like `init`, `add`, `commit`, and `branch`.
It automatically handles help text, argument parsing, and usage hints.
This keeps the CLI experience consistent with tools like Git itself.

### stretchr/testify

[**stretchr/testify**](https://github.com/stretchr/testify) is a Go testing toolkit that makes writing assertions and test cases easier.
It provides packages like `assert` and `require` for clear and concise test validation.
`notgit` uses the `require` package to ensure expected behavior in tests for objects, trees, commits, and repository operations.
It helps maintain reliability while keeping the test code readable.

---

## Project Structure

```
.
├── cmd/
│   └── notgit/
│       └── main.go          # CLI entry point
├── internal/
│   ├── commands/            # CLI commands (add, commit, branch, etc.)
│   ├── objects/             # Git object types (blob, tree, commit)
│   ├── repository/          # Repository logic (index, storage)
│   └── utils/               # Shared helpers (config, repo utils)
└── README.md
```

Each component is focused and self-contained.
There is no hidden magic, just Go structs, files, and hash calculations.

---

## Build and Prerequisites

### Requirements

* Go 1.24.5 or higher
* A terminal
* Basic understanding of how Git works (optional but helpful)

### Build

```bash
git clone https://github.com/Gr1shma/notgit.git
cd notgit
go build -o notgit ./cmd/notgit
```

Now run it:

```bash
./notgit
```

---

## Usage

### Initialize a repository

```bash
notgit init
```

### Add a file

```bash
notgit add README.md
```

### Commit changes

```bash
notgit commit -m "Initial commit"
```

### Check status

```bash
notgit status
```

### View logs

```bash
notgit log
```

For any command details:

```bash
notgit [command] --help
```

---

## What I Learned

Building Git from scratch changes how you see software tools.
You realize version control is not magic; it's structured data and smart design.

* Version control is just a set of objects linked by hashes.
* The index is a binary map connecting file paths to object hashes.
* Commits store snapshots of trees, not diffs.
* Branches are simple text files that point to commit hashes.
* Merging is graph traversal and conflict resolution.
* The simplicity of Git’s core makes its complexity justifiable.
* Go’s strong typing and file I/O model make system-level work approachable.
* Libraries like Cobra and Testify streamline CLI and testing without hiding complexity.

Rebuilding Git taught me how data models drive developer tools.
It made me respect the engineering behind everyday commands we take for granted.

---

## License

MIT License – see the [LICENSE](LICENSE) file.

Use it, learn from it, modify it.
The goal is education, not replacement.

---

*"Understanding a tool is the first step to mastering it."*
