package argbin

import (
	"slices"
)

// Root is the application structure where all data is held.
type Root struct {
	// AppName stores the name of your application.
	AppName string

	// AppVersion stores the version of your application.
	AppVersion string

	// HelpMenu is your Apps help menu.
	HelpMenu string

	// CommandList stores all the requires commands for your application.
	CommandList []*Command
}


// AddCommand takes in a Command pointer and adds it to the Roots CommandList
func (r *Root) AddCommand(command ...*Command) error {
	// Checks if the root object is nil
	if r == nil {
		return ErrNilRoot
	}

	// Checks if the CommandList is nil
	if r.CommandList == nil {
		r.CommandList = make([]*Command, 0)
	}

	// Validates the commands
	for _, cmd := range command {
		if err := cmd.validate(); err != nil {
			return err
		}

		// Searches command to make sure their is no duplicate names in list
		for _, prev := range r.CommandList {
			if cmd.Name == prev.Name {
				return ErrDuplicateCommandName
			}
		}

		// Appends command to Roots CommandList
		r.CommandList = append(r.CommandList, cmd)
	}

	return nil
}

// validate checks if a Root object is valid for argbin
func (r *Root) validate() error {
	// Checks if root is nil
	if r == nil {
		return ErrNilRoot
	}

	// Checks if AppName is nil
	if r.AppName == "" {
		return ErrEmptyRootName
	}

	// Checks if HelpMenu is nil
	if r.HelpMenu == "" {
		return ErrEmptyHelpMenu
	}

	// Checks if CommandList is nil
	if r.CommandList == nil {
		r.CommandList = make([]*Command, 0)
	} else {
		// Checks if all Commands in CommandList are valid for argbin
		if err := r.validateCommandList(); err != nil {
			return err
		}
	}

	return nil
}

// validateCommandList checks if every command in the Roots CommandList is valid.
func (r *Root) validateCommandList() error {
	// Loop through each command in CommandList
	for _, cmd := range r.CommandList {

		// Checks if a command is a nil
		if cmd == nil {
			return ErrNilCommand
		}

		// Validates the singular command
		if err := cmd.validate(); err != nil {
			return err
		}
	}

	return nil
}

// parseCommand checks if a given argument is apart of the Roots CommandList.
func (r *Root) parseCommand(ctx *Context, arg string) (*Command, error) {
	// Loop through each command in CommandList
	for _, cmd := range r.CommandList {
		// Validates each command
		if err := cmd.validate(); err != nil {
			return nil, err
		}

		// Checks if the command name is equal to the given arg
		if cmd.Name == arg {
			// Parses subcommands if they exist
			subCmd, err := checkSubcommands(cmd, ctx)
			if err != nil {
				return cmd, err
			}

			if err := validateSubCommand(subCmd); err != nil {
				return cmd, err
			} else {
				subCmd.isSubcommand = true
				return subCmd, nil
			}
		}

		// Checks command aliases to see if given arg is a command
		if slices.Contains(cmd.AdditionalNames, arg) {
			// Parses subcommands if they exist
			subCmd, err := checkSubcommands(cmd, ctx)
			if err != nil {
				return cmd, err
			}

			if err := validateSubCommand(subCmd); err != nil {
				ctx.Command = cmd
				return cmd, err
			} else {
				subCmd.isSubcommand = true
				ctx.Command = subCmd
				return subCmd, nil
			}
		}
	}

	// Returns nil if no commands are found.
	return nil, nil
}
