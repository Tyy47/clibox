package argbin

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCommandSearchFlag(t *testing.T) {
	verbose := &Flag{Name: "verbose"}
	output := &Flag{Name: "output"}
	command := &Command{
		Flags: map[string]*Flag{
			"verbose": verbose,
			"output":  output,
		},
	}

	tests := []struct {
		name string
		arg  string
		want *Flag
		ok   bool
	}{
		{name: "known key", arg: "verbose", want: verbose, ok: true},
		{name: "another known key", arg: "output", want: output, ok: true},
		{name: "cli long spelling is not normalized yet", arg: "--verbose", want: nil, ok: false},
		{name: "cli short spelling is not used yet", arg: "-v", want: nil, ok: false},
		{name: "unknown", arg: "missing", want: nil, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := command.searchFlag(tt.arg)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("searchFlag(%q) = (%p, %t), want (%p, %t)", tt.arg, got, ok, tt.want, tt.ok)
			}
		})
	}

	t.Run("nil command", func(t *testing.T) {
		var command *Command
		if got, ok := command.searchFlag("verbose"); ok || got != nil {
			t.Fatalf("searchFlag() = (%v, %t), want (nil, false)", got, ok)
		}
	})

	t.Run("nil flag map", func(t *testing.T) {
		command := &Command{}
		if got, ok := command.searchFlag("anything"); ok || got != nil {
			t.Fatalf("searchFlag() = (%v, %t), want (nil, false)", got, ok)
		}
	})
}

