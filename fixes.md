# `argbin` Functional Fixes

This review compares the current [`argbin/argbin.go`](argbin/argbin.go) with the previous recommendations. It intentionally ignores the existing tests until the package API and parsing behavior are complete.

No production code was changed during this review.

## Current status

| Area | Status | Notes |
|---|---|---|
| Command validation in `AddCommand` | Complete | `AddCommand` validates before insertion. |
| Command error wrapping | Complete | `Run` uses lowercase error text and `%w`. |
| Missing command handling | Complete | Empty input returns `ErrMissingCommand`. |
| Unknown command handling | Incomplete | `ErrUnknownCommand` exists but is never returned. |
| One-command execution | Partially complete | `Run` returns after execution, but searches for a command anywhere in the arguments. |
| Nil safety | Partially complete | Nil definitions and nil destination maps are handled in some paths; nil receivers and nil map entries can still panic. |
| Aliases | Removed | Alias collision handling is no longer needed. |
| Full configuration validation | Incomplete | Commands are validated, but unused flags are not. |
| Value-taking flags | Incomplete | `TakesValue` exists but parsing does not consume or expose a value. |

## 1. Define and enforce the command grammar - DONE

### Remaining problem

`Run` loops over every argument until one matches a command:

```go
for i := 0; i < len(args); i++ {
    command, ok := r.Commands[args[i]]
    // ...
    command.parseFlags(ctx, args[i+1:])
}
```

The early return prevents multiple commands from executing, but command boundaries remain ambiguous. Given:

```text
app first second --force
```

`first` is selected and scans `second --force` as its arguments. Input such as `app typo build` also silently selects `build` instead of reporting `typo` as an unknown command.

### Recommended fix

Use the conventional grammar:

```text
app <command> [command flags and arguments]
```

Only `args[0]` should be used for command lookup:

```go
func (r *Root) Run() error {
    args := *utils.GetArgs()

    if err := r.validate(); err != nil {
        return err
    }
    if len(args) == 0 {
        return ErrMissingCommand
    }

    command, ok := r.Commands[args[0]]
    if !ok {
        return fmt.Errorf("%w: %q", ErrUnknownCommand, args[0])
    }

    ctx := &Context{
        Command: command,
        Values:  make(map[string]any),
    }
    if err := command.parseFlags(ctx, args[1:]); err != nil {
        return err
    }
    if err := command.Execute(ctx); err != nil {
        return fmt.Errorf("command cannot be executed: %w", err)
    }
    return nil
}
```

### Code implications

- Removes the outer argument-search loop and its index bookkeeping.
- Guarantees at most one command execution.
- Makes every token after the command belong to that command.
- Turns a typo in the first token into an unknown-command error instead of searching later tokens.
- Establishes a grammar that can later be extended to nested subcommands explicitly.

## 2. Correct the sentinel errors - DONE

### Remaining problem

`ErrUnknownCommand` is declared but unused, and its message includes punctuation intended for a particular formatted error:

```go
ErrUnknownCommand = errors.New("unknown command: ")
```

Unknown non-empty input currently falls through and returns `ErrMissingCommand`, which reports the wrong condition. `NilCommand` and `NilFlag` also do not follow the conventional exported sentinel naming pattern.

### Recommended fix

Keep sentinel text stable and free of trailing formatting:

```go
var (
    ErrMissingCommand = errors.New("missing command")
    ErrUnknownCommand = errors.New("unknown command")
    ErrNilCommand     = errors.New("command cannot be nil")
    ErrNilFlag        = errors.New("flag cannot be nil")
    ErrUnknownFlag    = errors.New("unknown flag")
    ErrMissingValue   = errors.New("flag value is missing")
)
```

Add details by wrapping the sentinel:

```go
return fmt.Errorf("%w: %q", ErrUnknownCommand, args[0])
```

### Code implications

- Callers can distinguish failures with `errors.Is` without parsing strings.
- Renaming `NilCommand` and `NilFlag` to `ErrNilCommand` and `ErrNilFlag` is a public API change.
- Removing the trailing colon changes the direct error text but produces cleaner wrapped messages.
- `ErrUnknownFlag` and `ErrMissingValue` provide stable categories needed by flag parsing.

## 3. Finish nil safety

### What is already fixed

- `validCommandChecker` rejects nil commands.
- `validFlagChecker` rejects nil flags.
- `AddCommand` initializes a nil command map.
- `AddFlag` initializes a nil flag map.

### Remaining problem

These paths can still panic:

