package argbin

import (
	"errors"
	"os"
	"testing"
)

func withArgs(t *testing.T, args ...string) {
	t.Helper()
	oldArgs := os.Args
	os.Args = append([]string{"clibox-test"}, args...)
	t.Cleanup(func() { os.Args = oldArgs })
}

func TestContextValidate(t *testing.T) {
	t.Run("nil context", func(t *testing.T) {
		var ctx *Context
		if err := ctx.Validate(); !errors.Is(err, ErrNilContext) {
			t.Fatalf("Validate() error = %v, want %v", err, ErrNilContext)
		}
	})

	t.Run("nil command", func(t *testing.T) {
		ctx := &Context{}
		if err := ctx.Validate(); !errors.Is(err, ErrNilCommand) {
			t.Fatalf("Validate() error = %v, want %v", err, ErrNilCommand)
		}
	})

	t.Run("initializes nil values map", func(t *testing.T) {
		ctx := &Context{Command: &Command{Name: "run"}}
		if err := ctx.Validate(); err != nil {
			t.Fatalf("Validate() unexpected error: %v", err)
		}
		if ctx.Values == nil {
			t.Fatal("Validate() did not initialize Values")
		}
	})
}

func TestContextGetValue(t *testing.T) {
	ctx := &Context{Values: map[string]any{"enabled": true}}

	got, err := ctx.GetValue("enabled")
	if err != nil {
		t.Fatalf("GetValue(existing) unexpected error: %v", err)
	}
	if got != true {
		t.Fatalf("GetValue(existing) = %v, want true", got)
	}

	got, err = ctx.GetValue("missing")
	if err == nil {
		t.Fatalf("GetValue(missing) error = nil, want an error (got value %v)", got)
	}
	if got != nil {
		t.Fatalf("GetValue(missing) value = %v, want nil", got)
	}
}

func TestContextToggleValueSetsRequestedKey(t *testing.T) {
	ctx := &Context{Values: map[string]any{"verbose": false}}

	if err := ctx.ToggleValue("verbose", true); err != nil {
		t.Fatalf("ToggleValue() unexpected error: %v", err)
	}
	if got := ctx.Values["verbose"]; got != true {
		t.Fatalf("ToggleValue() set verbose = %v, want true", got)
	}
}

func TestRootRun(t *testing.T) {
	t.Run("missing arguments", func(t *testing.T) {
		withArgs(t)
		root := &Root{AppName: "app", AppVersion: "1.0.0", CommandList: []*Command{{Name: "run", Execute: func(*Context) error { return nil }}}}

		if err := root.Run(); !errors.Is(err, ErrMissingArguments) {
			t.Fatalf("Run() error = %v, want %v", err, ErrMissingArguments)
		}
	})

	t.Run("executes matching command and flag", func(t *testing.T) {
		withArgs(t, "run", "--verbose")
		var executed bool
		root := &Root{
			AppName:    "app",
			AppVersion: "1.0.0",
			CommandList: []*Command{{
				Name:  "run",
				Flags: Flags{"--verbose": func(ctx *Context) error {
					ctx.Values["verbose"] = true
					return nil
				}},
				Execute: func(ctx *Context) error {
					executed = true
					if ctx.Command == nil || ctx.Command.Name != "run" {
						t.Fatalf("Execute() ctx.Command = %#v, want run command", ctx.Command)
					}
					if ctx.Values["verbose"] != true {
						t.Fatalf("Execute() verbose = %v, want true", ctx.Values["verbose"])
					}
					return nil
				},
			}},
		}

		if err := root.Run(); err != nil {
			t.Fatalf("Run() unexpected error: %v", err)
		}
		if !executed {
			t.Fatal("Run() did not execute command")
		}
	})

	t.Run("unknown command returns error without executing", func(t *testing.T) {
		withArgs(t, "missing")
		var executed bool
		root := &Root{AppName: "app", AppVersion: "1.0.0", CommandList: []*Command{{Name: "run", Execute: func(*Context) error {
			executed = true
			return nil
		}}}}

		if err := root.Run(); !errors.Is(err, ErrNoCommandFound) {
			t.Fatalf("Run() error = %v, want %v", err, ErrNoCommandFound)
		}
		if executed {
			t.Fatal("Run() executed command despite unknown input")
		}
	})

	t.Run("returns validation errors before executing", func(t *testing.T) {
		withArgs(t, "run")
		var executed bool
		root := &Root{AppVersion: "1.0.0", CommandList: []*Command{{Name: "run", Execute: func(*Context) error {
			executed = true
			return nil
		}}}}

		if err := root.Run(); !errors.Is(err, ErrEmptyRootName) {
			t.Fatalf("Run() error = %v, want %v", err, ErrEmptyRootName)
		}
		if executed {
			t.Fatal("Run() executed command despite invalid root")
		}
	})

	t.Run("nil root returns error instead of panicking", func(t *testing.T) {
		withArgs(t, "run")
		var root *Root
		defer func() {
			if rec := recover(); rec != nil {
				t.Fatalf("Run() panicked: %v", rec)
			}
		}()

		if err := root.Run(); !errors.Is(err, ErrNilRoot) {
			t.Fatalf("Run() error = %v, want %v", err, ErrNilRoot)
		}
	})
}
