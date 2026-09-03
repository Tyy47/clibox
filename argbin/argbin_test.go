package argbin

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCommandSearchFlag(t *testing.T) {
	byKey := &Flag{Name: "verbose"}
	byName := &Flag{Name: "output"}
	byAlias := &Flag{Name: "force", Aliases: []string{"-f", "--force"}}
	command := &Command{
		Flags: map[string]*Flag{
			"--verbose":  byKey,
			"output-key": byName,
			"force-key":  byAlias,
		},
	}

	tests := []struct {
		name string
		arg  string
		want *Flag
		ok   bool
	}{
		{name: "map key", arg: "--verbose", want: byKey, ok: true},
		{name: "flag name", arg: "output", want: byName, ok: true},
		{name: "short alias", arg: "-f", want: byAlias, ok: true},
		{name: "long alias", arg: "--force", want: byAlias, ok: true},
		{name: "unknown", arg: "--missing", want: nil, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := command.searchFlag(tt.arg)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("searchFlag(%q) = (%p, %t), want (%p, %t)", tt.arg, got, ok, tt.want, tt.ok)
			}
		})
	}

	t.Run("nil flag map", func(t *testing.T) {
		command := &Command{}
		if got, ok := command.searchFlag("--anything"); ok || got != nil {
			t.Fatalf("searchFlag() = (%v, %t), want (nil, false)", got, ok)
		}
	})
}

func TestCommandParseFlagsMutatesSharedContext(t *testing.T) {
	var executionOrder []string
	command := &Command{
		Flags: map[string]*Flag{
			"--verbose": {
				Name:        "--verbose",
				Description: "enable verbose output",
				Aliases:     []string{"-v"},
				Execute: func(ctx *Context) error {
					executionOrder = append(executionOrder, "verbose")
					ctx.Values["verbose"] = true
					return nil
				},
			},
			"--force": {
				Name:        "--force",
				Description: "force the operation",
				Execute: func(ctx *Context) error {
					executionOrder = append(executionOrder, "force")
					ctx.Values["force"] = true
					return nil
				},
			},
		},
	}
	ctx := &Context{Command: command, Values: make(map[string]any)}

	err := command.parseFlags(ctx, []string{"ignored", "-v", "--force"})
	if err != nil {
		t.Fatalf("parseFlags() error = %v", err)
	}
	if got, ok := ctx.Values["verbose"].(bool); !ok || !got {
		t.Fatalf("verbose context value = %#v, want true", ctx.Values["verbose"])
	}
	if got, ok := ctx.Values["force"].(bool); !ok || !got {
		t.Fatalf("force context value = %#v, want true", ctx.Values["force"])
	}
	if got := strings.Join(executionOrder, ","); got != "verbose,force" {
		t.Fatalf("flag execution order = %q, want %q", got, "verbose,force")
	}
}

