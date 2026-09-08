package argbin

import (
	"errors"
	"strconv"
	"testing"
)

func TestRootRunTakesValueParseBool(t *testing.T) {
	for _, input := range []string{"true", "false", "invalid"} {
		t.Run(input, func(t *testing.T) {
			withArgs(t, "enabled", input)
			calls := 0
			var got bool
			var parseErr error
			root := &Root{
				AppName: "test", AppVersion: "1.0.0", Description: "boolean parsing example",
				CommandList: []*Command{{
					Name: "enabled", TakesValue: true,
					Execute: func(ctx *Context) error {
						calls++
						got, parseErr = strconv.ParseBool(ctx.ParsedValue)
						return parseErr
					},
				}},
			}
			err := root.Run()
			if calls != 1 {
				t.Fatalf("Execute calls = %d, want 1", calls)
			}
			if input == "invalid" {
				var numErr *strconv.NumError
				if !errors.As(err, &numErr) || !errors.Is(err, parseErr) {
					t.Fatalf("Run() error = %v, want original ParseBool error %v", err, parseErr)
				}
				return
			}
			if err != nil || got != (input == "true") {
				t.Fatalf("parsed value = %v, error = %v; want %s, nil", got, err, input)
			}
		})
	}
}

func TestRootRunTakesValueFlagError(t *testing.T) {
	withArgs(t, "set", "example", "--fail")
	flagErr := errors.New("flag failed")
	executed, flagCalls := false, 0
	root := &Root{
		AppName: "test", AppVersion: "1.0.0", Description: "flag failure example",
		CommandList: []*Command{{
			Name: "set", TakesValue: true,
			Flags: Flags{"--fail": func(*Context) error {
				flagCalls++
				return flagErr
			}},
			Execute: func(*Context) error {
				executed = true
				return nil
			},
		}},
	}
	if err := root.Run(); !errors.Is(err, flagErr) {
		t.Errorf("Run() error = %v, want %v", err, flagErr)
	}
	if flagCalls != 1 || executed {
		t.Errorf("flag calls = %d, Execute called = %v; want 1, false", flagCalls, executed)
	}
}
