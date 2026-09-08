package argbin

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNilCommand = errors.New("command cannot be nil")
	ErrEmptyCommandName = errors.New("command name cannot be empty")
	ErrNilArray = errors.New("array cannot be nil")
	ErrNilMap = errors.New("map cannot be nil")
	ErrNilFlags = errors.New("flags cannot be nil")
)

// Command is the storage where you'll input all of your command information.
type Command struct {
	// Name of the command that'll be ran
	Name string

	// Additional names is where aliases are stored for a command.
	AdditionalNames []string

	// Flags for Command are stored as k,v pairs as a string and a function.
	Flags Flags

	// TakesValue scans subsequent arguments to find a valid value.
	TakesValue bool

	// Execute runs the commands given function.
	Execute func(ctx *Context) error 
}

type Flags map[string]FlagFunction
type FlagFunction func(ctx *Context) error

func (c *Command) GetName() (string, error) {
	if c == nil {
		return "", ErrNilCommand
	}
	return c.Name, nil
}

func (c *Command) SetName(name string) error {
	if c == nil {
		return ErrNilCommand
	}

	if name == "" {
		return ErrEmptyCommandName
	}

	c.Name = name
	return nil
}

func (c *Command) GetAdditionalNames() ([]string, error) {
	if c == nil {
		return nil, ErrNilCommand
	}
	return c.AdditionalNames, nil
}

func (c *Command) SetAdditionalNames(names []string) error {
	if c == nil {
		return ErrNilCommand
	}

	if names == nil {
		return fmt.Errorf("string %w", ErrNilArray)
	}

	c.AdditionalNames = names
	return nil
}

func (c *Command) AddAdditionalName(name string) error {
	if c == nil {
		return ErrNilCommand
	}

	if name == "" {
		return fmt.Errorf("name argument cannot be empty")
	}

	if c.AdditionalNames == nil {
		return fmt.Errorf("additional names %w.", ErrNilArray)
	}

	c.AdditionalNames = append(c.AdditionalNames, name)
	return nil
}

func (c *Command) GetFlags() (Flags, error) {
	if c == nil {
		return nil, ErrNilCommand
	}

	if c.Flags == nil {
		return nil, ErrNilFlags
	}

	return c.Flags, nil
}

func (c *Command) SetFlags(flags Flags) error {
	if c == nil {
		return ErrNilCommand
	}

	if flags == nil {
		return ErrNilFlags
	}

	c.Flags = flags

	return nil
}

func (c *Command) AddFlag(key string, flagFunction FlagFunction) error {
	if c == nil {
		return ErrNilCommand
	}

	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if flagFunction == nil {
		return fmt.Errorf("flagFunction cannot be empty")
	}

	if c.Flags == nil {
		return ErrNilFlags
	}

	c.Flags[key] = flagFunction
	return nil
}

func (c *Command) parseFlags(args []string) ([]FlagFunction, []string, error) {

	var parsedOutput []FlagFunction
	var unknownFlags []string
	
	for _, arg := range args {
		if value, ok := c.Flags[arg]; ok {
			parsedOutput = append(parsedOutput, value)
		} else {
			unknownFlags = append(unknownFlags, arg)
		}
	}

	return parsedOutput, unknownFlags, nil
}

func (c *Command) executeFlags(ctx *Context, flags []FlagFunction) error {
	for _, single := range flags {
		if err := single(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) RunFlags(ctx *Context, args []string) error {
	functions, unknown, err := c.parseFlags(args)

	if err != nil {
		return err
	}


	if len(unknown) > 0 {
		return fmt.Errorf("unknown arguments: %v", strings.Join(unknown, " "))
	}
	
	if err := c.executeFlags(ctx, functions); err != nil {
		return err
	}

	return nil
}

func (c *Command) validate() error {
	if c == nil {
		return ErrNilCommand
	}

	if c.Name == "" {
		return ErrEmptyCommandName
	}

	if c.Execute == nil {
		return ErrNilCommandFunction
	}
	
	return nil
}

