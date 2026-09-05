package argbin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Tyy47/clibox/internal/utils"
)

type CommandList map[string]*Command

type Root struct {
	AppName     string
	Description string
	Commands    CommandList
}

type Command struct {
	Name        string
	Description string
	Flags       map[string]*Flag
	Execute     func(ctx *Context) error
}

type Flag struct {
	Name        string // logical name: "verbose"
	Short       string // optional short hand: "v"
	Description string
	Execute     func(ctx *Context) error
	TakesValue  bool
}

type Context struct {
	Command *Command
	Values  map[string]any
	Args    []string
}

type parsedFlag struct {
	flag  *Flag
	value *string
}

type parsedInput struct {
	flags []parsedFlag
	args  []string
}

// A collection of errors for argbin
var (
	ErrMissingCommand      = errors.New("no arguments provided")
	ErrUnknownCommand      = errors.New("unknown command")
	ErrNilRoot             = errors.New("root cannot be nil")
	ErrNilCommand          = errors.New("command cannot be nil")
	ErrNilFlag             = errors.New("flag cannot be nil")
	ErrNilContext          = errors.New("context cannot be nil")
	ErrUnknownFlag         = errors.New("unknown flag")
	ErrDuplicateFlag       = errors.New("duplicate flag")
	ErrMissingFlagValue    = errors.New("missing flag value")
	ErrUnexpectedFlagValue = errors.New("unexpected flag value")
)

// Roots validation method
func (r *Root) validate() error {
	if r == nil {
		return ErrNilRoot
	}

	if r.AppName == "" {
		return fmt.Errorf("root appname cannot be empty")
	}

	if r.Description == "" {
		return fmt.Errorf("root description cannot be empty")
	}

	// Validate users created key maps for commands to make sure names and keys match
	for key, command := range r.Commands {
		if err := command.validate(); err != nil {
			return fmt.Errorf("command %q: %w", key, err)
		}
		if key != command.Name {
			return fmt.Errorf("command key %q does not match command name %q", key, command.Name)
		}
	}

	return nil
}

func (c *Command) validate() error {
	if err := validCommandChecker(c); err != nil {
		return err
	}

	shorts := make(map[string]string)

	for key, flag := range c.Flags {
		if err := validFlagChecker(flag); err != nil {
			return fmt.Errorf("flag %q: %w", key, err)
		}
		if key != flag.Name {
			return fmt.Errorf("flag key %q does not match flag name %q", key, flag.Name)
		}
		if flag.Short != "" {
			if existing, ok := shorts[flag.Short]; ok {
				return fmt.Errorf("flag short %q is used by both %q and %q", flag.Short, existing, flag.Name)
			}
			shorts[flag.Short] = flag.Name
		}
	}
	return nil
}

func (c *Command) searchFlag(name string) (*Flag, bool) {
	if c == nil {
		return nil, false
	}

	flag, ok := c.Flags[name]
	return flag, ok
}

func (c *Command) searchShortFlag(short string) (*Flag, bool) {
	if c == nil {
		return nil, false
	}

	for _, flag := range c.Flags {
		if flag == nil {
			continue
		}

		if flag.Short == short {
			return flag, true
		}
	}

	return nil, false
}

func (c *Command) resolveFlagToken(token string) (*Flag, error) {
	if strings.HasPrefix(token, "--") {
		name := strings.TrimPrefix(token, "--")
		if name == "" {
			return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
		}

		flag, ok := c.searchFlag(name)
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
		}
		return flag, nil
	}

	if strings.HasPrefix(token, "-") {
		short := strings.TrimPrefix(token, "-")
		if len(short) != 1 {
			return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
		}

		flag, ok := c.searchShortFlag(short)
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
		}
		return flag, nil
	}

	return nil, fmt.Errorf("%w: %q", ErrUnknownFlag, token)
}

func (c *Command) parseInput(args []string) (*parsedInput, error) {
	if c == nil {
		return nil, ErrNilCommand
	}

	parsed := &parsedInput{}
	seen := make(map[string]struct{})

	for i := 0; i < len(args); i++ {
		token := args[i]

		if token == "--" {
			parsed.args = append(parsed.args, args[i+1:]...)
			break
		}

		if token == "" {
			parsed.args = append(parsed.args, token)
			continue
		}

		if token == "-" || !strings.HasPrefix(token, "-") {
			parsed.args = append(parsed.args, token)
			continue
		}

		if strings.Contains(token, "=") {
			return nil, fmt.Errorf("%w: %q", ErrUnexpectedFlagValue, token)
		}

		flag, err := c.resolveFlagToken(token)
		if err != nil {
			return nil, err
		}

		if _, ok := seen[flag.Name]; ok {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateFlag, flag.Name)
		}
		seen[flag.Name] = struct{}{}

		if !flag.TakesValue {
			parsed.flags = append(parsed.flags, parsedFlag{flag: flag})
			continue
		}

		if i+1 >= len(args) || args[i+1] == "--" || strings.HasPrefix(args[i+1], "-") {
			return nil, fmt.Errorf("%w: %s", ErrMissingFlagValue, flag.Name)
		}

		i++
		value := args[i]
		parsed.flags = append(parsed.flags, parsedFlag{flag: flag, value: &value})
	}

	return parsed, nil
}

