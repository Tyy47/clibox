package argbin

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestCommandSearchFlag(t *testing.T) {
	verbose := &Flag{Name: "verbose"}
	command := &Command{Flags: map[string]*Flag{"verbose": verbose}}

	got, ok := command.searchFlag("verbose")
	if !ok || got != verbose {
		t.Fatalf("searchFlag() = (%p, %t), want (%p, true)", got, ok, verbose)
	}

	if got, ok := command.searchFlag("--verbose"); ok || got != nil {
		t.Fatalf("searchFlag() with CLI spelling = (%v, %t), want (nil, false)", got, ok)
	}

	var nilCommand *Command
	if got, ok := nilCommand.searchFlag("verbose"); ok || got != nil {
		t.Fatalf("nil searchFlag() = (%v, %t), want (nil, false)", got, ok)
	}
}

func TestCommandSearchShortFlag(t *testing.T) {
	verbose := &Flag{Name: "verbose", Short: "v"}
	command := &Command{Flags: map[string]*Flag{"verbose": verbose}}

	got, ok := command.searchShortFlag("v")
	if !ok || got != verbose {
		t.Fatalf("searchShortFlag() = (%p, %t), want (%p, true)", got, ok, verbose)
	}

	if got, ok := command.searchShortFlag("x"); ok || got != nil {
		t.Fatalf("unknown searchShortFlag() = (%v, %t), want (nil, false)", got, ok)
	}

	var nilCommand *Command
	if got, ok := nilCommand.searchShortFlag("v"); ok || got != nil {
		t.Fatalf("nil searchShortFlag() = (%v, %t), want (nil, false)", got, ok)
	}
}

func TestCommandParseFlagsLongAndShortPresenceFlags(t *testing.T) {
	var executed []string
	command := &Command{Flags: map[string]*Flag{
		"verbose": {
			Name:        "verbose",
			Short:       "v",
			Description: "enable verbose output",
			Execute: func(ctx *Context) error {
				executed = append(executed, "verbose")
				return nil
			},
		},
		"force": {
			Name:        "force",
			Short:       "f",
			Description: "force action",
			Execute: func(ctx *Context) error {
				executed = append(executed, "force")
				return nil
			},
		},
	}}

	ctx := &Context{Command: command}
	if err := command.parseFlags(ctx, []string{"--verbose", "-f"}); err != nil {
		t.Fatalf("parseFlags() error = %v", err)
	}

	if got := strings.Join(executed, ","); got != "verbose,force" {
		t.Fatalf("execution order = %q, want verbose,force", got)
	}
	if got, ok := ctx.Values["verbose"].(bool); !ok || !got {
		t.Fatalf("ctx.Values[verbose] = %#v, want true", ctx.Values["verbose"])
	}
	if got, ok := ctx.Values["force"].(bool); !ok || !got {
		t.Fatalf("ctx.Values[force] = %#v, want true", ctx.Values["force"])
	}
}

func TestCommandParseFlagsCapturesPositionalsAndTerminator(t *testing.T) {
	command := &Command{Flags: map[string]*Flag{
		"verbose": {
			Name:        "verbose",
			Short:       "v",
			Description: "enable verbose output",
			Execute:     func(*Context) error { return nil },
		},
	}}

	ctx := &Context{Values: make(map[string]any)}
	args := []string{"input.txt", "--verbose", "--", "--literal", "tail"}
	if err := command.parseFlags(ctx, args); err != nil {
		t.Fatalf("parseFlags() error = %v", err)
	}

	wantArgs := []string{"input.txt", "--literal", "tail"}
	if !reflect.DeepEqual(ctx.Args, wantArgs) {
		t.Fatalf("ctx.Args = %#v, want %#v", ctx.Args, wantArgs)
	}
}

