# `argbin` Direct Implementation Plan

Scope: this document is written against the current [`argbin/argbin.go`](argbin/argbin.go). Production code was not changed.

Goal: keep the current public shape mostly intact and finish the parser with direct code that can be copied into `argbin.go` with minimal adaptation.

## Current production shape

Current types already include the fields needed for the next parser pass:

```go
type Flag struct {
    Name        string // logical name: "verbose"
    Short       string // optional short hand: "v"
    Description string
    Execute     func(ctx *Context) error
    TakesValue  bool
}

type Context struct {
    Command *Command
    Values  map[string]any
    Args    []string
}
```

Current gaps:

- `Short` is not used.
- `TakesValue` is not used.
- `Args` is not populated.
- CLI spellings `--verbose` and `-v` are not normalized.
- duplicate flags are not detected.
- parsing still executes flag callbacks before all tokens are known valid.

---

## 1. Add missing parser sentinels

In the existing error block, keep the current errors and add these:

```go
var (
    ErrMissingCommand   = errors.New("no arguments provided")
    ErrUnknownCommand   = errors.New("unknown command")
    ErrNilRoot          = errors.New("root cannot be nil")
    ErrNilCommand       = errors.New("command cannot be nil")
    ErrNilFlag          = errors.New("flag cannot be nil")
    ErrNilContext       = errors.New("context cannot be nil")
    ErrUnknownFlag      = errors.New("unknown flag")
    ErrDuplicateFlag    = errors.New("duplicate flag")
    ErrMissingFlagValue = errors.New("missing flag value")
)
```

Why:

- `ErrUnknownFlag` already exists.
- `ErrDuplicateFlag` lets callers detect repeated options.
- `ErrMissingFlagValue` lets callers detect `--output` without a value.

---

## 2. Add `strings` to imports

The parser needs prefix checks and trimming:

```go
import (
    "errors"
    "fmt"
    "strings"

    "github.com/Tyy47/clibox/internal/utils"
)
```

---

## 3. Add direct lookup helpers

Keep `searchFlag` for logical-name lookup, but add a short lookup helper.

Replace/keep `searchFlag` as:

```go
func (c *Command) searchFlag(name string) (*Flag, bool) {
    if c == nil {
        return nil, false
    }

    flag, ok := c.Flags[name]
    return flag, ok
}
```

Add this next to it:

```go
func (c *Command) searchShortFlag(short string) (*Flag, bool) {
    if c == nil {
        return nil, false
    }

    for _, flag := range c.Flags {
        if flag == nil {
            continue
        }
        if flag.Short == short {
            return flag, true
        }
    }

    return nil, false
}
```

This is still an O(n) scan, but it is deterministic if validation rejects duplicate `Short` values. A later optimization can build an index.

---

## 4. Add `Short` validation and duplicate-short detection

`validFlagChecker` should validate a single flag's `Short` format:

```go
func validFlagChecker(flag *Flag) error {
    if flag == nil {
        return ErrNilFlag
    }

    if flag.Name == "" {
        return fmt.Errorf("Flag name member can't be empty: %v", flag)
    }

    if flag.Short != "" {
        if len(flag.Short) != 1 {
            return fmt.Errorf("Flag short member must be one character: %v", flag)
        }
        if strings.Contains(flag.Short, "-") {
            return fmt.Errorf("Flag short member cannot contain '-': %v", flag)
        }
    }

    if flag.Description == "" {
        return fmt.Errorf("Flag description member can't be empty: %v", flag)
    }

    if flag.Execute == nil {
        return fmt.Errorf("Flag execute member can't be empty: %v", flag)
    }

    return nil
}
```

Then update `Command.validate` to reject duplicate shorts:

