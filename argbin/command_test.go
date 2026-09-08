package argbin

import (
	"errors"
	"slices"
	"testing"
)

func TestCommandNilReceiver(t *testing.T) {
	var c *Command
	tests := []struct {
		name string
		call func() error
	}{
		{"GetName", func() error {
			v, e := c.GetName()
			if v != "" {
				t.Error("nonzero name")
			}
			return e
		}},
		{"SetName", func() error { return c.SetName("run") }},
		{"GetDescription", func() error {
			v, e := c.GetDescription()
			if v != "" {
				t.Error("nonzero description")
			}
			return e
		}},
		{"SetDescription", func() error { return c.SetDescription("description") }},
		{"GetAdditionalNames", func() error {
			v, e := c.GetAdditionalNames()
			if v != nil {
				t.Error("non-nil aliases")
			}
			return e
		}},
		{"SetAdditionalNames", func() error { return c.SetAdditionalNames([]string{"r"}) }},
		{"AddAdditionalName", func() error { return c.AddAdditionalName("r") }},
		{"GetFlags", func() error {
			v, e := c.GetFlags()
			if v != nil {
				t.Error("non-nil flags")
			}
			return e
		}},
		{"SetFlags", func() error { return c.SetFlags(Flags{}) }},
		{"AddFlag", func() error { return c.AddFlag("--verbose", func(*Context) error { return nil }) }},
		{"validate", func() error { return c.validate() }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); !errors.Is(err, ErrNilCommand) {
				t.Fatalf("error = %v, want %v", err, ErrNilCommand)
			}
		})
	}
}

