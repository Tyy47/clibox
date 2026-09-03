# `argbin` Completion Plan

This document reviews the current [`argbin/argbin.go`](argbin/argbin.go), verifies previously marked fixes, and combines the remaining work into a single implementation plan. The goal is to settle the parser contract now so argument parsing does not need to be redesigned later.

No production code was changed during this review. Existing tests are intentionally out of scope until the API is functionally complete.

## Verified status

| Area | Status | Current behavior |
|---|---|---|
| One command per invocation | **Fixed** | `Run` treats `args[0]` as the command and executes at most one command. |
| Missing command | **Fixed** | Empty input returns `ErrMissingCommand`. |
| Unknown command | **Fixed** | An unknown `args[0]` wraps `ErrUnknownCommand`. |
| Command execution errors | **Fixed** | Errors use lowercase text and wrap the cause with `%w`. |
| Nil command/flag validators | **Fixed** | The validators reject nil definitions before field access. |
| Nil map initialization in add methods | **Fixed** | `AddCommand` and `AddFlag` initialize their maps. |
| Root command key/name validation | **Fixed** | `Root.validate` checks each command map key against `Command.Name`. |
| Nil receiver safety | **Broken by reversed conditions** | The `isItNil` helper is used backwards in three command methods and both add methods. |
| Flag key/name validation | **Not fixed** | Only command keys are checked by `Root.validate`. |
| Full command-tree validation | **Not fixed** | `Root.validate` does not call command or flag validation. |
| Value-taking flags | **Not implemented** | `TakesValue` is not read by the parser. |
| Unknown token policy | **Not defined** | Every unrecognized token after the command is silently ignored. |

---

# Part I: Correctness blockers

These issues should be resolved before extending flag syntax because they prevent otherwise valid code from running reliably.

## 1. Remove or correct `isItNil`

### Current defect

`isItNil` returns `true` when its argument is nil:

```go
func isItNil[T *Root | *Command | *Flag](item T) bool {
    return item == nil
}
```

The callers then negate it:

```go
if !isItNil(c) {
    return ErrNilCommand
}
```

As a result, a **non-nil** command is treated as nil. In the current code:

- `parseFlags` rejects every valid command.
- `searchFlag` refuses to search every valid command.
- `AddCommand` rejects every valid root.
- `AddFlag` rejects every valid command.

Conversely, a genuinely nil command can pass the condition and panic when its fields are accessed.

### Recommended fix

Do not use a generic helper for a one-token nil comparison. Direct checks are clearer and much harder to invert accidentally:

```go
if r == nil {
    return ErrNilRoot
}
```

```go
if c == nil {
    return ErrNilCommand
}
```

For lookup methods that cannot return an error, return a failed lookup on a nil receiver:

```go
if c == nil {
    return nil, false
}
```

### Code implications

- Remove `isItNil`; it provides no meaningful abstraction.
- Correct `AddCommand`, `AddFlag`, `parseFlags`, and `searchFlag` independently.
- Nil errors remain classifiable through the existing sentinels.
- Direct checks make static review straightforward and eliminate the double-negative logic.

## 2. Never create and discard errors

### Current defect

`searchFlag` contains calls such as:

```go
fmt.Errorf(ErrNilCommand.Error())
```

The resulting error is immediately discarded. This neither reports nor logs anything. It also uses a dynamic string as a format string unnecessarily.

### Recommended fix

`searchFlag` currently returns only `(*Flag, bool)`, so it should use `false` solely to mean “not found” and leave malformed-tree detection to validation:

```go
func (c *Command) searchFlag(name string) (*Flag, bool) {
    if c == nil {
        return nil, false
    }
    flag, ok := c.Flags[name]
    return flag, ok && flag != nil
}
```

If lookup itself must distinguish malformed data from absence, change its signature to `(*Flag, error)`. Do not manufacture an error that cannot reach the caller.

### Code implications

- Full-tree validation becomes responsible for reporting nil map entries.
- Lookup becomes a small constant-time operation.
- The parser no longer conflates invalid configuration with a missing user-supplied flag.

## 3. Guard nil `Context`

### Current defect

`parseFlags` initializes `ctx.Values`, but dereferences `ctx` first. A nil context still panics:

```go
if ctx.Values == nil {
    ctx.Values = make(map[string]any)
}
```

### Recommended fix