func (c *Command) parseFlags(ctx *Context, args []string) error {
	if c == nil {
		return ErrNilCommand
	}

	if ctx == nil {
		return ErrNilContext
	}

	parsed, err := c.parseInput(args)
	if err != nil {
		return err
	}

	return c.applyParsed(ctx, parsed)
}

func (c *Command) applyParsed(ctx *Context, parsed *parsedInput) error {
	if c == nil {
		return ErrNilCommand
	}
	if ctx == nil {
		return ErrNilContext
	}
	if parsed == nil {
		return nil
	}
	if ctx.Values == nil {
		ctx.Values = make(map[string]any)
	}

	ctx.Args = append(ctx.Args, parsed.args...)

	for _, occurrence := range parsed.flags {
		if occurrence.value == nil {
			ctx.Values[occurrence.flag.Name] = true
		} else {
			ctx.Values[occurrence.flag.Name] = *occurrence.value
		}

		if err := occurrence.flag.Execute(ctx); err != nil {
			return fmt.Errorf("flag %q cannot be executed: %w", occurrence.flag.Name, err)
		}
	}

	return nil
}

// validCommandChecker checks if a given command is valid to work with argbins parser
func validCommandChecker(commands ...*Command) error {
	for _, invalid := range commands {
		if invalid == nil {
			return ErrNilCommand
		}
	}

	for _, single := range commands {
		if single.Name == "" {
			return fmt.Errorf("Command name member can't be blank. Command: %v", single)
		}

		if strings.HasPrefix(single.Name, "-") || strings.Contains(single.Name, "=") || strings.ContainsAny(single.Name, " \t\n") {
			return fmt.Errorf("Command name must not start with a dash or contain equals or whitespace: %v", single)
		}

		if single.Description == "" {
			return fmt.Errorf("Command description member can't be blank. Command: %v", single)
		}

		if single.Execute == nil {
			return fmt.Errorf("Command execute member can't be blank. Command: %v", single)
		}
	}

	return nil
}

// validFlagChecker checks if a given flag is valid to work with argbins parser
func validFlagChecker(flag *Flag) error {
	if flag == nil {
		return ErrNilFlag
	}

	if flag.Name == "" {
		return fmt.Errorf("Flag name member can't be empty: %v", flag)
	}

	if strings.HasPrefix(flag.Name, "-") || strings.Contains(flag.Name, "=") || strings.ContainsAny(flag.Name, " \t\n") {
		return fmt.Errorf("Flag name must be a logical name without a leading dash, equals, or whitespace: %v", flag)
	}

	if flag.Short != "" {
		if len(flag.Short) != 1 {
			return fmt.Errorf("Flag short member must be one character: %v", flag)
		}
		if strings.Contains(flag.Short, "-") {
			return fmt.Errorf("Flag short member cannot contain '-': %v", flag)
		}
	}

	if flag.Description == "" {
		return fmt.Errorf("Flag description member can't be empty: %v", flag)
	}

	if flag.Execute == nil {
		return fmt.Errorf("Flag execute member can't be empty: %v", flag)
	}

	return nil
}

func (r *Root) AddCommand(command *Command) error {
	if r == nil {
		return ErrNilRoot
	}

	if err := validCommandChecker(command); err != nil {
		return err
	}

	if r.Commands == nil {
		r.Commands = make(CommandList)
	}

	for takenName := range r.Commands {
		if command.Name == takenName {
			return fmt.Errorf("Command name %s already exists in command list: %v", command.Name, r.Commands)
		}
	}
	r.Commands[command.Name] = command
	return nil
}

func (c *Command) AddFlag(flag *Flag) error {
	if c == nil {
		return ErrNilCommand
	}

	if err := validFlagChecker(flag); err != nil {
		return err
	}

	if c.Flags == nil {
		c.Flags = make(map[string]*Flag)
	}

	for takenName, takenFlag := range c.Flags {
		if takenName == flag.Name {
			return fmt.Errorf("Flag name: %s is already taken by %s", takenName, flag.Name)
		}
		if flag.Short != "" && takenFlag != nil && takenFlag.Short == flag.Short {
			return fmt.Errorf("Flag short: %s is already taken by %s", flag.Short, takenFlag.Name)
		}
	}

	c.Flags[flag.Name] = flag
	return nil
}

func (r *Root) Run() error {
	// Checks if Root is nil
	if r == nil {
		return ErrNilRoot
	}

	// Validates root before fully executing to make sure the struct is valid
	if err := r.validate(); err != nil {
		return err
	}

	// Users arguments
	args := *utils.GetArgs()

	if len(args) == 0 {
		return ErrMissingCommand
	}

	command, ok := r.Commands[args[0]]
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownCommand, args[0])
	}

	ctx := &Context{
		Command: command,
		Values:  make(map[string]any),
	}

	if err := command.parseFlags(ctx, args[1:]); err != nil {
		return err
	}

	if err := command.Execute(ctx); err != nil {
		return fmt.Errorf("command cannot be executed: %w", err)
	}

	return nil
}
