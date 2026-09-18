package argbin

import (
	"errors"
	"fmt"
	"slices"
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

	// Subcommands stores commands for branching commands
	Subcommands []*Command

	// isSubcommand is an internal field to mark if this command is a subcommand
	isSubcommand bool
}

type (
	Flags      map[string]*Flag
	parsedFlag struct {
		flag  *Flag
		value string
	}
)

// GetName returns the commands Name.
func (c *Command) GetName() (string, error) {
	// Checks if the command is nil
	if c == nil {
		return "", ErrNilCommand
	}

	// Checks if the Name is empty
	if c.Name == "" {
		return "", ErrEmptyCommandName
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
		return "", ErrEmptyDescription
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
		return fmt.Errorf("des %w", ErrEmptyArgument)
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
		return nil, ErrEmptyAliasList
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
		return fmt.Errorf("name %w", ErrEmptyArgument)
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
		return fmt.Errorf("key %w", ErrEmptyArgument)
	}

	// Checks if the flagFunction argument is empty
	if flag == nil {
		return ErrNilFlag
	}

	// Checks if c.Flags is nil
	if c.Flags == nil {
		c.Flags = make(Flags)
	}

	// Adds new flag to flag map
	c.Flags[key] = flag
	return nil
}

// AddSubcommands adds given commands to the Subcommands array.
func (c *Command) AddSubcommands(cmds ...*Command) error {
	if c == nil {
		return ErrNilCommand
	}

	if len(c.Subcommands) == 0 {
		c.Subcommands = make([]*Command, 0)
	}

	for _, cmd := range cmds {
		if cmd == nil {
			return ErrNilCommand
		}

		c.Subcommands = append(c.Subcommands, cmd)
	}

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
				return nil, nil, fmt.Errorf("%s %w", arg, ErrNilFlag)
			}

			// Create flag for parsed flag array
			entry := parsedFlag{flag: flag}

			// Checks if the flag takes in a value
			if flag.TakesValue {
				if i+1 >= len(args) {
					// Returns an error if no value is given
					return nil, nil, fmt.Errorf("%s %w", arg, ErrEmptyFlagValue)
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
		return fmt.Errorf("%w %v", ErrUnknownArguments, strings.Join(unknown, " "))
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

	// Checks if the command execution & subcommands are nil
	if c.Execute == nil && len(c.Subcommands) == 0 {
		return ErrNilCommandFunction
	}

	if c.Execute != nil && len(c.Subcommands) > 0 {
		return ErrExecuteSubcommandsToggled
	}

	return nil
}

// parseSubcommand finds a subcommand based on the given args. Returns the subcommand if found, if not, it'll return the original command and an error.
func (c *Command) parseSubcommand(args []string, ctx *Context) (*Command, error) {
	// Check to see if the command is valid
	if err := c.validate(); err != nil {
		return nil, err
	}

	// Checks if a command is a retainer for Subcommands
	if len(args) == 0 || len(c.Subcommands) == 0 {
		return c, nil
	}

	// Loop over each subcommand stored in a Command
	for _, child := range c.Subcommands {
		// Checks if a subcommand is nil
		if child == nil {
			return nil, ErrNilCommand
		}
		
		// Checks if an arg is a subcommand
		if child.Name == args[0] || slices.Contains(child.AdditionalNames, args[0]) {
			// Matched a child; recurse with the remaining arguments.
			if err := child.validate(); err != nil {
				return nil, err
			}
			// Increment argument count for later incrementing arguments
			ctx.argCount += 1
			return child.parseSubcommand(args[1:], ctx)
		}
	}
	
	// Returns the original command and an unknown subcommand error
	return c, fmt.Errorf("%w %s", ErrUnknownSubcommand, args[0])
}

// checkSubcommands checks if a given subcommand is valid. If it is valid, it will return the subcommand, if not, it will return the original command.
func checkSubcommands(cmd *Command, ctx *Context) (*Command, error) {
	// Check to see if given command is valid
	if err := cmd.validate(); err != nil {
		return nil, err
	}

	// Checks if a command execution is nil
	if cmd.Execute != nil {
		return cmd, nil
	}

	// Checks if a given command is a retainer for more subcommands
	if cmd.Subcommands != nil || len(cmd.Subcommands) >= 1 {
		subCmd, err := cmd.parseSubcommand(ctx.Args[1:], ctx)
		if err != nil {
			return cmd, fmt.Errorf("%w %s", ErrUnknownCommand, ctx.Args[1])
		}

		ctx.Command = subCmd
		return subCmd, nil
	}

	return cmd, nil
}

// validateSubCommand checks if the given subCmd is a valid command that can be ran through argbin.
func validateSubCommand(subCmd *Command) error {
	// Checks if the cmd is nil
	if subCmd == nil {
		return ErrNilCommand
	}
	
	// Runs internal validate to check if the command is valid
	if err := subCmd.validate(); err != nil {
		return err
	}
	
	// If the command has subcommands, it'll return an error stating the command is missing arguments through ErrSubcommandMissingArgs.
	if len(subCmd.Subcommands) >= 1 {
		return fmt.Errorf("%s %w", subCmd.Name, ErrSubcommandMissingArgs)
	}
	
	// Internal marker to track subcommands
	subCmd.isSubcommand = true

	return nil
}