Add a sentinel such as `ErrNilContext` and check it before use:

```go
if ctx == nil {
    return ErrNilContext
}
if ctx.Values == nil {
    ctx.Values = make(map[string]any)
}
```

Alternatively, make context creation private and ensure only one constructor/path can call parsing. Even then, a defensive check is inexpensive.

### Code implications

- Internal parser mistakes return an error rather than panic.
- `parseFlags` remains safe if reused by a future `ParseArgs` or `RunArgs` API.

---

# Part II: Definition and validation model

The maps and exported fields allow callers to bypass `AddCommand` and `AddFlag`. Therefore, validation at insertion time is useful but not sufficient; every invocation must validate the complete definition tree before parsing.

## 4. Make validation recursive and side-effect free

### Current defect

`Root.validate` currently validates only:

- `AppName`.
- `Description`.
- A nil `Commands` map, which it mutates into an empty map.
- Nil command entries.
- Command map key/name equality.

It does **not** validate command names, command descriptions, command execute functions, flag definitions, or flag map key/name equality. A directly constructed root can therefore reach a nil `command.Execute` call or contain invalid unused flags.

Validation also should not mutate the object. Initializing `Commands` during `validate` hides whether the root has no command registry and makes a method named “validate” have side effects.

### Recommended structure

Use one validator per layer:

