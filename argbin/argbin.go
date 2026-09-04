package argbin

import (
	"errors"
	"fmt"

	"github.com/Tyy47/clibox/internal/utils"
)

type CommandList map[string]*Command

type Root struct {
	AppName string
	Description string	
	Commands CommandList
}

type Command struct {
	Name string
	Description string
	Flags map[string]*Flag
	Execute func(ctx *Context) error
}

type Flag struct {
	Name string
	Description string
	Execute func(ctx *Context) error
	TakesValue bool
}

type Context struct {
	Command *Command
	Values map[string]any
}

// A collection of errors for argbin
var (
	ErrMissingCommand = errors.New("no arguments provided")
	ErrUnknownCommand = errors.New("unknown command")
	ErrNilRoot = errors.New("root cannot be nil")
	ErrNilCommand = errors.New("command cannot be nil")
	ErrNilFlag = errors.New("flag cannot be nil")
	ErrNilContext = errors.New("context cannot be nil")
)

// Roots validation method
func (r *Root) validate() error {
	if r.AppName == "" {
		return fmt.Errorf("root appname cannot be empty")
	}

	if r.Description == "" {
		return fmt.Errorf("root description cannot be empty")
	}

	if r.Commands == nil {
		r.Commands = make(CommandList)
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

    for key, flag := range c.Flags {
        if err := validFlagChecker(flag); err != nil {
            return fmt.Errorf("flag %q: %w", key, err)
        }
        if key != flag.Name {
            return fmt.Errorf("flag key %q does not match flag name %q", key, flag.Name)
        }
    }
    return nil
}

func (c *Command) searchFlag(name string) (*Flag, bool) {
	
	if c == nil {
		return nil, false
	}

	for key, flag := range c.Flags {
		if flag == nil {
			return nil, false
		}
		
		if key == name || flag.Name == name {
			return flag, true
		}
	}


	return nil, false
}

func (c *Command) parseFlags(ctx *Context, args []string) error {
	
	if c == nil {
		return ErrNilCommand
	}

	if ctx.Values == nil {
		return ErrNilContext
	}

	for _, arg := range args {
		flag, ok := c.searchFlag(arg)
		if !ok {
			continue
		}

		if err := validFlagChecker(flag); err != nil {
			return err
		}
		
		if flag.Execute == nil {
			continue
		}

		if err := flag.Execute(ctx); err != nil {
			return err
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

	for takenFlag := range c.Flags {
		if takenFlag == flag.Name {
			return fmt.Errorf("Flag name: %s is already taken by %s", takenFlag, flag.Name)
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
		Values: make(map[string]any),
	}

	if err := command.parseFlags(ctx, args[1:]); err != nil {
		return err
	}

	if err := command.Execute(ctx); err != nil {
		return fmt.Errorf("command cannot be executed: %w", err)
	}

	return nil
}