func TestCommandParseFlagsTakesValue(t *testing.T) {
	var callbackValue string
	command := &Command{Flags: map[string]*Flag{
		"output": {
			Name:        "output",
			Short:       "o",
			Description: "set output path",
			TakesValue:  true,
			Execute: func(ctx *Context) error {
				callbackValue, _ = ctx.Values["output"].(string)
				return nil
			},
		},
	}}

	ctx := &Context{Values: make(map[string]any)}
	if err := command.parseFlags(ctx, []string{"--output", "file.txt"}); err != nil {
		t.Fatalf("parseFlags() long value error = %v", err)
	}
	if got := ctx.Values["output"]; got != "file.txt" {
		t.Fatalf("ctx.Values[output] = %#v, want file.txt", got)
	}
	if callbackValue != "file.txt" {
		t.Fatalf("callback value = %q, want file.txt", callbackValue)
	}

	ctx = &Context{}
	callbackValue = ""
	if err := command.parseFlags(ctx, []string{"-o", "short.txt"}); err != nil {
		t.Fatalf("parseFlags() short value error = %v", err)
	}
	if got := ctx.Values["output"]; got != "short.txt" {
		t.Fatalf("ctx.Values[output] = %#v, want short.txt", got)
	}
	if callbackValue != "short.txt" {
		t.Fatalf("short callback value = %q, want short.txt", callbackValue)
	}
}

