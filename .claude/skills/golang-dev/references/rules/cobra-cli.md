# Cobra CLI Rules

**Rules for structuring the CLI with Cobra (`github.com/spf13/cobra`) — `main.go` + `cmd/` package pattern.**

---

## Overview

- **Tool:** [spf13/cobra](https://github.com/spf13/cobra) — CLI framework
- **Structure:** thin `main.go` → `cmd/` package owns commands
- **Entry:** `cmd.Execute()` is the only exported function; `main()` calls it
- **Registration:** each subcommand lives in its own file and self-registers via `init()`

**Core rules:**
- ✅ Keep `main.go` minimal — recovery/observability + `cmd.Execute()` only
- ✅ Use `RunE` (returns error), not `Run` — let `Execute()` handle exit codes
- ✅ One file per subcommand (`cmd/<name>.go`), registered in its own `init()`
- ✅ Build logger + load config in the composition root (`RunE`), NOT via `cobra.OnInitialize` globals
- ✅ Build the app inside `RunE` (composition root) — see [dependency-injection.md](dependency-injection.md)
- ❌ Never call `os.Exit()` inside `RunE` — return the error instead
- ❌ Never put business logic in `cmd/` — delegate to the composition root

---

## Entry Point (`main.go`)

**`main()` does only panic/observability setup, then hands off to `cmd.Execute()`:**

```go
package main

import (
    "fmt"
    "runtime/debug"
    "time"

    "github.com/getsentry/sentry-go"
    _ "go.uber.org/automaxprocs"          // side-effect: set GOMAXPROCS to container quota

    "github.com/haipham22/golang-sample/cmd"
)

func main() {
    defer sentry.Flush(2 * time.Second)   // flush telemetry before exit
    defer func() {                        // capture panic for logging + re-report
        if r := recover(); r != nil {
            fmt.Println(string(debug.Stack()))
            defer sentry.Recover()
            panic(r)
        }
    }()

    cmd.Execute()                          // single handoff to CLI
}
```

**Rules:**
- ✅ `main()` is ~10–20 lines: defer recovery/telemetry, call `cmd.Execute()`
- ✅ Side-effect imports (`automaxprocs`) belong here, not in `cmd/`
- ❌ Never parse flags, load config, or build the app in `main()`

---

## Root Command (`cmd/root.go`)

**`rootCmd` is an unexported package var; `Execute()` is the only exported entry:**

```go
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "golang-sample",
    Short: "Sample Golang application with best practices",
}

// Execute runs the root command and exits on error.
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)                       // single place that sets exit code
    }
}

var cfgFile string

func init() {
    // Optional config-file flag; default ".env". Add only if the app loads config from a file.
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".env", "config file (optional)")
}
```

**Rules:**
- ✅ `rootCmd` unexported — callers use `Execute()`, never the var directly
- ✅ `os.Exit` lives **only** in `Execute()` (and `main` recovery) — nowhere else
- ✅ Root-level flags are `PersistentFlags` (inherited by subcommands)
- ✅ The `--config` flag is **optional** — provide a default; apps may use env/defaults instead of a file
- ✅ `init()` registers flags only — do NOT build logger/config here (see *Global Setup* below)
- ❌ Never give `rootCmd` a `Run`/`RunE` unless the binary has no subcommands
- ❌ Never export `rootCmd` — subcommands register onto it internally

---

## Subcommand Registration

**One file per subcommand; each self-registers via `init()`:**

```go
// cmd/serverd.go
package cmd

import (
    "github.com/haipham22/golang-sample/internal/handler/rest"
    // ...
)

var serverCmd = &cobra.Command{
    Use:   "serverd",
    Short: "Start production API server with govern integration",
    RunE: func(cmd *cobra.Command, _ []string) error {
        // ... parse flags, build app, run (see Composition Root below)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(serverCmd)        // self-register
    serverCmd.Flags().Int64("port", 8080, "API server port")
    serverCmd.Flags().Int64("shutdown_time", 10, "Graceful shutdown timeout (s)")
}
```

**Rules:**
- ✅ One command per file, named after the command (`cmd/serverd.go`)
- ✅ Register in the file's `init()` — no central registration file to forget
- ✅ Command-local flags via `cmd.Flags()` (NOT `PersistentFlags`)
- ❌ Never register all commands in one giant `init()` in `root.go`

---

## `RunE` over `Run`

**Always use `RunE` so errors propagate to `Execute()` for a clean exit code:**

```go
// GOOD — RunE returns error; Execute() prints + os.Exit(1)
RunE: func(cmd *cobra.Command, _ []string) error {
    port, err := cmd.Flags().GetInt64("port")
    if err != nil {
        return err
    }
    // ...
    return err
},

// BAD — Run can't return error; forces manual os.Exit (breaks Execute())
Run: func(cmd *cobra.Command, _ []string) {
    os.Exit(1)                           // hides error from Execute(), no cleanup defer
},
```

**Rules:**
- ✅ `RunE` returns `error`; `Execute()` handles printing + exit
- ✅ Return wrapped errors: `return fmt.Errorf("start server: %w", err)`
- ❌ Never `os.Exit` inside `RunE` — skips deferred cleanups (`defer cleanup()`)

---

## Flags

**Persistent (root, inherited) vs local (command-only):**

```go
// Root — inherited by ALL subcommands (defined in root.go init())
rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".env", "config file")
// → `golang-sample --config prod.env serverd`

// Command-local — only on this command (defined in serverd.go init())
serverCmd.Flags().Int64("port", 8080, "API server port")
// → `golang-sample serverd --port 9090`

// Reading flags inside RunE
port, err := cmd.Flags().GetInt64("port")
if err != nil {
    return err
}
```

**Rules:**
- ✅ Cross-command flags (config, verbosity) → `PersistentFlags`
- ✅ Command-specific flags (port, timeout) → local `Flags()`
- ✅ Provide sensible defaults so flags are optional
- ✅ Validate flag values in `RunE` (range checks, required combos)
- ❌ Never read `os.Args` directly — let Cobra parse

---

## Global Setup — prefer the composition root

**Logger and config are application infrastructure. Build them in the composition root (`RunE` → `bootstrap.NewApp`), NOT in `cobra.OnInitialize`.**

The current `cmd/root.go` uses `cobra.OnInitialize(initDependency)` → `initLog()` / `initConfig()`, which set **globals** (`zap.ReplaceGlobals`, `config.ENV`). This is a **legacy pattern** — migrate toward the composition root:

```go
// PREFER — construct logger + load config inside RunE, pass explicitly (no globals)
RunE: func(cmd *cobra.Command, _ []string) error {
    cfgFile, _ := cmd.Flags().GetString("config")
    log := bootstrap.NewLogger()
    cfg, err := config.LoadConfig(cfgFile, log)
    if err != nil {
        return fmt.Errorf("load config: %w", err)   // fail fast → Execute() exits
    }
    app, cleanup, err := bootstrap.NewApp(cfg, log)
    // ...
}
```

**Why NOT `OnInitialize` for logger/config:**
- ⚠️ Runs for **every** command including `--help` and unknown commands — wasteful to load full config there, can fail spuriously
- ⚠️ No cleanup path — logger/config resources can't be torn down
- ⚠️ Globals (`zap.ReplaceGlobals`, `config.ENV`) hide dependencies and hurt testability

**Rules:**
- ✅ Build logger + load config in the composition root; pass them down explicitly
- ✅ Fail fast by returning the error from `RunE` (not `panic`)
- ✅ Reserve `OnInitialize` for cheap, truly-every-command concerns only (or drop it entirely)
- ❌ Never set `zap.ReplaceGlobals` / `config.ENV` — pass logger + config as constructor deps
- 🔗 See [dependency-injection.md](dependency-injection.md) → *Composition Root* & *Config Injection*

---

## Composition Root in `RunE`

**`RunE` is where the app is built, run, and torn down** — it owns the composition root lifecycle:

```go
RunE: func(cmd *cobra.Command, _ []string) error {
    cfgFile, _ := cmd.Flags().GetString("config")
    port, _ := cmd.Flags().GetInt64("port")
    shutdownTime, _ := cmd.Flags().GetInt64("shutdown_time")

    log := bootstrap.NewLogger()                  // construct logger (no global)
    cfg, err := config.LoadConfig(cfgFile, log)   // load config from --config (no global)
    if err != nil {
        return err
    }

    // 1. Build app (composition root)
    app, cleanup, err := bootstrap.NewApp(cfg, log)
    if err != nil {
        return err
    }
    defer cleanup()                               // teardown resources

    // 2. Signal context for graceful shutdown
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // 3. Run with graceful shutdown
    return govern.Run(ctx, log, time.Duration(shutdownTime)*time.Second, app.Server)
},
```

**Rules:**
- ✅ `RunE` builds → `defer cleanup()` → runs → returns
- ✅ Defer cleanup **immediately** after a successful build
- ✅ Signal handling (`signal.NotifyContext`) lives in `RunE`, not the app
- 🔗 See [dependency-injection.md](dependency-injection.md) for the `(T, cleanup, error)` triple

---

## Argument Validation

**Use Cobra's built-in validators instead of manual checks:**

```go
var migrateCmd = &cobra.Command{
    Use:   "migrate [direction]",
    Short: "Run database migrations",
    Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
    ValidArgs: []string{"up", "down"},
    RunE: func(cmd *cobra.Command, args []string) error {
        return runMigration(args[0])
    },
}
```

**Rules:**
- ✅ `cobra.ExactArgs(n)`, `cobra.NoArgs`, `cobra.MaximumNArgs(n)` for arity
- ✅ `ValidArgs` + `cobra.OnlyValidArgs` for enum-style positional args
- ✅ Custom validator: `Args: func(cmd *cobra.Command, args []string) error { ... }`
- ❌ Never validate args inside `RunE` when a built-in validator exists

---

## Best Practices & Pitfalls

**✅ DO:**
- Name command files after the command (`cmd/serverd.go`)
- Keep `RunE` focused on orchestration; delegate work to the app
- Return errors from `RunE`; reserve `os.Exit` for `Execute()`
- Defer cleanup right after building the app

**❌ DON'T:**
- Put `os.Exit()` in `RunE` (skips defers, breaks teardown)
- Use `Run` when you need to surface errors
- Read globals deep in `cmd/` beyond the composition root
- Build the app in `OnInitialize` (cleanup won't run)
- Register commands in a central `init()` (use per-file registration)

**Pitfalls:**
```go
// BAD — os.Exit in RunE skips defer cleanup()
RunE: func(cmd *cobra.Command, _ []string) error {
    app, cleanup, _ := bootstrap.NewApp(cfg, log)
    defer cleanup()
    if err := app.Run(); err != nil {
        os.Exit(1)                       // cleanup() NEVER runs → leaked DB pool
    }
    return nil
}

// BAD — using Run instead of RunE
Run: func(cmd *cobra.Command, args []string) { ... } // can't return error

// BAD — central registration in root.go
func init() {
    rootCmd.AddCommand(serverCmd, migrateCmd, workerCmd) // scales poorly; easy to forget
}

// BAD — building app in OnInitialize
cobra.OnInitialize(func() { app, _, _ = bootstrap.NewApp(...) }) // no cleanup path, runs for --help
```

---

## Quick Reference

```go
// main.go — thin
func main() { defer recoverAndFlush(); cmd.Execute() }

// cmd/root.go — root + Execute + flag registration only
var rootCmd = &cobra.Command{Use: "app", Short: "..."}
func Execute() { if err := rootCmd.Execute(); err != nil { os.Exit(1) } }
func init() { rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".env", "") } // logger/config built in RunE

// cmd/<name>.go — one subcommand, self-registering
var fooCmd = &cobra.Command{Use: "foo", RunE: func(cmd *cobra.Command, _ []string) error { ... }}
func init() { rootCmd.AddCommand(fooCmd); fooCmd.Flags().Int("port", 8080, "") }
```

| Concern | Rule |
|---------|------|
| `main()` | recovery + `cmd.Execute()` only |
| Exit codes | `os.Exit` only in `Execute()` |
| Error surfacing | `RunE` returns `error` |
| Cross-cmd flags | `PersistentFlags` |
| Local flags | `Flags()` |
| Logger + config | build/load in composition root, not `OnInitialize` globals |
| App lifecycle | build + `defer cleanup()` + run, inside `RunE` |
| Registration | per-file `init()` → `rootCmd.AddCommand` |

---

## References

- [spf13/cobra](https://github.com/spf13/cobra) — CLI framework
- [dependency-injection.md](dependency-injection.md) — composition root & `(T, cleanup, error)`
- [infrastructure-rules.md](infrastructure-rules.md) — config loading & logger setup
- Project entry points: `main.go`, `cmd/root.go`, `cmd/serverd.go` (paths are project-specific)
