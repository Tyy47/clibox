package argbin

import (
	"errors"
	"testing"
)

func noopCommand(*Context) error { return nil }

func TestCommandParseSubcommandRecursesToDeepestMatch(t *testing.T) {
	leaf := &Command{Name: "leaf", Execute: noopCommand}
	middle := &Command{Name: "middle", Subcommands: []*Command{leaf}}
	parent := &Command{Name: "parent", Subcommands: []*Command{middle}}

	got, err := parent.parseSubcommand([]string{"middle", "leaf"})
	if err != nil {
		t.Fatalf("parseSubcommand returned an error: %v", err)
	}
	if got != leaf {
		t.Fatalf("parseSubcommand returned %q, want deepest matching command %q", got.Name, leaf.Name)
	}
}

func TestCommandParseSubcommandMatchesAliases(t *testing.T) {
	child := &Command{
		Name:            "deploy",
		AdditionalNames: []string{"d"},
		Execute:         noopCommand,
	}
	parent := &Command{Name: "app", Subcommands: []*Command{child}}

	got, err := parent.parseSubcommand([]string{"d"})
	if err != nil {
		t.Fatalf("parseSubcommand returned an error: %v", err)
	}
	if got != child {
		t.Fatalf("parseSubcommand returned %q, want alias target %q", got.Name, child.Name)
	}
}

func TestCommandParseSubcommandStopsAtCurrentCommand(t *testing.T) {
	parent := &Command{Name: "app", Subcommands: []*Command{{Name: "child", Execute: noopCommand}}}

	got, err := parent.parseSubcommand(nil)
	if err != nil {
		t.Fatalf("parseSubcommand returned an error: %v", err)
	}
	if got != parent {
		t.Fatalf("parseSubcommand returned %q, want current command %q", got.Name, parent.Name)
	}
}

func TestCommandParseSubcommandRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		cmd  *Command
		args []string
		want error
	}{
		{
			name: "nil command",
			cmd:  nil,
			want: ErrNilCommand,
		},
		{
			name: "nil child",
			cmd:  &Command{Name: "app", Subcommands: []*Command{nil}},
			args: []string{"child"},
			want: ErrNilCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.cmd.parseSubcommand(tt.args)
			if !errors.Is(err, tt.want) {
				t.Fatalf("parseSubcommand error = %v, want %v", err, tt.want)
			}
		})
	}
}

// These tests define the failure behavior needed to avoid selecting a grouping
// command and subsequently calling its nil Execute function.
func TestCommandParseSubcommandRejectsUnknownSubcommand(t *testing.T) {
	parent := &Command{Name: "app", Subcommands: []*Command{{Name: "known", Execute: noopCommand}}}

	got, err := parent.parseSubcommand([]string{"unknown"})
	if err == nil {
		t.Fatalf("parseSubcommand returned (%q, nil) for an unknown subcommand; want an error", got.Name)
	}
}

func TestRootParseCommandDoesNotMatchAnUnrelatedSubcommandTree(t *testing.T) {
	group := &Command{Name: "group", Subcommands: []*Command{{Name: "child", Execute: noopCommand}}}
	root := &Root{CommandList: []*Command{group}}
	ctx := &Context{Args: []string{"unknown"}}

	got, err := root.parseCommand(ctx, "unknown")
	if err != nil {
		t.Fatalf("parseCommand returned an unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("parseCommand returned %q for an unknown root command, want nil", got.Name)
	}
}

func TestRootParseCommandResolvesNestedSubcommand(t *testing.T) {
	leaf := &Command{Name: "leaf", Execute: noopCommand}
	middle := &Command{Name: "middle", Subcommands: []*Command{leaf}}
	parent := &Command{Name: "parent", Subcommands: []*Command{middle}}
	root := &Root{CommandList: []*Command{parent}}
	ctx := &Context{Args: []string{"parent", "middle", "leaf"}}

	got, err := root.parseCommand(ctx, "parent")
	if err != nil {
		t.Fatalf("parseCommand returned an error: %v", err)
	}
	if got != leaf || ctx.Command != leaf {
		t.Fatalf("parseCommand selected %q (context %q), want leaf", got.Name, ctx.Command.Name)
	}
}

func TestRootParseCommandRejectsUnknownNestedSubcommand(t *testing.T) {
	parent := &Command{Name: "parent", Subcommands: []*Command{{Name: "known", Execute: noopCommand}}}
	root := &Root{CommandList: []*Command{parent}}
	ctx := &Context{Args: []string{"parent", "unknown"}}

	_, err := root.parseCommand(ctx, "parent")
	if err == nil {
		t.Fatal("parseCommand returned nil for an unknown nested subcommand, want an error")
	}
}

func TestRootParseCommandRejectsGroupWithoutSubcommand(t *testing.T) {
	parent := &Command{Name: "parent", Subcommands: []*Command{{Name: "child", Execute: noopCommand}}}
	root := &Root{CommandList: []*Command{parent}}
	ctx := &Context{Args: []string{"parent"}}

	_, err := root.parseCommand(ctx, "parent")
	if err == nil {
		t.Fatal("parseCommand selected a non-executable group without a subcommand, want an error")
	}
}

func TestCommandValidateAllowsGroupCommandsButNotEmptyCommands(t *testing.T) {
	group := &Command{Name: "group", Subcommands: []*Command{{Name: "child", Execute: noopCommand}}}
	if err := group.validate(); err != nil {
		t.Fatalf("group command validation error = %v, want nil", err)
	}

	empty := &Command{Name: "empty"}
	if err := empty.validate(); !errors.Is(err, ErrNilCommandFunction) {
		t.Fatalf("empty command validation error = %v, want %v", err, ErrNilCommandFunction)
	}
}
