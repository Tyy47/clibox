package argbin

import (
	"errors"
	"slices"
	"strconv"
	"testing"
)

// These tests change os.Args via withArgs and must not run in parallel.
func TestRootRunTakesValue(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		takesValue bool
		wantValue  string
		wantCalls  int
		wantFlags  int
		wantError  string
	}{
		{name: "string", args: []string{"set", "Tyler"}, takesValue: true, wantValue: "Tyler", wantCalls: 1},
		{name: "alias", args: []string{"s", "example"}, takesValue: true, wantValue: "example", wantCalls: 1},
		{name: "spaces", args: []string{"set", "hello world"}, takesValue: true, wantValue: "hello world", wantCalls: 1},
		{name: "explicit empty value", args: []string{"set", ""}, takesValue: true, wantCalls: 1},
		{name: "negative number", args: []string{"set", "-42"}, takesValue: true, wantValue: "-42", wantCalls: 1},
		{name: "command name as value", args: []string{"set", "set"}, takesValue: true, wantValue: "set", wantCalls: 1},
		{name: "missing value", args: []string{"set"}, takesValue: true, wantError: "command set requires a value"},
		{name: "disabled", args: []string{"set"}, wantCalls: 1},
		{name: "disabled with flag", args: []string{"set", "--verbose"}, wantCalls: 1, wantFlags: 1},
		{name: "value then flag", args: []string{"set", "example", "--verbose"}, takesValue: true, wantValue: "example", wantCalls: 1, wantFlags: 1},
		// Flag parsing precedes value assignment, as accepted in ISSUES.md.
		{name: "registered flag as value", args: []string{"set", "--verbose"}, takesValue: true, wantValue: "--verbose", wantCalls: 1, wantFlags: 1},
		{name: "unknown flag as value", args: []string{"set", "--literal"}, takesValue: true, wantError: "unknown arguments: --literal"},
		{name: "skip non-command before command", args: []string{"ignored", "set", "example"}, takesValue: true, wantValue: "example", wantCalls: 1},
		{name: "value after two skipped tokens", args: []string{"ignored", "also-ignored", "set", "example"}, takesValue: true, wantValue: "example", wantCalls: 1},
		{name: "alias after two skipped tokens", args: []string{"ignored", "also-ignored", "s", "example"}, takesValue: true, wantValue: "example", wantCalls: 1},
		{name: "missing value after skipped tokens", args: []string{"one", "two", "three", "set"}, takesValue: true, wantError: "command set requires a value"},
		{name: "unknown trailing flag", args: []string{"set", "example", "--unknown"}, takesValue: true, wantError: "unknown arguments: --unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withArgs(t, tt.args...)
			defer func() {
				if rec := recover(); rec != nil {
					t.Fatalf("Run() panicked: %v", rec)
				}
			}()

			calls, flags := 0, 0
			cmd := &Command{
				Name: "set", AdditionalNames: []string{"s"}, TakesValue: tt.takesValue,
				Flags: Flags{"--verbose": func(*Context) error { flags++; return nil }},
			}
			cmd.Execute = func(ctx *Context) error {
				calls++
				if ctx.Command != cmd {
					t.Error("Execute received the wrong command")
				}
				if ctx.ParsedValue != tt.wantValue {
					t.Errorf("ParsedValue = %q, want %q", ctx.ParsedValue, tt.wantValue)
				}
				if !slices.Equal(ctx.Args, tt.args) {
					t.Errorf("Args = %q, want %q", ctx.Args, tt.args)
				}
				return nil
			}
			root := &Root{AppName: "test", AppVersion: "1.0.0", Description: "value tests", CommandList: []*Command{cmd}}
			err := root.Run()
			if tt.wantError == "" {
				if err != nil {
					t.Errorf("Run() unexpected error: %v", err)
				}
			} else if err == nil || err.Error() != tt.wantError {
				t.Errorf("Run() error = %v, want %q", err, tt.wantError)
			}
			if calls != tt.wantCalls {
				t.Errorf("Execute calls = %d, want %d", calls, tt.wantCalls)
			}
			if flags != tt.wantFlags {
				t.Errorf("flag calls = %d, want %d", flags, tt.wantFlags)
			}
		})
	}
}

func TestRootRunTakesValueParseInteger(t *testing.T) {
	for _, input := range []string{"42", "not-a-number"} {
		t.Run(input, func(t *testing.T) {
			withArgs(t, "age", input)
			calls, got := 0, 0
			var parseErr error
			root := &Root{
				AppName: "test", AppVersion: "1.0.0", Description: "integer parsing example",
				CommandList: []*Command{{
					Name: "age", TakesValue: true,
					Execute: func(ctx *Context) error {
						calls++
						got, parseErr = strconv.Atoi(ctx.ParsedValue)
						return parseErr
					},
				}},
			}
			err := root.Run()
			if calls != 1 {
				t.Fatalf("Execute calls = %d, want 1", calls)
			}
			if input == "42" {
				if err != nil || got != 42 {
					t.Fatalf("value = %d, error = %v; want 42, nil", got, err)
				}
			} else {
				var numErr *strconv.NumError
				if !errors.As(err, &numErr) {
					t.Fatalf("error = %v, want *strconv.NumError", err)
				}
				if !errors.Is(err, parseErr) {
					t.Fatalf("error = %v, want original callback error %v", err, parseErr)
				}
			}
		})
	}
}
