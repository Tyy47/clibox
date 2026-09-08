package argbin

import (
	"slices"
	"testing"
)

// AdditionalArgs deliberately starts three tokens after the command:
// command, value-or-flag, secondary-value, additional...
func TestRootRunTakesValueAdditionalArgs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		commandIndex   int
		wantValue      string
		wantAdditional []string
		wantOutput     string
	}{
		{name: "value only", args: []string{"set", "example"}, wantValue: "example"},
		{name: "secondary value only", args: []string{"set", "example", "secondary"}, wantValue: "example"},
		{name: "additional values", args: []string{"set", "example", "secondary", "one", "two"}, wantValue: "example", wantAdditional: []string{"one", "two"}},
		{name: "output destination", args: []string{"set", "--output", "out.txt"}, wantValue: "--output", wantOutput: "out.txt"},
		{name: "output with extras", args: []string{"set", "--output", "out.txt", "one", "two"}, wantValue: "--output", wantOutput: "out.txt", wantAdditional: []string{"one", "two"}},
		{name: "skipped prefix with no secondary value", args: []string{"ignored", "also-ignored", "set", "example"}, commandIndex: 2, wantValue: "example"},
		{name: "skipped prefix at slice boundary", args: []string{"ignored", "also-ignored", "set", "example", "secondary"}, commandIndex: 2, wantValue: "example"},
		{name: "skipped prefix with extras", args: []string{"ignored", "also-ignored", "set", "example", "secondary", "extra"}, commandIndex: 2, wantValue: "example", wantAdditional: []string{"extra"}},
		{name: "alias with output and extras", args: []string{"ignored", "also-ignored", "s", "--output", "out.txt", "extra"}, commandIndex: 2, wantValue: "--output", wantOutput: "out.txt", wantAdditional: []string{"extra"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withArgs(t, tt.args...)
			defer func() {
				if rec := recover(); rec != nil {
					t.Fatalf("Run() panicked: %v", rec)
				}
			}()
			calls, flagCalls := 0, 0
			cmd := &Command{Name: "set", AdditionalNames: []string{"s"}, TakesValue: true}
			cmd.Flags = Flags{"--output": func(ctx *Context) error {
				flagCalls++
				// Flags run before ParsedValue/AdditionalArgs are assigned.
				// The destination remains accessible in the original Args.
				ctx.Values["output"] = ctx.Args[tt.commandIndex+2]
				return nil
			}}
			cmd.Execute = func(ctx *Context) error {
				calls++
				if ctx.Command != cmd || ctx.ParsedValue != tt.wantValue {
					t.Errorf("command/value = %p/%q, want %p/%q", ctx.Command, ctx.ParsedValue, cmd, tt.wantValue)
				}
				if !slices.Equal(ctx.AdditionalArgs, tt.wantAdditional) {
					t.Errorf("AdditionalArgs = %q, want %q", ctx.AdditionalArgs, tt.wantAdditional)
				}
				if !slices.Equal(ctx.Args, tt.args) {
					t.Errorf("Args = %q, want %q", ctx.Args, tt.args)
				}
				if tt.wantOutput != "" && ctx.Values["output"] != tt.wantOutput {
					t.Errorf("output = %v, want %q", ctx.Values["output"], tt.wantOutput)
				}
				return nil
			}
			root := &Root{AppName: "test", AppVersion: "1.0.0", Description: "argument layout", CommandList: []*Command{cmd}}
			if err := root.Run(); err != nil {
				t.Fatalf("Run() unexpected error: %v", err)
			}
			if calls != 1 {
				t.Errorf("Execute calls = %d, want 1", calls)
			}
			wantFlags := 0
			if tt.wantOutput != "" {
				wantFlags = 1
			}
			if flagCalls != wantFlags {
				t.Errorf("flag calls = %d, want %d", flagCalls, wantFlags)
			}
		})
	}
}