func TestCommandParseFlagsErrors(t *testing.T) {
	command := &Command{Flags: map[string]*Flag{
		"verbose": {
			Name:        "verbose",
			Short:       "v",
			Description: "enable verbose output",
			Execute:     func(*Context) error { return nil },
		},
		"output": {
			Name:        "output",
			Short:       "o",
			Description: "set output path",
			TakesValue:  true,
			Execute:     func(*Context) error { return nil },
		},
	}}

	tests := []struct {
		name string
		args []string
		want error
	}{
		{name: "unknown long", args: []string{"--missing"}, want: ErrUnknownFlag},
		{name: "unknown short", args: []string{"-x"}, want: ErrUnknownFlag},
		{name: "grouped short unsupported", args: []string{"-vf"}, want: ErrUnknownFlag},
		{name: "equals unsupported", args: []string{"--output=file.txt"}, want: ErrUnexpectedFlagValue},
		{name: "duplicate long", args: []string{"--verbose", "--verbose"}, want: ErrDuplicateFlag},
		{name: "duplicate long short", args: []string{"--verbose", "-v"}, want: ErrDuplicateFlag},
		{name: "missing value", args: []string{"--output"}, want: ErrMissingFlagValue},
		{name: "value is terminator", args: []string{"--output", "--"}, want: ErrMissingFlagValue},
		{name: "value looks like flag", args: []string{"--output", "-x"}, want: ErrMissingFlagValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := command.parseFlags(&Context{}, tt.args)
			if !errors.Is(err, tt.want) {
				t.Fatalf("parseFlags() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCommandParseFlagsParsesBeforeExecutingCallbacks(t *testing.T) {
	executed := false
	command := &Command{Flags: map[string]*Flag{
		"verbose": {
			Name:        "verbose",
			Description: "enable verbose output",
			Execute: func(*Context) error {
				executed = true
				return nil
			},
		},
	}}

	err := command.parseFlags(&Context{}, []string{"--verbose", "--missing"})
	if !errors.Is(err, ErrUnknownFlag) {
		t.Fatalf("parseFlags() error = %v, want ErrUnknownFlag", err)
	}
	if executed {
		t.Fatal("flag callback executed before the entire input was validated")
	}
}

func TestCommandParseFlagsNilGuards(t *testing.T) {
	var command *Command
	if err := command.parseFlags(&Context{}, nil); !errors.Is(err, ErrNilCommand) {
		t.Fatalf("nil command parseFlags() error = %v, want ErrNilCommand", err)
	}

	command = &Command{}
	if err := command.parseFlags(nil, nil); !errors.Is(err, ErrNilContext) {
		t.Fatalf("nil context parseFlags() error = %v, want ErrNilContext", err)
	}
}

func TestCommandParseFlagsWrapsCallbackError(t *testing.T) {
	wantErr := errors.New("flag failed")
	command := &Command{Flags: map[string]*Flag{
		"fail": {
			Name:        "fail",
			Description: "return an error",
			Execute:     func(*Context) error { return wantErr },
		},
	}}

	err := command.parseFlags(&Context{Values: make(map[string]any)}, []string{"--fail"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("parseFlags() error = %v, want wrapped %v", err, wantErr)
	}
	if err == nil || !strings.Contains(err.Error(), `flag "fail" cannot be executed`) {
		t.Fatalf("parseFlags() error = %v, want flag context", err)
	}
}

func TestValidCommandChecker(t *testing.T) {
	validExecute := func(*Context) error { return nil }
	tests := []struct {
		name     string
		commands []*Command
		wantErr  error
		wantText string
	}{
		{name: "empty list"},
		{name: "nil command", commands: []*Command{nil}, wantErr: ErrNilCommand},
		{name: "valid", commands: []*Command{{Name: "build", Description: "build it", Execute: validExecute}}},
		{name: "missing name", commands: []*Command{{Description: "build it", Execute: validExecute}}, wantText: "name"},
		{name: "leading dash", commands: []*Command{{Name: "-build", Description: "build it", Execute: validExecute}}, wantText: "name"},
		{name: "equals in name", commands: []*Command{{Name: "build=all", Description: "build it", Execute: validExecute}}, wantText: "name"},
		{name: "whitespace in name", commands: []*Command{{Name: "build all", Description: "build it", Execute: validExecute}}, wantText: "name"},
		{name: "internal hyphen allowed", commands: []*Command{{Name: "dry-run", Description: "dry run", Execute: validExecute}}},
		{name: "missing description", commands: []*Command{{Name: "build", Execute: validExecute}}, wantText: "description"},
		{name: "missing execute", commands: []*Command{{Name: "build", Description: "build it"}}, wantText: "execute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validCommandChecker(tt.commands...)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("validCommandChecker() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if tt.wantText == "" {
				if err != nil {
					t.Fatalf("validCommandChecker() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), tt.wantText) {
				t.Fatalf("validCommandChecker() error = %v, want error containing %q", err, tt.wantText)
			}
		})
	}
}

func TestValidFlagChecker(t *testing.T) {
	validExecute := func(*Context) error { return nil }
	tests := []struct {
		name     string
		flag     *Flag
		wantErr  error
		wantText string
	}{
		{name: "valid", flag: &Flag{Name: "force", Short: "f", Description: "force it", Execute: validExecute}},
		{name: "nil flag", flag: nil, wantErr: ErrNilFlag},
		{name: "missing name", flag: &Flag{Description: "force it", Execute: validExecute}, wantText: "name"},
		{name: "leading dash in name", flag: &Flag{Name: "--force", Description: "force it", Execute: validExecute}, wantText: "logical name"},
		{name: "equals in name", flag: &Flag{Name: "force=true", Description: "force it", Execute: validExecute}, wantText: "logical name"},
		{name: "whitespace in name", flag: &Flag{Name: "force now", Description: "force it", Execute: validExecute}, wantText: "logical name"},
		{name: "internal hyphen allowed", flag: &Flag{Name: "dry-run", Description: "dry run", Execute: validExecute}},
		{name: "short too long", flag: &Flag{Name: "force", Short: "ff", Description: "force it", Execute: validExecute}, wantText: "short"},
		{name: "short contains dash", flag: &Flag{Name: "force", Short: "-", Description: "force it", Execute: validExecute}, wantText: "short"},
		{name: "missing description", flag: &Flag{Name: "force", Execute: validExecute}, wantText: "description"},
		{name: "missing execute", flag: &Flag{Name: "force", Description: "force it"}, wantText: "execute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validFlagChecker(tt.flag)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("validFlagChecker() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if tt.wantText == "" {
				if err != nil {
					t.Fatalf("validFlagChecker() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), tt.wantText) {
				t.Fatalf("validFlagChecker() error = %v, want error containing %q", err, tt.wantText)
			}
		})
	}
}

func TestRootValidate(t *testing.T) {
	validExecute := func(*Context) error { return nil }
	validFlagExecute := func(*Context) error { return nil }

	t.Run("valid root", func(t *testing.T) {
		root := &Root{AppName: "app", Description: "test app", Commands: CommandList{
			"build": {
				Name:        "build",
				Description: "build it",
				Execute:     validExecute,
				Flags: map[string]*Flag{
					"verbose": {Name: "verbose", Short: "v", Description: "verbose output", Execute: validFlagExecute},
				},
			},
		}}
		if err := root.validate(); err != nil {
			t.Fatalf("validate() error = %v", err)
		}
	})

	t.Run("nil root", func(t *testing.T) {
		var root *Root
		if err := root.validate(); !errors.Is(err, ErrNilRoot) {
			t.Fatalf("validate() error = %v, want ErrNilRoot", err)
		}
	})

	t.Run("command key mismatch", func(t *testing.T) {
		root := &Root{AppName: "app", Description: "test app", Commands: CommandList{
			"wrong": {Name: "build", Description: "build it", Execute: validExecute},
		}}
		if err := root.validate(); err == nil || !strings.Contains(err.Error(), "command key") {
			t.Fatalf("validate() error = %v, want command key mismatch", err)
		}
	})

	t.Run("flag key mismatch", func(t *testing.T) {
		root := &Root{AppName: "app", Description: "test app", Commands: CommandList{
			"build": {
				Name:        "build",
				Description: "build it",
				Execute:     validExecute,
				Flags: map[string]*Flag{
					"wrong": {Name: "verbose", Description: "verbose output", Execute: validFlagExecute},
				},
			},
		}}
		if err := root.validate(); err == nil || !strings.Contains(err.Error(), "flag key") {
			t.Fatalf("validate() error = %v, want flag key mismatch", err)
		}
	})

	t.Run("duplicate short", func(t *testing.T) {
		root := &Root{AppName: "app", Description: "test app", Commands: CommandList{
			"build": {
				Name:        "build",
				Description: "build it",
				Execute:     validExecute,
				Flags: map[string]*Flag{
					"verbose": {Name: "verbose", Short: "v", Description: "verbose output", Execute: validFlagExecute},
					"version": {Name: "version", Short: "v", Description: "show version", Execute: validFlagExecute},
				},
			},
		}}
		if err := root.validate(); err == nil || !strings.Contains(err.Error(), "flag short") {
			t.Fatalf("validate() error = %v, want duplicate short error", err)
		}
	})
}

func TestRootAddCommand(t *testing.T) {
	root := &Root{}
	command := &Command{Name: "build", Description: "build it", Execute: func(*Context) error { return nil }}

	if err := root.AddCommand(command); err != nil {
		t.Fatalf("AddCommand() error = %v", err)
	}
	if got := root.Commands["build"]; got != command {
		t.Fatalf("stored command = %p, want %p", got, command)
	}
	if err := root.AddCommand(&Command{Name: "build", Description: "build it", Execute: func(*Context) error { return nil }}); err == nil {
		t.Fatal("AddCommand() accepted a duplicate name")
	}

	var nilRoot *Root
	if err := nilRoot.AddCommand(command); !errors.Is(err, ErrNilRoot) {
		t.Fatalf("nil root AddCommand() error = %v, want ErrNilRoot", err)
	}
}

func TestCommandAddFlag(t *testing.T) {
	command := &Command{}
	flag := &Flag{Name: "verbose", Short: "v", Description: "enable verbose output", Execute: func(*Context) error { return nil }}

	if err := command.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag() error = %v", err)
	}
	if got := command.Flags["verbose"]; got != flag {
		t.Fatalf("stored flag = %p, want %p", got, flag)
	}
	if err := command.AddFlag(flag); err == nil {
		t.Fatal("AddFlag() accepted a duplicate name")
	}
	if err := command.AddFlag(&Flag{Name: "version", Short: "v", Description: "show version", Execute: func(*Context) error { return nil }}); err == nil {
		t.Fatal("AddFlag() accepted a duplicate short name")
	}
	if err := command.AddFlag(&Flag{Name: "--invalid", Description: "invalid logical name", Execute: func(*Context) error { return nil }}); err == nil {
		t.Fatal("AddFlag() accepted a non-logical name")
	}

	invalid := &Flag{Name: "invalid", Description: "missing execute"}
	if err := command.AddFlag(invalid); err == nil {
		t.Fatal("AddFlag() accepted an invalid flag")
	}

	var nilCommand *Command
	if err := nilCommand.AddFlag(flag); !errors.Is(err, ErrNilCommand) {
		t.Fatalf("nil command AddFlag() error = %v, want ErrNilCommand", err)
	}
}

func TestRootRunExecutesParsedFlagsThenCommand(t *testing.T) {
	setTestArgs(t, "build", "input.go", "--output", "bin/app", "-v")

	var events []string
	command := &Command{Name: "build", Description: "build the project"}
	command.Flags = map[string]*Flag{
		"verbose": {
			Name:        "verbose",
			Short:       "v",
			Description: "enable verbose output",
			Execute: func(ctx *Context) error {
				events = append(events, "verbose")
				return nil
			},
		},
		"output": {
			Name:        "output",
			Short:       "o",
			Description: "set output path",
			TakesValue:  true,
			Execute: func(ctx *Context) error {
				events = append(events, "output")
				if got := ctx.Values["output"]; got != "bin/app" {
					t.Fatalf("ctx.Values[output] in callback = %#v, want bin/app", got)
				}
				return nil
			},
		},
	}
	command.Execute = func(ctx *Context) error {
		events = append(events, "command")
		if ctx.Command != command {
			t.Fatalf("context command = %p, want %p", ctx.Command, command)
		}
		if got := ctx.Values["verbose"]; got != true {
			t.Fatalf("ctx.Values[verbose] = %#v, want true", got)
		}
		if !reflect.DeepEqual(ctx.Args, []string{"input.go"}) {
			t.Fatalf("ctx.Args = %#v, want input.go", ctx.Args)
		}
		return nil
	}

	root := &Root{AppName: "app", Description: "test app", Commands: CommandList{"build": command}}
	if err := root.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := strings.Join(events, ","); got != "output,verbose,command" {
		t.Fatalf("events = %q, want output,verbose,command", got)
	}
}

func TestRootRunReturnsFlagErrorWithoutExecutingCommand(t *testing.T) {
	setTestArgs(t, "build", "--fail")

	wantErr := errors.New("bad flag")
	commandExecuted := false
	command := &Command{
		Name:        "build",
		Description: "build the project",
		Flags: map[string]*Flag{
			"fail": {Name: "fail", Description: "fail parsing", Execute: func(*Context) error { return wantErr }},
		},
		Execute: func(*Context) error {
			commandExecuted = true
			return nil
		},
	}

	err := (&Root{AppName: "app", Description: "test app", Commands: CommandList{"build": command}}).Run()
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
	if commandExecuted {
		t.Fatal("command executed after one of its flags failed")
	}
}

func TestRootRunReportsCommandError(t *testing.T) {
	setTestArgs(t, "build")

	wantErr := errors.New("build failed")
	command := &Command{Name: "build", Description: "build the project", Execute: func(*Context) error { return wantErr }}

	err := (&Root{AppName: "app", Description: "test app", Commands: CommandList{"build": command}}).Run()
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want wrapped %v", err, wantErr)
	}
	if err == nil || !strings.Contains(err.Error(), "command cannot be executed: build failed") {
		t.Fatalf("Run() error = %v, want contextual command error", err)
	}
}

func TestRootRunReportsMissingAndUnknownCommand(t *testing.T) {
	root := &Root{AppName: "app", Description: "test app", Commands: CommandList{
		"build": {Name: "build", Description: "build the project", Execute: func(*Context) error { return nil }},
	}}

	t.Run("missing command", func(t *testing.T) {
		setTestArgs(t)
		if err := root.Run(); !errors.Is(err, ErrMissingCommand) {
			t.Fatalf("Run() error = %v, want ErrMissingCommand", err)
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		setTestArgs(t, "unknown", "--flag")
		if err := root.Run(); !errors.Is(err, ErrUnknownCommand) {
			t.Fatalf("Run() error = %v, want ErrUnknownCommand", err)
		}
	})
}

func setTestArgs(t *testing.T, args ...string) {
	t.Helper()

	original := os.Args
	os.Args = append([]string{"clibox-test"}, args...)
	t.Cleanup(func() {
		os.Args = original
	})
}