```go
func (c *Command) validate() error {
    if err := validCommandChecker(c); err != nil {
        return err
    }

    shorts := make(map[string]string)

    for key, flag := range c.Flags {
        if err := validFlagChecker(flag); err != nil {
            return fmt.Errorf("flag %q: %w", key, err)
        }
        if key != flag.Name {
            return fmt.Errorf("flag key %q does not match flag name %q", key, flag.Name)
        }
        if flag.Short != "" {
            if existing, ok := shorts[flag.Short]; ok {
                return fmt.Errorf("flag short %q is used by both %q and %q", flag.Short, existing, flag.Name)
            }
            shorts[flag.Short] = flag.Name
        }
    }

    return nil
}
```

---

## 5. Add parsed input structs

Place these near `Context` or near parsing methods:

```go
type parsedFlag struct {
    flag  *Flag
    value *string
}

type parsedInput struct {
    flags []parsedFlag
    args  []string
}
```

Meaning:

- `value == nil` means a presence flag.
- `value != nil` means a `TakesValue` flag supplied a string.
- `args` stores positional arguments for `Context.Args`.

---

## 6. Replace immediate callback parsing with side-effect-free parsing

Add this method below `searchShortFlag`.

```go
func (c *Command) parseInput(args []string) (*parsedInput, error) {
    if c == nil {
        return nil, ErrNilCommand
    }

    parsed := &parsedInput{}
    seen := make(map[string]struct{})

    for i := 0; i < len(args); i++ {
        token := args[i]

        if token == "--" {
            parsed.args = append(parsed.args, args[i+1:]...)
            break
        }

        if token == "" {
            parsed.args = append(parsed.args, token)
            continue
        }

        if token == "-" || !strings.HasPrefix(token, "-") {
            parsed.args = append(parsed.args, token)
            continue
        }

        if strings.Contains(token, "=") {
            return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
        }

        flag, err := c.resolveFlagToken(token)
        if err != nil {
            return nil, err
        }

        if _, ok := seen[flag.Name]; ok {
            return nil, fmt.Errorf("%w: %s", ErrDuplicateFlag, flag.Name)
        }
        seen[flag.Name] = struct{}{}

        if !flag.TakesValue {
            parsed.flags = append(parsed.flags, parsedFlag{flag: flag})
            continue
        }

        if i+1 >= len(args) || args[i+1] == "--" || strings.HasPrefix(args[i+1], "-") {
            return nil, fmt.Errorf("%w: %s", ErrMissingFlagValue, flag.Name)
        }

        i++
        value := args[i]
        parsed.flags = append(parsed.flags, parsedFlag{flag: flag, value: &value})
    }

    return parsed, nil
}
```

Notes:

- This intentionally rejects `--output=file.txt` because the desired grammar is no equals form.
- This rejects `--output -1` as missing because `-1` starts with `-`. If negative values should be valid, loosen that condition later.
- No flag callback runs in this method.

---

## 7. Add flag-token resolution

Add this helper used by `parseInput`:

```go
func (c *Command) resolveFlagToken(token string) (*Flag, error) {
    if strings.HasPrefix(token, "--") {
        name := strings.TrimPrefix(token, "--")
        if name == "" {
            return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
        }

        flag, ok := c.searchFlag(name)
        if !ok {
            return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
        }
        return flag, nil
    }

    if strings.HasPrefix(token, "-") {
        short := strings.TrimPrefix(token, "-")
        if len(short) != 1 {
            return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
        }

        flag, ok := c.searchShortFlag(short)
        if !ok {
            return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
        }
        return flag, nil
    }

    return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
}
```

Examples this supports:

```text
--verbose  -> Flag{Name: "verbose"}
-v         -> Flag{Short: "v"}
--output   -> Flag{Name: "output", TakesValue: true}
-o         -> Flag{Short: "o", TakesValue: true}
```

---

## 8. Add apply step for parsed input

Add this method:

