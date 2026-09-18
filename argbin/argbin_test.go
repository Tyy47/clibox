package argbin

import (
	"errors"
	"os"
	"testing"
)

func withArgs(t *testing.T, args ...string) {
	t.Helper()
	old := os.Args
	os.Args = append([]string{"argbin-test"}, args...)
	t.Cleanup(func() { os.Args = old })
}

func validRoot(command *Command) *Root {
	return &Root{
		AppName:     "test",
		HelpMenu: "test application",
		CommandList: []*Command{command},
	}
}

func TestContextValidate(t *testing.T) {
	var nilContext *Context
	if err := nilContext.Validate(); !errors.Is(err, ErrNilContext) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrNilContext)
	}

	ctx := &Context{}
	if err := ctx.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
	if ctx.Command == nil {
		t.Fatal("Validate() did not initialize Command")
	}
	if ctx.Values == nil {
		t.Fatal("Validate() did not initialize Values")
	}
}

func TestContextValues(t *testing.T) {
	ctx := &Context{Values: map[string]any{"enabled": "initial"}}

	if err := ctx.ToggleValue("enabled", true); err != nil {
		t.Fatalf("ToggleValue() error = %v, want nil", err)
	}
	got, err := ctx.GetValue("enabled")
	if err != nil {
		t.Fatalf("GetValue() error = %v, want nil", err)
	}
	if got != true {
		t.Fatalf("GetValue(\"enabled\") = %v, want true", got)
	}

	if _, err := ctx.GetValue("missing"); !errors.Is(err, ErrMissingContextValue) {
		t.Fatalf("GetValue() error = %v, want %v", err, ErrMissingContextValue)
	}
	if err := ctx.ToggleValue("missing", true); !errors.Is(err, ErrMissingContextValue) {
		t.Fatalf("ToggleValue() error = %v, want %v", err, ErrMissingContextValue)
	}
}

func TestContextGatherParsedValue(t *testing.T) {
	cmd := &Command{Name: "serve"}
	tests := []struct {
		name string
		args []string
		want string
		err  error
	}{
		{name: "immediate value", args: []string{"value"}, want: "value"},
		{name: "no value", err: ErrRequiresInput},
		{name: "flag is not a command value", args: []string{"--verbose", "value"}, err: ErrRequiresInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &Context{}
			err := ctx.gatherParsedValue(tt.args, cmd)
			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("gatherParsedValue() error = %v, want %v", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("gatherParsedValue() error = %v, want nil", err)
			}
			if ctx.ParsedValue != tt.want {
				t.Fatalf("ParsedValue = %q, want %q", ctx.ParsedValue, tt.want)
			}
		})
	}
}

func TestRootRunExecutesCommand(t *testing.T) {
	withArgs(t, "serve")

	called := false
	root := validRoot(&Command{Name: "serve", Execute: func(ctx *Context) error {
		called = true
		if ctx.Command.Name != "serve" {
			t.Errorf("Context.Command.Name = %q, want serve", ctx.Command.Name)
		}
		if len(ctx.Args) != 1 || ctx.Args[0] != "serve" {
			t.Errorf("Context.Args = %q, want [serve]", ctx.Args)
		}
		return nil
	}})

	if err := root.Run(); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if !called {
		t.Fatal("Run() did not execute the command")
	}
}

func TestRootRunTakesValue(t *testing.T) {
	t.Run("root command", func(t *testing.T) {
		withArgs(t, "serve", "value")
		root := validRoot(&Command{Name: "serve", TakesValue: true, Execute: func(ctx *Context) error {
			if ctx.ParsedValue != "value" {
				t.Errorf("ParsedValue = %q, want value", ctx.ParsedValue)
			}
			return nil
		}})

		if err := root.Run(); err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	})

	t.Run("nested subcommand", func(t *testing.T) {
		withArgs(t, "parent", "child", "value")
		child := &Command{Name: "child", TakesValue: true, Execute: func(ctx *Context) error {
			if ctx.ParsedValue != "value" {
				t.Errorf("ParsedValue = %q, want value", ctx.ParsedValue)
			}
			return nil
		}}
		root := validRoot(&Command{Name: "parent", Subcommands: []*Command{child}})

		if err := root.Run(); err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	})
}

func TestRootRunPropagatesCommandError(t *testing.T) {
	withArgs(t, "serve")
	want := errors.New("command failed")
	root := validRoot(&Command{Name: "serve", Execute: func(*Context) error { return want }})

	if err := root.Run(); !errors.Is(err, want) {
		t.Fatalf("Run() error = %v, want %v", err, want)
	}
}

func TestRootRunErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		root *Root
		want error
	}{
		{
			name: "missing arguments",
			root: validRoot(&Command{Name: "serve", Execute: func(*Context) error { return nil }}),
			want: ErrMissingArguments,
		},
		{
			name: "unknown command",
			args: []string{"unknown"},
			root: validRoot(&Command{Name: "serve", Execute: func(*Context) error { return nil }}),
			want: ErrUnknownCommand,
		},
		{
			name: "missing command value",
			args: []string{"serve"},
			root: validRoot(&Command{Name: "serve", TakesValue: true, Execute: func(*Context) error { return nil }}),
			want: ErrRequiresInput,
		},
		{
			name: "flag is not consumed as command value",
			args: []string{"serve", "--verbose"},
			root: validRoot(&Command{
				Name:       "serve",
				TakesValue: true,
				Flags:      Flags{"--verbose": {Execute: func(*Context) error { return nil }}},
				Execute:    func(*Context) error { return nil },
			}),
			want: ErrRequiresInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withArgs(t, tt.args...)
			if err := tt.root.Run(); !errors.Is(err, tt.want) {
				t.Fatalf("Run() error = %v, want %v", err, tt.want)
			}
		})
	}
}