- Calling `AddCommand` on a nil `*Root`. - Fixed
- Calling `AddFlag`, `searchFlag`, or `parseFlags` on a nil `*Command`. - Fixed
- `searchFlag` encountering a nil flag stored in `c.Flags`, because it reads `flag.Name`. - Fixed
- Calling `Run` on a nil `*Root`. - Fixed
- A flag implementation writing to a nil `Context.Values` map when `parseFlags` is called outside `Run`. - Fixed

### Recommended fix

Guard public and parser entry points before dereferencing receivers:

```go
if r == nil {
    return errors.New("root cannot be nil")
}
```

```go
if c == nil {
    return errors.New("command cannot be nil")
}
```

Lookup must reject or skip nil stored values:

```go
func (c *Command) searchFlag(name string) (*Flag, bool) {
    if c == nil {
        return nil, false
    }
    flag, ok := c.Flags[name]
    return flag, ok && flag != nil
}
```

Ensure parser-created context state is initialized. If `parseFlags` remains callable internally with an existing context, validate `ctx != nil` and initialize `ctx.Values` when nil.

### Code implications

- Zero-value `Root` and `Command` instances remain usable through `AddCommand` and `AddFlag`.
- Nil receivers return errors instead of process-ending panics.
- Directly inserted nil map values are caught deterministically.
- Direct map lookup becomes possible once map keys are required to equal definition names.

## 4. Enforce map-key invariants - Fixed: added to root validation function that gets ran at the start of Run()

### Remaining problem

Both `Commands` and `Flags` are maps keyed by name, but definitions also contain a `Name` field. External callers can construct inconsistent values:

```go
root.Commands["build"] = &Command{Name: "compile", ...}
command.Flags["--force"] = &Flag{Name: "--verbose", ...}
```

`Run` looks up commands by map key, while `searchFlag` accepts either the key or `flag.Name`. This creates implicit aliases and can make behavior depend on malformed configuration.

### Recommended fix

Choose one invariant:

```text
map key == object's Name
```

Validate it for directly constructed maps:

```go
for key, command := range r.Commands {
    if command == nil {
        return ErrNilCommand
    }
    if key != command.Name {
        return fmt.Errorf("command key %q does not match command name %q", key, command.Name)
    }
}
```

Apply the same rule to flags. Once validated, simplify flag lookup to:

```go
flag, ok := c.Flags[name]
```

### Code implications

- Eliminates accidental aliases after aliases were intentionally removed.
- Makes lookup constant-time without scanning the map.
- Makes manually assembled maps behave the same as maps populated through `AddCommand` and `AddFlag`.
- Rejecting previously tolerated mismatched keys is a behavior change.
- A stronger future alternative is making maps unexported and exposing read-only/accessor APIs, but that would be a larger breaking API change.

## 5. Validate the complete command tree before parsing

### Remaining problem

`Run` validates every command but does not validate each command's flags. `parseFlags` validates only flags that appear in input. An invalid unused flag therefore goes unnoticed, and a nil flag may panic during lookup.

### Recommended fix

Give each layer a focused validator:

```go
func (c *Command) validate() error {
    if err := validCommandChecker(c); err != nil {
        return err
    }

    for key, flag := range c.Flags {
        if err := validFlagChecker(flag); err != nil {
            return fmt.Errorf("flag %q: %w", key, err)
        }
        if key != flag.Name {
            return fmt.Errorf("flag key %q does not match flag name %q", key, flag.Name)
        }
    }
    return nil
}
```

```go
func (r *Root) validate() error {
    if r == nil {
        return errors.New("root cannot be nil")
    }

    for key, command := range r.Commands {
        if err := command.validate(); err != nil {
            return fmt.Errorf("command %q: %w", key, err)
        }
        if key != command.Name {
            return fmt.Errorf("command key %q does not match command name %q", key, command.Name)
        }
    }
    return nil
}
```

Call `r.validate()` once at the beginning of `Run`. After complete validation, remove `validFlagChecker(flag)` and the unreachable `flag.Execute == nil` branch from `parseFlags`.

### Code implications

- Configuration errors are reported before any flag or command side effects occur.
- Invalid flags are rejected whether or not users supply them.
- Validation logic moves out of parsing, making `parseFlags` smaller.
- Exported mutable fields remain supportable because `Run` revalidates the tree.
- Wrapping errors with `%w` preserves sentinel classification through command/flag context.

## 6. Complete value-taking flag behavior

### Remaining problem

