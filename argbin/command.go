package argbin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Command is the storage where you'll input all of your command information.
type Command struct {
	// Name of the command that'll be ran
	Name string

	// Description is your Command help menu
	Description string

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

// GetName returns the commands Name.
func (c *Command) GetName() (string, error) {

	// Checks if the command is nil
	if c == nil {
		return "", ErrNilCommand
	}
	
	// Checks if the Name is empty
	if c.Name == "" {
		return "", fmt.Errorf("cannot get empty command name")
	}

	// Returns the commands Name
	return c.Name, nil
}

// SetName sets the name of a command.
func (c *Command) SetName(name string) error {

	// Checks if the command is nil
	if c == nil {
		return ErrNilCommand
	}
	
	// Checks if the name argument is empty
	if name == "" {
		return ErrEmptyCommandName
	}

	// Assigns name to the command Name
	c.Name = name
	return nil
}

// GetDescription returns the commands Description.
func (c *Command) GetDescription() (string, error) {

	// Checks if a command is nil
	if c == nil {
		return "", ErrNilCommand
	}

	// Checks if the command descript is empty
	if c.Description == "" {
		return "", fmt.Errorf("command description cannot be blank")
	}

	// Returns the commands Description
	return c.Description, nil
}

// SetDescription sets the Description of a command.
func (c *Command) SetDescription(des string) error {

	// Checks if a command is nil
	if c == nil {
		return ErrNilCommand
	}

	// Checks if the des argument is empty
	if des == "" {
		return fmt.Errorf("des cannot be blank when setting description")
	}
	
	// Assigns the des argument to commands Description
	c.Description = des
	return nil
}

// GetAdditionalNames returns all of a commands additional names into a string array.
func (c *Command) GetAdditionalNames() ([]string, error) {

	// Checks if command is nil
	if c == nil {
		return nil, ErrNilCommand
	}
	
	// Checks if AdditionalNames is nil
	if c.AdditionalNames == nil {
		return nil, fmt.Errorf("cannot get empty additional names list")
	}

	// Returns commands AdditionalNames
	return c.AdditionalNames, nil
}

// SetAdditionalNames sets the additional names of a command. 
func (c *Command) SetAdditionalNames(names []string) error {
	// Checks if a command is nil
	if c == nil {
		return ErrNilCommand
	}
	
	// Checks if the names argument is nil
	if names == nil {
		return fmt.Errorf("string %w", ErrNilArray)
	}

	c.AdditionalNames = names
	return nil
}

// AddAdditionalName adds a single name to the AdditionalNames array.
func (c *Command) AddAdditionalName(name string) error {

	// Checks if command is nil
	if c == nil {
		return ErrNilCommand
	}

	// Checks if the name argument is empty
	if name == "" {
		return fmt.Errorf("name argument cannot be empty")
	}

	// Checks if AdditionalNames is nil
	if c.AdditionalNames == nil {
		// If it doesn't exist, it creates the array.
		c.AdditionalNames = make([]string, 0)
	}

	// Appends the new name to the AdditionalNames array
	c.AdditionalNames = append(c.AdditionalNames, name)
	return nil
}

// GetFlags returns the map stored in c.Flags.
func (c *Command) GetFlags() (Flags, error) {

	// Checks if command is nil
	if c == nil {
		return nil, ErrNilCommand
	}

	// Checks if Flags is nil
	if c.Flags == nil {
		return nil, ErrNilFlags
	}

	// Return flag map
	return c.Flags, nil
}

// SetFlags sets the flags in a Command with a given flags argument.
func (c *Command) SetFlags(flags Flags) error {

	// Checks if command is nil
	if c == nil {
		return ErrNilCommand
	}
	
	// Checks if the flags argument is nil
	if flags == nil {
		return ErrNilFlags
	}

	// Checks if the c.Flags is nil, creates a map if it is nil
	if c.Flags == nil {
		c.Flags = make(Flags)
	}

	// Sets c.Flags to the flags argument
	c.Flags = flags
	return nil
}

// AddFlag adds a flag to c.Flags map. A Flag is consistant of a key as the identifier, and a flagFunction that stores the execution of the flag.
func (c *Command) AddFlag(key string, flagFunction FlagFunction) error {

	// Checks if the command is nil
	if c == nil {
		return ErrNilCommand
	}
	
	// Checks if the key argument is empty
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	
	// Checks if the flagFunction argument is empty
	if flagFunction == nil {
		return fmt.Errorf("flagFunction cannot be empty")
	}
	
	// Checks if c.Flags is nil
	if c.Flags == nil {
		c.Flags = make(Flags)
	}

	// Adds new flag to flag map
	c.Flags[key] = flagFunction
	return nil
}

// parseFlags reads through a commands Flags map and finds valid flags that are called through user arguments.
func (c *Command) parseFlags(args []string) ([]FlagFunction, []string, error) {
	
	// Stores all gathered flags from args
	var parsedOutput []FlagFunction

	// Stores unknown flags that we're input
	var unknownFlags []string
	
	// Loops over each argument
	for _, arg := range args {

		// If a flag is found, it's added to parsedOutput
		if value, ok := c.Flags[arg]; ok {
			parsedOutput = append(parsedOutput, value)
			continue
		}
		
		// If the argument doesn't have a dash or double dashes, it's skipped as it's a potential value for commands or flags.
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		
		// Parses the argument to see if it's a numeric value
		_, err := strconv.ParseFloat(arg, 64)
		if err == nil || errors.Is(err, strconv.ErrRange) {
			continue
		}
		
		// Unknown flags are added to unknownFlags
		unknownFlags = append(unknownFlags, arg)
	}
	
	return parsedOutput, unknownFlags, nil
}

// executeFlags runs each flags related function. Valid flags are gathered from parseFlags().
func (c *Command) executeFlags(ctx *Context, flags []FlagFunction) error {
	// Loops over all flag functions
	for _, single := range flags {
		// Executes the flag function and errors out if it can't execute
		if err := single(ctx); err != nil {
			return err
		}
	}
	return nil
}

// runFlags handles the parsing and execution of flags from given arguments.
func (c *Command) runFlags(ctx *Context, args []string) error {

	// Gather functions, unknown flags, and error
	functions, unknown, err := c.parseFlags(args)
	
	// Error check for parseFlags
	if err != nil {
		return err
	}

	// Errors if unknown arguments are input
	if len(unknown) > 0 {
		return fmt.Errorf("unknown arguments: %v", strings.Join(unknown, " "))
	}
	
	// Executes flags
	if err := c.executeFlags(ctx, functions); err != nil {
		return err
	}

	return nil
}

// validate checks if a Command is valid for argbin.
func (c *Command) validate() error {

	// Checks if the command is nil
	if c == nil {
		return ErrNilCommand
	}
	
	// Checks if the name is empty
	if c.Name == "" {
		return ErrEmptyCommandName
	}
	
	// Checks if the command execution is nil
	if c.Execute == nil {
		return ErrNilCommandFunction
	}
	
	return nil
}