func TestCommandParseFlagsStopsOnError(t *testing.T) {
	wantErr := errors.New("flag failed")
	laterExecuted := false
	command := &Command{
		Flags: map[string]*Flag{
			"--fail": {
				Name:        "--fail",
				Description: "return an error",
				Execute: func(*Context) error {
					return wantErr
				},
			},
			"--later": {
				Name:        "--later",
				Description: "must not execute",
				Execute: func(*Context) error {
					laterExecuted = true
					return nil
				},
			},
		},
	}

	err := command.parseFlags(&Context{Values: make(map[string]any)}, []string{"--fail", "--later"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("parseFlags() error = %v, want %v", err, wantErr)
	}
	if laterExecuted {
		t.Fatal("flag after a failed flag was executed")
	}
}

func TestCommandParseFlagsRejectsInvalidKnownFlag(t *testing.T) {
	command := &Command{
		Flags: map[string]*Flag{
			"--invalid": {
				Name:        "--invalid",
				Description: "",
				Execute:     func(*Context) error { return nil },
			},
		},
	}

	err := command.parseFlags(&Context{Values: make(map[string]any)}, []string{"--invalid"})
	if err == nil || !strings.Contains(err.Error(), "description") {
		t.Fatalf("parseFlags() error = %v, want missing-description error", err)
	}
}

func TestValidCommandChecker(t *testing.T) {
	validExecute := func(*Context) error { return nil }
	tests := []struct {
		name     string
		commands CommandList
		wantText string
	}{
		{name: "empty list", commands: CommandList{}},
		{name: "valid", commands: CommandList{"build": {Name: "build", Description: "build it", Execute: validExecute}}},
		{name: "missing name", commands: CommandList{"build": {Description: "build it", Execute: validExecute}}, wantText: "name"},
		{name: "missing description", commands: CommandList{"build": {Name: "build", Execute: validExecute}}, wantText: "description"},
		{name: "missing execute", commands: CommandList{"build": {Name: "build", Description: "build it"}}, wantText: "execute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validCommandChecker(tt.commands)
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
		wantText string
	}{
		{name: "valid", flag: &Flag{Name: "--force", Description: "force it", Execute: validExecute}},
		{name: "missing name", flag: &Flag{Description: "force it", Execute: validExecute}, wantText: "name"},
		{name: "missing description", flag: &Flag{Name: "--force", Execute: validExecute}, wantText: "description"},
		{name: "missing execute", flag: &Flag{Name: "--force", Description: "force it"}, wantText: "execute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validFlagChecker(tt.flag)
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

func TestRootAddCommand(t *testing.T) {
	root := &Root{Commands: make(CommandList)}
	command := &Command{Name: "build"}

	if err := root.AddCommand(command); err != nil {
		t.Fatalf("AddCommand() error = %v", err)
	}
	if got := root.Commands["build"]; got != command {
		t.Fatalf("stored command = %p, want %p", got, command)
	}
	if err := root.AddCommand(&Command{Name: "build"}); err == nil {
		t.Fatal("AddCommand() accepted a duplicate name")
	}
}

func TestCommandAddFlag(t *testing.T) {
	command := &Command{Flags: make(map[string]*Flag)}
	flag := &Flag{
		Name:        "--verbose",
		Description: "enable verbose output",
		Execute:     func(*Context) error { return nil },
	}

	if err := command.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag() error = %v", err)
	}
	if got := command.Flags["--verbose"]; got != flag {
		t.Fatalf("stored flag = %p, want %p", got, flag)
	}
	if err := command.AddFlag(flag); err == nil {
		t.Fatal("AddFlag() accepted a duplicate name")
	}

	invalid := &Flag{Name: "--invalid", Description: "missing execute"}
	if err := command.AddFlag(invalid); err == nil {
		t.Fatal("AddFlag() accepted an invalid flag")
	}
}

func TestRootRunExecutesFlagsBeforeCommand(t *testing.T) {
	setTestArgs(t, "build", "unknown", "-v")

	var flagExecuted bool
	command := &Command{Name: "build", Description: "build the project"}
	command.Flags = map[string]*Flag{
		"--verbose": {
			Name:        "--verbose",
			Description: "enable verbose output",
			Aliases:     []string{"-v"},
			Execute: func(ctx *Context) error {
				flagExecuted = true
				ctx.Values["verbose"] = true
				return nil
			},
		},
	}
	command.Execute = func(ctx *Context) error {
		if !flagExecuted {
			t.Fatal("command executed before its flag")
		}
		if ctx.Command != command {
			t.Fatalf("context command = %p, want %p", ctx.Command, command)
		}
		if verbose, ok := ctx.Values["verbose"].(bool); !ok || !verbose {
			t.Fatalf("verbose context value = %#v, want true", ctx.Values["verbose"])
		}
		return nil
	}

	root := &Root{Commands: CommandList{"build": command}}
	if err := root.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
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
			"--fail": {
				Name:        "--fail",
				Description: "fail parsing",
				Execute:     func(*Context) error { return wantErr },
			},
		},
		Execute: func(*Context) error {
			commandExecuted = true
			return nil
		},
	}

	err := (&Root{Commands: CommandList{"build": command}}).Run()
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
	if commandExecuted {
		t.Fatal("command executed after one of its flags failed")
	}
}

func TestRootRunReportsCommandError(t *testing.T) {
	setTestArgs(t, "build")

	command := &Command{
		Name:        "build",
		Description: "build the project",
		Execute:     func(*Context) error { return errors.New("build failed") },
	}

	err := (&Root{Commands: CommandList{"build": command}}).Run()
	if err == nil || !strings.Contains(err.Error(), "Command cannot be executed: build failed") {
		t.Fatalf("Run() error = %v, want contextual command error", err)
	}
}

func TestRootRunWithNoMatchingCommandIsNoOp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "no arguments"},
		{name: "unknown argument", args: []string{"unknown", "--flag"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestArgs(t, tt.args...)
			executed := false
			root := &Root{Commands: CommandList{
				"build": {
					Name:        "build",
					Description: "build the project",
					Execute: func(*Context) error {
						executed = true
						return nil
					},
				},
			}}

			if err := root.Run(); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if executed {
				t.Fatal("command executed without a matching command argument")
			}
		})
	}
}

func setTestArgs(t *testing.T, args ...string) {
	t.Helper()

	original := os.Args
	os.Args = append([]string{"clibox-test"}, args...)
	t.Cleanup(func() {
		os.Args = original
	})
}
