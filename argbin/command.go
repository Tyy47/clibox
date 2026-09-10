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

	// Flags for Command are stored as k,v pairs as a string and a Flag object.
	Flags Flags

	// TakesValue scans subsequent arguments to find a valid value.
	TakesValue bool

	// Execute runs the commands given function.
	Execute func(ctx *Context) error 
}

type Flags map[string]*Flag
type parsedFlag struct {
	flag *Flag
	value string
}

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

// AddFlag adds a new flag to Commands Flag map. key is the identifier & name of a flag. flag is the Flag object that'll be stored as a value.
func (c *Command) AddFlag(key string, flag *Flag) error {

	// Checks if the command is nil
	if c == nil {
		return ErrNilCommand
	}
	
	// Checks if the key argument is empty
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	
	// Checks if the flagFunction argument is empty
	if flag == nil {
		return fmt.Errorf("flag cannot be empty")
	}
	
	// Checks if c.Flags is nil
	if c.Flags == nil {
		c.Flags = make(Flags)
	}

	// Adds new flag to flag map
	c.Flags[key] = flag
	return nil
}

// parseFlags reads through a commands Flags map and finds valid flags that are called through user arguments.
func (c *Command) parseFlags(args []string) ([]parsedFlag, []string, error) {
	
	// Stores all gathered flags from args
	var parsed []parsedFlag

	// Stores unknown flags that we're input
	var unknownFlags []string
	
	// Loops over each argument
	for i, arg := range args {

		// If a flag is found, it's added to parsedOutput
		if flag, ok := c.Flags[arg]; ok {

			// Errors out if there is no flag
			if flag == nil {
				return nil, nil, fmt.Errorf("flag %s cannot be nil", arg)
			}
			
			// Create flag for parsed flag array
			entry := parsedFlag{flag: flag}
	
			// Checks if the flag takes in a value
			if flag.TakesValue {
				if i+1 >= len(args) {
					// Returns an error if no value is given
					return nil, nil, fmt.Errorf("flag %s requires a value", arg)
				}
				
				// Increment and assign argument to flag value
				i++
				if !strings.HasPrefix(args[i], "-") {
					entry.value = args[i]
				}
			}
			
			// Add parsed flag to parsed array
			parsed = append(parsed, entry)
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
	
	return parsed, unknownFlags, nil
}

// executeFlags runs each flags related function. Valid flags are gathered from parseFlags().
func (c *Command) executeFlags(ctx *Context, flags []parsedFlag) error {
	// Loop through each parsed flag
	for _, entry := range flags {
		// Assign flag value to context
		ctx.ParsedFlagValue = entry.value
	
		// Validate flag
		if err := entry.flag.validate(); err != nil {
			return err
		}
		
		// Execute flag
		if err := entry.flag.Execute(ctx); err != nil {
			return err
		}

	}
	return nil
}

// runFlags handles the parsing and execution of flags from given arguments.
func (c *Command) runFlags(ctx *Context, args []string) error {

	// Gather functions, unknown flags, and error
	flags, unknown, err := c.parseFlags(args)
	
	// Error check for parseFlags
	if err != nil {
		return err
	}

	// Errors if unknown arguments are input
	if len(unknown) > 0 {
		return fmt.Errorf("unknown arguments: %v", strings.Join(unknown, " "))
	}
	
	return c.executeFlags(ctx, flags)
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