`Flag.TakesValue` has been added, but it is currently unused. `parseFlags` uses `range`, cannot consume the token after a flag, and does not place a value in `Context`. `Flag.Execute` also receives no explicit value.

### Recommended minimal, backward-compatible design

Keep the current execute signature and place parsed values in `Context.Values` under the canonical flag name:

```go
func (c *Command) parseFlags(ctx *Context, args []string) error {
    for i := 0; i < len(args); i++ {
        name, inlineValue, hasInlineValue := splitFlag(args[i])

        flag, ok := c.searchFlag(name)
        if !ok {
            return fmt.Errorf("%w: %q", ErrUnknownFlag, name)
        }

        if flag.TakesValue {
            switch {
            case hasInlineValue:
                ctx.Values[flag.Name] = inlineValue
            case i+1 < len(args):
                i++
                ctx.Values[flag.Name] = args[i]
            default:
                return fmt.Errorf("%w: %s", ErrMissingValue, flag.Name)
            }
        } else {
            if hasInlineValue {
                return fmt.Errorf("flag %s does not take a value", flag.Name)
            }
            ctx.Values[flag.Name] = true
        }

        if err := flag.Execute(ctx); err != nil {
            return fmt.Errorf("flag %s cannot be executed: %w", flag.Name, err)
        }
    }
    return nil
}
```

`splitFlag` can support `--output=value` with `strings.Cut`. The exact accepted syntax should be documented before implementation.

### Alternative API design

Pass the value directly:

```go
type Flag struct {
    Name        string
    Description string
    TakesValue  bool
    Execute     func(ctx *Context, value string) error
}
```

This is more explicit, but changing `Execute` is a breaking API change. A `*string` value can distinguish “no value” from an explicitly empty value if that distinction matters.

### Code implications

- `parseFlags` must use an index loop so it can consume values.
- A policy is required for `--name=value`, values beginning with `-`, repeated flags, and empty values.
- With the minimal design, existing presence-flag callbacks keep compiling and read their state from `ctx.Values`.
- Canonical names should be used as context keys so map keys and spelling variants do not create duplicate state.
- Repeated value flags need a defined policy: last value wins, first wins, append to a slice, or error.

## 7. Decide how unknown tokens and positional arguments work

### Remaining problem

`parseFlags` currently ignores every token that is not a known flag. Once value consumption is added, silently ignoring tokens makes it impossible to distinguish typos, unsupported positional arguments, and misplaced values.

### Recommended fix

Add an explicit policy. For the simplest initial API, reject every unrecognized token:

```go
flag, ok := c.searchFlag(name)
if !ok {
    return fmt.Errorf("%w: %q", ErrUnknownFlag, name)
}
```

If commands need positional arguments, model them separately in `Context`:

```go
type Context struct {
    Command    *Command
    Values     map[string]any
    Positionals []string
}
```

A conventional `--` token can end flag parsing and append all remaining tokens to `Positionals`.

### Code implications

- Rejecting unknown tokens improves typo detection but changes current permissive behavior.
- Supporting positionals requires defining whether they may appear before, between, or only after flags.
- Supporting `--` avoids treating positional values beginning with `-` as flags.
- This decision must be made together with value parsing to avoid ambiguous token consumption.

## 8. Separate argument acquisition from parsing

### Remaining problem

`Run` reads global process state directly through `utils.GetArgs()`. This couples parsing to `os.Args` and prevents callers from parsing an explicit argument slice for embedded or repeated use.

### Recommended fix

Keep `Run` as a convenience wrapper and put behavior in `RunArgs`:

```go
func (r *Root) Run() error {
    return r.RunArgs(*utils.GetArgs())
}

func (r *Root) RunArgs(args []string) error {
    // validate, select command, parse flags, execute
}
```

An alternative is a parsing-only method that returns a context, followed by a separate execution method. That offers more control but increases API surface.

### Code implications

- Core parsing no longer depends on global process state.
- Applications can invoke a root repeatedly with different argument slices.
- `Run` remains source-compatible for normal CLI use.
- A parsing/execution split would allow callers to inspect parsed state before side effects, but requires a clear lifecycle API.

## Recommended implementation order

1. Enforce the one-command grammar and return the correct unknown-command error.
2. Standardize sentinel errors and finish nil-receiver/map-entry handling.
3. Enforce command and flag map-key invariants.
4. Add full-tree validation and simplify `parseFlags` afterward.
5. Define unknown-token and positional-argument behavior.
6. Implement value consumption for `TakesValue`.
7. Extract `RunArgs` once parsing behavior is stable.
