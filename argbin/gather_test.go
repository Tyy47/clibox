package argbin

import (
	"errors"
	"reflect"
	"testing"
)

func TestContextGatherMultipleParsedValues(t *testing.T) {
	cmd := &Command{Name: "serve"}
	tests := []struct {
		name string
		ctx  *Context
		args []string
		want []string
		err  error
	}{
		{
			name: "gathers values in order",
			ctx:  &Context{},
			args: []string{"one", "two", "three"},
			want: []string{"one", "two", "three"},
		},
		{
			name: "stops at first flag",
			ctx:  &Context{},
			args: []string{"one", "two", "--verbose", "three"},
			want: []string{"one", "two"},
		},
		{
			name: "requires at least one argument",
			ctx:  &Context{},
			err:  ErrRequiresInput,
		},
		{
			name: "nil context",
			err:  ErrNilContext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ctx.gatherMultipleParsedValues(tt.args, cmd)
			if !errors.Is(err, tt.err) {
				t.Fatalf("gatherMultipleParsedValues() error = %v, want %v", err, tt.err)
			}
			if tt.err == nil && !reflect.DeepEqual(tt.ctx.MultipleParsedValues, tt.want) {
				t.Fatalf("MultipleParsedValues = %q, want %q", tt.ctx.MultipleParsedValues, tt.want)
			}
		})
	}
}

func TestRootRunGathersMultipleCommandValues(t *testing.T) {
	withArgs(t, "serve", "first", "second", "third")

	root := validRoot(&Command{
		Name:       "serve",
		TakesValue: true,
		Execute: func(ctx *Context) error {
			want := []string{"first", "second", "third"}
			if !reflect.DeepEqual(ctx.MultipleParsedValues, want) {
				t.Errorf("MultipleParsedValues = %q, want %q", ctx.MultipleParsedValues, want)
			}
			return nil
		},
	})

	if err := root.Run(); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
}

func TestRootRunGathersMultipleFlagValues(t *testing.T) {
	withArgs(t, "serve", "--tag", "first", "--tag", "second")

	calls := 0
	flag := &Flag{
		TakesValue: true,
		Execute: func(ctx *Context) error {
			calls++
			wantValue := []string{"first", "second"}[calls-1]
			if ctx.ParsedFlagValue != wantValue {
				t.Errorf("ParsedFlagValue = %q, want %q", ctx.ParsedFlagValue, wantValue)
			}
			wantValues := []string{"first", "second"}[:calls]
			if !reflect.DeepEqual(ctx.MultipleParsedFlagValues, wantValues) {
				t.Errorf("MultipleParsedFlagValues = %q, want %q", ctx.MultipleParsedFlagValues, wantValues)
			}
			return nil
		},
	}
	root := validRoot(&Command{
		Name:    "serve",
		Flags:   Flags{"--tag": flag},
		Execute: func(*Context) error { return nil },
	})

	if err := root.Run(); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if calls != 2 {
		t.Fatalf("flag Execute calls = %d, want 2", calls)
	}
}