```go
func (c *Command) applyParsed(ctx *Context, parsed *parsedInput) error {
    if c == nil {
        return ErrNilCommand
    }
    if ctx == nil {
        return ErrNilContext
    }
    if parsed == nil {
        return nil
    }
    if ctx.Values == nil {
        ctx.Values = make(map[string]any)
    }

    ctx.Args = append(ctx.Args, parsed.args...)

    for _, occurrence := range parsed.flags {
        if occurrence.value == nil {
            ctx.Values[occurrence.flag.Name] = true
        } else {
            ctx.Values[occurrence.flag.Name] = *occurrence.value
        }

        if err := occurrence.flag.Execute(ctx); err != nil {
            return fmt.Errorf("flag %q cannot be executed: %w", occurrence.flag.Name, err)
        }
    }

    return nil
}
```

This is where callbacks run, after the whole input has already parsed successfully.

---

## 9. Replace `parseFlags` with wrapper around parse/apply

To preserve the existing method name and test shape, `parseFlags` can become:

```go
func (c *Command) parseFlags(ctx *Context, args []string) error {
    if c == nil {
        return ErrNilCommand
    }
    if ctx == nil {
        return ErrNilContext
    }

    parsed, err := c.parseInput(args)
    if err != nil {
        return err
    }

    return c.applyParsed(ctx, parsed)
}
```

This keeps `Run` mostly unchanged while improving parser correctness.

---

## 10. Keep `Run` mostly as-is

With the parser changes above, `Run` can stay nearly identical:

```go
func (r *Root) Run() error {
    if r == nil {
        return ErrNilRoot
    }

    if err := r.validate(); err != nil {
        return err
    }

    args := *utils.GetArgs()

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

No `RunArgs` is required if the preferred API is only `Run()`.

---

## 11. Example command after the parser update

This is what user code would look like with logical names and short aliases:

```go
root := &Root{
    AppName:     "app",
    Description: "example app",
    Commands: CommandList{
        "build": {
            Name:        "build",
            Description: "build the project",
            Flags: map[string]*Flag{
                "verbose": {
                    Name:        "verbose",
                    Short:       "v",
                    Description: "enable verbose output",
                    Execute: func(ctx *Context) error {
                        verbose, _ := ctx.Values["verbose"].(bool)
                        fmt.Println("verbose:", verbose)
                        return nil
                    },
                },
                "output": {
                    Name:        "output",
                    Short:       "o",
                    Description: "output path",
                    TakesValue:  true,
                    Execute: func(ctx *Context) error {
                        output, _ := ctx.Values["output"].(string)
                        fmt.Println("output:", output)
                        return nil
                    },
                },
            },
            Execute: func(ctx *Context) error {
                fmt.Println("positionals:", ctx.Args)
                return nil
            },
        },
    },
}
```

Supported invocation examples after implementation:

```text
app build --verbose
app build -v
app build --output dist/app
app build -o dist/app
app build input.go --verbose
app build -- input.go --not-a-flag
```

Rejected examples after implementation:

```text
app build --missing
app build -x
app build --output
app build --output=file
app build --verbose --verbose
```

---

## 12. Tests to update after parser implementation

After applying the production changes above, update tests around these expectations:

- `parseFlags(ctx, []string{"--verbose"})` executes `verbose`.
- `parseFlags(ctx, []string{"-v"})` executes `verbose`.
- `parseFlags(ctx, []string{"input.txt", "--verbose"})` stores `ctx.Args == []string{"input.txt"}`.
- `parseFlags(ctx, []string{"--output", "file.txt"})` stores `ctx.Values["output"] == "file.txt"`.
- `parseFlags(ctx, []string{"--output"})` wraps `ErrMissingFlagValue`.
- `parseFlags(ctx, []string{"--verbose", "--verbose"})` wraps `ErrDuplicateFlag`.
- If a later token is invalid, earlier flag callbacks do not execute.

---

## Recommended implementation order

1. Add `strings` import.
2. Add `ErrDuplicateFlag` and `ErrMissingFlagValue`.
3. Add `searchShortFlag`.
4. Add `Short` validation and duplicate-short validation.
5. Add `parsedFlag` and `parsedInput`.
6. Add `resolveFlagToken`.
7. Add `parseInput`.
8. Add `applyParsed`.
9. Replace `parseFlags` with the parse/apply wrapper.
10. Update tests to assert real CLI spellings and value parsing.
