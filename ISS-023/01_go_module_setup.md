# Task 01 — Go Module Setup

**Depends on:** Nothing  
**Produces:** Compilable (empty) `atd` binary  

## Context

Currently `scripts/go.work` references 8 individual tool modules. We need a single unified module at `scripts/cmd/atd/` that will host all subcommands.

## Steps

### 1.1 Create directory structure

```bash
mkdir -p scripts/cmd/atd/cmd
mkdir -p scripts/internal/atom
mkdir -p scripts/internal/ollama
mkdir -p scripts/internal/cosine
mkdir -p scripts/internal/prompt
mkdir -p scripts/internal/pipeline
```

### 1.2 Create `scripts/cmd/atd/go.mod`

```
module atd

go 1.24.4

require (
    github.com/spf13/cobra v1.8.1
    github.com/mattn/go-sqlite3 v1.14.24
    atd-tools v0.0.0
)

replace atd-tools => ../../
```

Run `cd scripts/cmd/atd && go mod tidy`.

### 1.3 Create `scripts/cmd/atd/main.go`

```go
package main

import "atd/cmd"

func main() {
    cmd.Execute()
}
```

### 1.4 Create `scripts/cmd/atd/cmd/root.go`

Minimal root command:

```go
package cmd

import (
    "fmt"
    "os"
    "atd-tools/config"
    "github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
    Use:   "atd",
    Short: "Atomic Traceable Documentation toolkit",
    Long: `atd is the unified CLI for managing Atomic Traceable Documentation.
It provides subcommands for dissecting documents, auditing atoms,
indexing codebases, searching semantically, and more.

Configuration is loaded from the .atd file found at the project root.`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return config.Load()
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
```

### 1.5 Update `scripts/go.work`

Add the new module. Keep existing modules for now (they'll be removed when their code is fully migrated):

```
go 1.24.4

use (
    .
    ./cmd/atd
    ./atd-audit-fixer
    ./atd-compare
    ./atd-ollama-audit
    ./atd-ollama-generate
    ./atd-ollama-indexer
    ./atd-ollama-search
    ./atd-roadmap-builder
    ./atd-update
)
```

### 1.6 Build and verify

```bash
cd scripts && go build -o bin/atd ./cmd/atd/
./bin/atd --help
```

## Acceptance Criteria

- [ ] `./bin/atd --help` prints the tool description and `--verbose` flag
- [ ] No compilation errors
- [ ] `go.work` includes the new module