```go
func (r *Root) validate() error {
    if r == nil {
        return ErrNilRoot
    }
    if r.AppName == "" {
        return errors.New("root app name cannot be empty")
    }
    if r.Description == "" {
        return errors.New("root description cannot be empty")
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

A nil map is safe to range over and can represent an empty collection. Initialize maps in mutation methods or constructors, not validators.

### Code implications

- All configuration errors are found before any flag or command callback runs.
- Invalid unused flags can no longer remain hidden.
- `parseFlags` can remove its repeated `validFlagChecker` call.
- The `flag.Execute == nil` branch in `parseFlags` becomes unreachable and should be removed.
- Errors gain command/flag context while preserving wrapped sentinel causes.
- `validCommandChecker` can eventually become the implementation of `Command.validate`, avoiding two overlapping validation APIs.

## 5. Enforce one source of truth for names

### Current defect

Definitions have both a map key and a `Name` field. Command key/name equality is checked, but flag equality is not. `searchFlag` scans every flag and accepts either form:

```go
if key == name || flag.Name == name
```

This creates accidental aliases and turns a map lookup into an O(n) scan.

### Recommended invariant

Require:

```text
Root.Commands key == Command.Name
Command.Flags key == Flag.Name
```

After recursive validation enforces this, lookup is simply:

```go
flag, ok := c.Flags[name]
```

`AddCommand` and `AddFlag` already store definitions under their names, so this mainly affects callers who construct maps directly.

### Code implications

- Removes accidental alias behavior after aliases were intentionally removed.
- Makes lookup deterministic and constant-time.
- Rejects malformed manually assembled maps early.
- Keeping both exported maps and exported names means runtime validation remains necessary.
- Making maps private would enforce the invariant structurally, but that is a larger public API change and is not required for the first complete implementation.

## 6. Normalize definition names

A permanent syntax needs permanent naming rules. Without them, definitions can include whitespace, `=`, `--`, or other tokens that the parser cannot split safely.

### Recommended rules

Use logical names in definitions, not CLI spellings:

```go
Flag{Name: "output"}  // CLI spelling is --output
Command{Name: "build"}
```

Recommended validation:

- Command names: start with an ASCII letter or digit; remaining characters may be letters, digits, or `-`.
- Flag names: start with an ASCII letter; remaining characters may be letters, digits, or `-`.
- Reject leading `-`, whitespace, `=`, and empty segments.
- Reserve `help` and `version` only if the framework will provide them automatically.
- Keep names case-sensitive and recommend lowercase.

If retaining current literal flag names such as `"--output"`, enforce exactly two leading dashes and reject additional `=` or whitespace. Logical names are preferable because internal keys should describe data rather than terminal spelling.

### Code implications

- Moving from literal `"--output"` names to logical `"output"` names is a breaking definition/API change.
- Parsing must remove the `--` prefix before lookup.
- `Context.Values` gets stable keys such as `"output"` rather than presentation strings.
- Future rendering of help text can add `--` itself consistently.

---

# Part III: Final argument syntax

The following grammar is recommended as the stable first version. It is intentionally conventional, deterministic, and limited enough to implement once without accumulating ambiguous special cases.

## 7. Recommended CLI grammar

```text
invocation     = app command { argument }
argument       = long-flag | positional | end-of-flags
long-flag      = "--" flag-name [ "=" value ]
end-of-flags   = "--"
positional     = any token not interpreted as a flag
```

### Supported forms

Assuming `verbose` is a presence flag and `output` takes a required value:

| Input | Meaning |
|---|---|
| `app build --verbose` | Enable `verbose`. |
| `app build --output file.txt` | Set `output` to `file.txt`. |
| `app build --output=file.txt` | Set `output` to `file.txt`. |
| `app build source.go --verbose` | Positional arguments may be interspersed with flags. |
| `app build -- --verbose` | Treat `--verbose` as positional after the terminator. |
| `app build --output=-1` | A flag-looking or negative value is unambiguous in inline form. |
| `app build --output -1` | Consume `-1` as the value because it is not a long flag token. |

### Explicitly unsupported in version one

- Short flags such as `-v`.
- Grouped short flags such as `-abc`.
- Optional flag values.
- Abbreviated long names such as `--verb` for `--verbose`.
- Automatically negated flags such as `--no-color`.
- Multiple commands in one invocation.
- Nested subcommands.

These can be added later as explicit features without changing the long-flag grammar.

## 8. Exact token rules

### Command token

- The first argument is always the command.
- It is never searched for later in the slice.
- Missing input returns `ErrMissingCommand`.
- An unknown first token wraps `ErrUnknownCommand`.

### Presence flags

A presence flag has no value. Valid:

```text
--verbose
```

Invalid:

```text
--verbose=true
--verbose=
```

The parser records `Context.Values["verbose"] = true` and schedules its callback.

### Required-value flags

A required-value flag accepts either:

```text
--output file.txt
--output=file.txt
```

For separated syntax, use these deterministic rules:

1. If there is no next token, return `ErrMissingFlagValue`.
2. If the next token is exactly `--`, return `ErrMissingFlagValue`.
3. If the next token begins with `--`, return `ErrMissingFlagValue` rather than swallowing what is probably another flag.
4. Any other next token is consumed, including `-1` and `-filename`.
5. A value that genuinely begins with `--` must use inline syntax, for example `--output=--literal`.

Inline empty values should be accepted for value flags:

```text
--output=
```

This distinguishes an explicitly supplied empty string from a missing value.

### Unknown flags

Any unrecognized token beginning with `--` returns `ErrUnknownFlag`. Never silently ignore it.

Any token beginning with one `-`, except the lone token `-`, should return an unsupported/unknown flag error. This catches accidental short flags instead of treating them as files. A lone `-` remains positional for stdin/stdout conventions.

### Positional arguments

- Bare tokens are appended to `Context.Args` in their original order.
- Flags may appear before, after, or between positionals until `--` is seen.
- After `--`, every token is positional without interpretation.
- A command may validate positional count or meaning inside its execute callback. A later `Args` schema can be added independently of flag tokenization.

### Repeated flags

Use a strict default: reject a second occurrence with `ErrDuplicateFlag`.

This avoids executing side-effecting flag callbacks twice and avoids silently overwriting a value. If repetition is needed later, add an explicit `Repeatable bool` field and store repeated values as `[]string`. Do not make repetition implicit.

### Argument order

Preserve flag occurrence order for callbacks. Although definitions live in a map, execution order must come from the argument slice, never map iteration.

---

# Part IV: Parser data model and algorithm

## 9. Replace `TakesValue` with an explicit arity

A boolean works for two states but becomes unclear as features grow. Optional values are intentionally unsupported, but an enum documents the contract better:

```go
type FlagArity uint8

const (
    FlagNoValue FlagArity = iota
    FlagRequiredValue
)