func TestCommandNameAndDescription(t *testing.T) {
	c := &Command{}
	if got, err := c.GetName(); err != nil || got != "" {
		t.Fatalf("GetName() = %q, %v", got, err)
	}
	if _, err := c.GetDescription(); err == nil {
		t.Fatal("empty description should error")
	}
	if err := c.SetName("run"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetDescription("Run the app"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetName(""); !errors.Is(err, ErrEmptyCommandName) {
		t.Fatalf("SetName error = %v", err)
	}
	if err := c.SetDescription(""); err == nil {
		t.Fatal("empty description should error")
	}
	if got, err := c.GetName(); err != nil || got != "run" {
		t.Fatalf("GetName() = %q, %v", got, err)
	}
	if got, err := c.GetDescription(); err != nil || got != "Run the app" {
		t.Fatalf("GetDescription() = %q, %v", got, err)
	}
}

func TestCommandAdditionalNames(t *testing.T) {
	c := &Command{}
	if got, err := c.GetAdditionalNames(); err != nil || got != nil {
		t.Fatalf("GetAdditionalNames() = %v, %v", got, err)
	}
	if err := c.AddAdditionalName("r"); !errors.Is(err, ErrNilArray) {
		t.Fatalf("AddAdditionalName error = %v", err)
	}
	if err := c.SetAdditionalNames([]string{}); err != nil {
		t.Fatal(err)
	}
	if err := c.AddAdditionalName("r"); err != nil {
		t.Fatal(err)
	}
	if err := c.AddAdditionalName("start"); err != nil {
		t.Fatal(err)
	}
	if err := c.AddAdditionalName(""); err == nil {
		t.Fatal("empty alias should error")
	}
	if err := c.SetAdditionalNames(nil); !errors.Is(err, ErrNilArray) {
		t.Fatalf("SetAdditionalNames error = %v", err)
	}
	if got, err := c.GetAdditionalNames(); err != nil || !slices.Equal(got, []string{"r", "start"}) {
		t.Fatalf("aliases = %v, %v", got, err)
	}
	if err := c.SetAdditionalNames([]string{"go"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := c.GetAdditionalNames(); !slices.Equal(got, []string{"go"}) {
		t.Fatalf("replacement aliases = %v", got)
	}
}

func TestCommandFlagAccessors(t *testing.T) {
	c := &Command{}
	fn := func(ctx *Context) error { ctx.Values["called"] = true; return nil }
	if got, err := c.GetFlags(); got != nil || !errors.Is(err, ErrNilFlags) {
		t.Fatalf("GetFlags() = %v, %v", got, err)
	}
	if err := c.AddFlag("--verbose", fn); !errors.Is(err, ErrNilFlags) {
		t.Fatalf("AddFlag error = %v", err)
	}
	if err := c.SetFlags(Flags{}); err != nil {
		t.Fatal(err)
	}
	if err := c.AddFlag("", fn); err == nil {
		t.Fatal("empty key should error")
	}
	if err := c.AddFlag("--nil", nil); err == nil {
		t.Fatal("nil callback should error")
	}
	if len(c.Flags) != 0 {
		t.Fatal("failed additions modified flags")
	}
	if err := c.AddFlag("--verbose", fn); err != nil {
		t.Fatal(err)
	}
	if err := c.SetFlags(nil); !errors.Is(err, ErrNilFlags) {
		t.Fatalf("SetFlags error = %v", err)
	}
	got, err := c.GetFlags()
	if err != nil || len(got) != 1 || got["--verbose"] == nil {
		t.Fatalf("flags = %v, error = %v", got, err)
	}
	ctx := &Context{Values: map[string]any{}}
	if err := got["--verbose"](ctx); err != nil || ctx.Values["called"] != true {
		t.Fatalf("stored callback failed: %v", err)
	}
	replacement := func(ctx *Context) error { ctx.Values["called"] = false; return nil }
	if err := c.AddFlag("--verbose", replacement); err != nil {
		t.Fatal(err)
	}
	if err := c.Flags["--verbose"](ctx); err != nil || ctx.Values["called"] != false {
		t.Fatalf("replacement callback failed: %v", err)
	}
	if err := c.SetFlags(Flags{}); err != nil || len(c.Flags) != 0 {
		t.Fatalf("SetFlags did not replace map: %v", err)
	}
}

func TestCommandValidate(t *testing.T) {
	for _, tt := range []struct {
		name string
		cmd  *Command
		want error
	}{
		{"nil", nil, ErrNilCommand},
		{"empty name", &Command{}, ErrEmptyCommandName},
		{"missing execute", &Command{Name: "run"}, ErrNilCommandFunction},
		{"valid minimal", &Command{Name: "run", Execute: func(*Context) error { return nil }}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cmd.validate(); !errors.Is(err, tt.want) {
				t.Fatalf("validate() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCommandParseFlags(t *testing.T) {
	var calls []string
	c := &Command{Flags: Flags{
		"--one": func(*Context) error { calls = append(calls, "one"); return nil },
		"-t":    func(*Context) error { calls = append(calls, "two"); return nil },
	}}
	functions, unknown, err := c.parseFlags([]string{"value", "--one", "--unknown", "-t", "--one"})
	if err != nil || !slices.Equal(unknown, []string{"--unknown"}) || len(functions) != 3 {
		t.Fatalf("parseFlags: functions=%d unknown=%v error=%v", len(functions), unknown, err)
	}
	if len(calls) != 0 {
		t.Fatal("parseFlags executed callbacks")
	}
	for _, fn := range functions {
		if err := fn(&Context{}); err != nil {
			t.Fatal(err)
		}
	}
	if !slices.Equal(calls, []string{"one", "two", "one"}) {
		t.Fatalf("callback order = %v", calls)
	}
	functions, unknown, err = c.parseFlags(nil)
	if err != nil || len(functions) != 0 || len(unknown) != 0 {
		t.Fatal("empty input should produce no flags")
	}
	t.Run("unknown short flag", func(t *testing.T) {
		_, unknown, err := c.parseFlags([]string{"-x"})
		if err != nil || !slices.Equal(unknown, []string{"-x"}) {
			t.Fatalf("unknown = %v, error = %v; want [-x], nil", unknown, err)
		}
	})
}

func TestCommandExecuteFlags(t *testing.T) {
	c := &Command{}
	ctx := &Context{}
	stop := errors.New("stop")
	calls := 0
	functions := []FlagFunction{
		func(got *Context) error {
			calls++
			if got != ctx {
				t.Error("wrong context")
			}
			return nil
		},
		func(*Context) error { calls++; return stop },
		func(*Context) error { t.Error("executed after error"); return nil },
	}
	if err := c.executeFlags(ctx, functions); !errors.Is(err, stop) {
		t.Fatalf("executeFlags error = %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if err := c.executeFlags(ctx, nil); err != nil {
		t.Fatal(err)
	}
}

func TestCommandRunFlags(t *testing.T) {
	stop := errors.New("flag error")
	for _, tt := range []struct {
		name         string
		args         []string
		wantCalls    int
		wantErr      bool
		wantSentinel error
	}{
		{name: "empty"},
		{name: "positional", args: []string{"example"}},
		{name: "known", args: []string{"--ok", "--ok"}, wantCalls: 2},
		{name: "unknown prevents callbacks", args: []string{"--ok", "--missing"}, wantErr: true},
		{name: "callback error stops execution", args: []string{"--fail", "--ok"}, wantCalls: 1, wantErr: true, wantSentinel: stop},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			ctx := &Context{}
			c := &Command{Flags: Flags{
				"--ok": func(got *Context) error {
					calls++
					if got != ctx {
						t.Error("wrong context")
					}
					return nil
				},
				"--fail": func(*Context) error { calls++; return stop },
			}}
			err := c.RunFlags(ctx, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantError = %v", err, tt.wantErr)
			}
			if tt.wantSentinel != nil && !errors.Is(err, tt.wantSentinel) {
				t.Errorf("error = %v, want %v", err, tt.wantSentinel)
			}
			if calls != tt.wantCalls {
				t.Errorf("calls = %d, want %d", calls, tt.wantCalls)
			}
		})
	}
}