func TestCommandParseFlagsExecutesKnownFlagsInOrder(t *testing.T) {
	var executionOrder []string
	command := &Command{
		Flags: map[string]*Flag{
			"verbose": {
				Name:        "verbose",
				Description: "enable verbose output",
				Execute: func(ctx *Context) error {
					executionOrder = append(executionOrder, "verbose")
					ctx.Values["verbose"] = true
					return nil
				},
			},
			"force": {
				Name:        "force",
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

	err := command.parseFlags(ctx, []string{"verbose", "force"})
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

func TestCommandParseFlagsRejectsUnknownFlag(t *testing.T) {
	command := &Command{Flags: map[string]*Flag{}}

	err := command.parseFlags(&Context{Values: make(map[string]any)}, []string{"missing"})
	if !errors.Is(err, ErrUnknownFlag) {
		t.Fatalf("parseFlags() error = %v, want ErrUnknownFlag", err)
	}
}

func TestCommandParseFlagsNilGuards(t *testing.T) {
	ctx := &Context{Values: make(map[string]any)}

	var command *Command
	if err := command.parseFlags(ctx, nil); !errors.Is(err, ErrNilCommand) {
		t.Fatalf("nil command parseFlags() error = %v, want ErrNilCommand", err)
	}

	command = &Command{}
	if err := command.parseFlags(nil, nil); !errors.Is(err, ErrNilContext) {
		t.Fatalf("nil context parseFlags() error = %v, want ErrNilContext", err)
	}
}

func TestCommandParseFlagsStopsOnError(t *testing.T) {
	wantErr := errors.New("flag failed")
	laterExecuted := false
	command := &Command{
		Flags: map[string]*Flag{
			"fail": {
				Name:        "fail",
				Description: "return an error",
				Execute: func(*Context) error {
					return wantErr
				},
			},
			"later": {
				Name:        "later",
				Description: "must not execute",
				Execute: func(*Context) error {
					laterExecuted = true
					return nil
				},
			},
		},
	}

	err := command.parseFlags(&Context{Values: make(map[string]any)}, []string{"fail", "later"})
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
			"invalid": {
				Name:        "invalid",
				Description: "",
				Execute:     func(*Context) error { return nil },
			},
		},
	}

	err := command.parseFlags(&Context{Values: make(map[string]any)}, []string{"invalid"})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "description") {
		t.Fatalf("parseFlags() error = %v, want missing-description error", err)
	}
}

func TestCommandParseFlagsDoesNotConsumeTakesValueYet(t *testing.T) {
	executed := false
	command := &Command{
		Flags: map[string]*Flag{
			"output": {
				Name:        "output",
				Description: "set output",
				TakesValue:  true,
				Execute: func(*Context) error {
					executed = true
					return nil
				},
			},
		},
	}

	err := command.parseFlags(&Context{Values: make(map[string]any)}, []string{"output", "file.txt"})
	if !errors.Is(err, ErrUnknownFlag) {
		t.Fatalf("parseFlags() error = %v, want ErrUnknownFlag for unconsumed value", err)
	}
	if !executed {
		t.Fatal("current parser should execute the flag before failing on the value token")
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
		{name: "valid", flag: &Flag{Name: "force", Description: "force it", Execute: validExecute}},
		{name: "nil flag", flag: nil, wantErr: ErrNilFlag},
		{name: "missing name", flag: &Flag{Description: "force it", Execute: validExecute}, wantText: "name"},
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
		root := &Root{
			AppName:     "app",
			Description: "test app",
			Commands: CommandList{
				"build": {
					Name:        "build",
					Description: "build it",
					Execute:     validExecute,
					Flags: map[string]*Flag{
						"verbose": {Name: "verbose", Description: "verbose output", Execute: validFlagExecute},
					},
				},
			},
		}
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
		root := &Root{
			AppName:     "app",
			Description: "test app",
			Commands: CommandList{
				"wrong": {Name: "build", Description: "build it", Execute: validExecute},
			},
		}
		if err := root.validate(); err == nil || !strings.Contains(err.Error(), "command key") {
			t.Fatalf("validate() error = %v, want command key mismatch", err)
		}
	})

	t.Run("flag key mismatch", func(t *testing.T) {
		root := &Root{
			AppName:     "app",
			Description: "test app",
			Commands: CommandList{
				"build": {
					Name:        "build",
					Description: "build it",
					Execute:     validExecute,
					Flags: map[string]*Flag{
						"wrong": {Name: "verbose", Description: "verbose output", Execute: validFlagExecute},
					},
				},
			},
		}
		if err := root.validate(); err == nil || !strings.Contains(err.Error(), "flag key") {
			t.Fatalf("validate() error = %v, want flag key mismatch", err)
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
	flag := &Flag{
		Name:        "verbose",
		Description: "enable verbose output",
		Execute:     func(*Context) error { return nil },
	}

	if err := command.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag() error = %v", err)
	}
	if got := command.Flags["verbose"]; got != flag {
		t.Fatalf("stored flag = %p, want %p", got, flag)
	}
	if err := command.AddFlag(flag); err == nil {
		t.Fatal("AddFlag() accepted a duplicate name")
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

func TestRootRunExecutesFlagsBeforeCommand(t *testing.T) {
	setTestArgs(t, "build", "verbose")

	var flagExecuted bool
	command := &Command{Name: "build", Description: "build the project"}
	command.Flags = map[string]*Flag{
		"verbose": {
			Name:        "verbose",
			Description: "enable verbose output",
			Short:       "v",
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

	root := &Root{AppName: "app", Description: "test app", Commands: CommandList{"build": command}}
	if err := root.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRootRunReturnsFlagErrorWithoutExecutingCommand(t *testing.T) {
	setTestArgs(t, "build", "fail")

	wantErr := errors.New("bad flag")
	commandExecuted := false
	command := &Command{
		Name:        "build",
		Description: "build the project",
		Flags: map[string]*Flag{
			"fail": {
				Name:        "fail",
				Description: "fail parsing",
				Execute:     func(*Context) error { return wantErr },
			},
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
	command := &Command{
		Name:        "build",
		Description: "build the project",
		Execute:     func(*Context) error { return wantErr },
	}

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
		setTestArgs(t, "unknown", "flag")
		if err := root.Run(); !errors.Is(err, ErrUnknownCommand) {
			t.Fatalf("Run() error = %v, want ErrUnknownCommand", err)
		}
	})
}

func TestRootRunRejectsUnknownFlag(t *testing.T) {
	setTestArgs(t, "build", "unknown")
	executed := false
	root := &Root{AppName: "app", Description: "test app", Commands: CommandList{
		"build": {
			Name:        "build",
			Description: "build the project",
			Execute: func(*Context) error {
				executed = true
				return nil
			},
		},
	}}

	err := root.Run()
	if !errors.Is(err, ErrUnknownFlag) {
		t.Fatalf("Run() error = %v, want ErrUnknownFlag", err)
	}
	if executed {
		t.Fatal("command executed after unknown flag")
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