type Flag struct {
    Name        string
    Description string
    Arity       FlagArity
    Execute     func(ctx *Context, value *string) error
}
```

`value == nil` means a presence flag. A non-nil pointer means a required value was supplied; it may point to an empty string.

### Code implications

- Replacing `TakesValue` is a breaking API change, but doing it while the package is unfinished avoids permanent ambiguity.
- Changing `Flag.Execute` is also breaking, but gives callbacks their own parsed value directly.
- `Context.Values` can still be populated centrally for commands that prefer reading all options at once.
- Invalid `FlagArity` values must be rejected during definition validation.

### Compatibility alternative

Keep `TakesValue` and the existing callback signature, then place values in `ctx.Values` before invoking the callback. This avoids a signature change but couples every flag callback to string map lookups and type assertions. For a “write once and forget it” API, the explicit callback value is preferable.

## 10. Expand `Context`

Recommended context shape:

```go
type Context struct {
    Command *Command
    Values  map[string]any
    Args    []string
}
```

For the initial parser:

- Presence flag: `Values[name] = true`.
- Required-value flag: `Values[name] = stringValue`.
- Positionals: append to `Args`.

If duplicate flags become supported, changing a value from `string` to `[]string` dynamically would make the map inconsistent. Prefer adding a dedicated accessor API before supporting repetition.

### Code implications

- Existing commands can continue sharing state through `Values`.
- Positionals no longer need to be silently discarded.
- Direct flag callback values avoid forcing callbacks to type-assert from `Values`.
- Typed accessors such as `ctx.String("output")` and `ctx.Bool("verbose")` can be added later without changing tokenization.

## 11. Parse completely before executing callbacks

### Current problem

The current `parseFlags` executes a flag immediately after recognizing it. If a later token is malformed, earlier callbacks have already caused side effects even though the command never executes.

### Recommended two-phase design

First parse into an internal representation without side effects:

```go
type flagOccurrence struct {
    flag  *Flag
    value *string
}

type parsedArgs struct {
    flags      []flagOccurrence
    positionals []string
}
```

Then populate context and execute callbacks only after the complete token slice is valid.

### Code implications

- Syntax errors produce no flag callback side effects.
- Definition validation, syntax parsing, and execution become separate responsibilities.
- Callback execution can still partially apply if a callback itself fails; rollback of arbitrary user side effects is impossible and should not be promised.
- Callback errors should wrap their cause with flag context:

```go
return fmt.Errorf("flag --%s cannot be applied: %w", occurrence.flag.Name, err)
```

## 12. Detailed parser algorithm

The parser should operate by index because value flags consume tokens:

```go
func (c *Command) parseArgs(args []string) (*parsedArgs, error) {
    result := &parsedArgs{}
    seen := make(map[string]struct{})

    for i := 0; i < len(args); i++ {
        token := args[i]

        if token == "--" {
            result.positionals = append(result.positionals, args[i+1:]...)
            break
        }

        if token == "-" || !strings.HasPrefix(token, "-") {
            result.positionals = append(result.positionals, token)
            continue
        }

        if !strings.HasPrefix(token, "--") {
            return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
        }

        spelling, inlineValue, hasInlineValue := strings.Cut(token[2:], "=")
        if spelling == "" {
            return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
        }

        flag, ok := c.Flags[spelling]
        if !ok {
            return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, "--"+spelling)
        }
        if _, duplicate := seen[flag.Name]; duplicate {
            return nil, fmt.Errorf("%w: --%s", ErrDuplicateFlag, flag.Name)
        }
        seen[flag.Name] = struct{}{}

        switch flag.Arity {
        case FlagNoValue:
            if hasInlineValue {
                return nil, fmt.Errorf("%w: --%s", ErrUnexpectedFlagValue, flag.Name)
            }
            result.flags = append(result.flags, flagOccurrence{flag: flag})

        case FlagRequiredValue:
            value := inlineValue
            if !hasInlineValue {
                if i+1 >= len(args) || args[i+1] == "--" || strings.HasPrefix(args[i+1], "--") {
                    return nil, fmt.Errorf("%w: --%s", ErrMissingFlagValue, flag.Name)
                }
                i++
                value = args[i]
            }
            result.flags = append(result.flags, flagOccurrence{flag: flag, value: &value})

        default:
            return nil, fmt.Errorf("flag --%s has invalid arity", flag.Name)
        }
    }

    return result, nil
}
```

This is design-level code. Before implementation, choose whether definitions use logical names (`"output"`) as recommended or literal names (`"--output"`), then keep that choice consistent throughout validation, maps, context keys, and help output.

## 13. Apply parsed arguments

After successful parsing:

```go
func applyParsed(ctx *Context, parsed *parsedArgs) error {
    ctx.Args = append(ctx.Args, parsed.positionals...)

    for _, occurrence := range parsed.flags {
        if occurrence.value == nil {
            ctx.Values[occurrence.flag.Name] = true
        } else {
            ctx.Values[occurrence.flag.Name] = *occurrence.value
        }

        if err := occurrence.flag.Execute(ctx, occurrence.value); err != nil {
            return fmt.Errorf(
                "flag --%s cannot be applied: %w",
                occurrence.flag.Name,
                err,
            )
        }
    }
    return nil
}
```

The command runs only after `applyParsed` succeeds.

### Code implications

- A syntactically invalid invocation never runs any callback.
- Flag callbacks run in user-supplied order.
- The command sees a fully populated context.
- If a callback fails, later callbacks and the command do not run.

---

# Part V: Public execution API

## 14. Separate argument acquisition from parsing

`Run` currently reads process-global arguments directly. Preserve it as a convenience wrapper, but make explicit argument execution the core API:

```go
func (r *Root) Run() error {
    return r.RunArgs(*utils.GetArgs())
}

func (r *Root) RunArgs(args []string) error {
    // validate tree
    // resolve args[0]
    // parse args[1:]
    // apply parsed flags
    // execute command
}
```

### Code implications

- Normal CLI callers still use `Run()`.
- Embedded callers can invoke the same root with explicit argument slices.
- The parser no longer depends internally on mutable global process state.
- Parsing can eventually be exposed separately if callers need inspection without execution.

## 15. Complete error taxonomy

Recommended stable sentinels:

```go
var (
    ErrMissingCommand      = errors.New("missing command")
    ErrUnknownCommand      = errors.New("unknown command")
    ErrUnknownFlag         = errors.New("unknown flag")
    ErrDuplicateFlag       = errors.New("duplicate flag")
    ErrMissingFlagValue    = errors.New("missing flag value")
    ErrUnexpectedFlagValue = errors.New("unexpected flag value")
    ErrNilRoot             = errors.New("root cannot be nil")
    ErrNilCommand          = errors.New("command cannot be nil")
    ErrNilFlag             = errors.New("flag cannot be nil")
    ErrNilContext          = errors.New("context cannot be nil")
)
```

Wrap these with token details rather than creating unique unclassifiable strings. Configuration errors that callers do not need to branch on may remain descriptive errors without dedicated sentinels.

---

# Recommended implementation sequence

1. Remove `isItNil` and fix all direct nil checks.
2. Remove discarded `fmt.Errorf` calls and make lookup a direct map operation.
3. Implement recursive, side-effect-free root/command/flag validation.
4. Decide and enforce logical flag names versus literal `--name` definitions; logical names are recommended.
5. Finalize `FlagArity`, the callback signature, and `Context.Args` while the API is still unfinished.
6. Add the parser sentinels and implement the exact grammar above in a side-effect-free `parseArgs` phase.
7. Implement `applyParsed`, then execute the command only after parsing and flag application succeed.
8. Extract `RunArgs`; keep `Run` as the process-argument wrapper.
9. Document unsupported syntax explicitly so later additions remain deliberate rather than accidental.
10. Rewrite tests against the finalized grammar and public API.

# Final design decisions to lock before coding

The plan above recommends the following answers:

- **Command location:** exactly `args[0]`.
- **Flag spelling:** long flags only, `--name`.
- **Definition names:** logical names without dashes.
- **Presence flags:** `--verbose`; assigned boolean `true`.
- **Value flags:** `--output value` and `--output=value`.
- **Optional flag values:** unsupported.
- **Empty value:** supported only explicitly through `--output=`.
- **Flag-looking values:** use inline form, such as `--output=--value`.
- **Short flags and aliases:** unsupported.
- **Unknown flags:** errors, never ignored.
- **Bare tokens:** positional arguments.
- **End-of-flags marker:** `--` sends all remaining tokens to positionals.
- **Repeated flags:** errors by default.
- **Parsing side effects:** none; callbacks run only after all syntax is valid.
- **Callback order:** argument occurrence order.
- **Command execution:** only after validation, parsing, and flag callbacks all succeed.

Locking these rules gives `parseFlags` one deterministic interpretation for every token and leaves clear extension points for future short flags, repetition, typed values, or nested subcommands without requiring another foundational rewrite.
