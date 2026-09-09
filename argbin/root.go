package argbin

import (
	"fmt"
	"slices"
)

// Root is the application structure where all data is held.
type Root struct {
	// AppName stores the name of your application.
	AppName string
	
	// AppVersion stores the version of your application.
	AppVersion string 

	// Description is your Apps help menu.
	Description string
	
	// CommandList stores all the requires commands for your application.
	CommandList []*Command
}

// GetAppName returns the name of the application.
func (r *Root) GetAppName() (string, error) {

	// Checks if the root object is nil
	if r == nil {
		return "", ErrNilRoot
	}

	// Checks if the appname is empty, if so, returns an error.
	if r.AppName == "" {
		return "", fmt.Errorf("cannot get empty app name")
	}

	// Returns the app name
	return r.AppName, nil
}

// SetAppName sets the name of the application.
func (r *Root) SetAppName(name string) error {

	// Checks if the root object is nil
	if r == nil {
		return ErrNilRoot
	}

	// Check if the given name is blank
	if name == "" {
		return fmt.Errorf("name cannot be empty when setting app name.")
	}
	
	// Sets app name to given name argument
	r.AppName = name
	return nil
}

// GetAppVersion returns the version of the application.
func (r *Root) GetAppVersion() (string, error) {

	// Checks if root is nil
	if r == nil {
		return "", ErrNilRoot
	}

	// Checks if the app version is blank
	if r.AppVersion == "" {
		return "", fmt.Errorf("cannot get empty app version")
	}

	// Returns the app version
	return r.AppVersion, nil
}

// SetAppVersion sets the version of the application.
func (r *Root) SetAppVersion(version string) error {

	// Checks if root is nil
	if r == nil {
		return ErrNilRoot
	}

	// Checks if the version argument is empty
	if version == "" {
		return fmt.Errorf("version number cannot be set to empty string")
	}
	
	// Assigns version to AppVersion
	r.AppVersion = version
	return nil
}

// GetDescription returns the Root's Description field alongside an error.
func (r *Root) GetDescription() (string, error) {

	// Checks if root is nil
	if r == nil {
		return "", ErrNilRoot
	}
	
	// Returns an error if roots Description is empty
	if r.Description == "" {
		return "", fmt.Errorf("root description is empty")
	}
	
	// Returns the apps Description
	return r.Description, nil
}

// SetDescription sets the Description field as des
func (r *Root) SetDescription(des string) error {

	// Checks if root is nil
	if r == nil {
		return ErrNilRoot
	}

	// Checks if the des argument is empty
	if des == "" {
		return fmt.Errorf("des value cannot be blank when setting description")
	}
	
	// Sets the roots Description to des
	r.Description = des
	return nil
}

// GetCommandList returns the Roots CommandList field
func (r *Root) GetCommandList() ([]*Command, error) {

	// Checks if root is nil
	if r == nil {
		return nil, ErrNilRoot
	}
	
	// Checks if the CommandList is nil, if so, it'll create an empty *Command array.
	if r.CommandList == nil {
		r.CommandList = make([]*Command, 0)
	}
	
	// Returns the roots CommandList
	return r.CommandList, nil
}

// SetCommandList takes an array of Command pointers and assigns it to r.CommandList. 
func (r *Root) SetCommandList(commandList []*Command) error {

	// Checks if root is nil
	if r == nil {
		return ErrNilRoot
	}
	
	// Checks if the commandList argument is empty
	if len(commandList) == 0 {
		return ErrEmptyCommandList
	}

	// Assigns commandList to r.CommandList
	r.CommandList = commandList
	return nil
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

	// Checks if AppVersion is nil
	if r.AppVersion == "" {
		return ErrEmptyVersionNumber
	}

	// Checks if Description is nil
	if r.Description == "" {
		return ErrEmptyDescription
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
			ctx.Command = cmd
			return cmd, nil
		}

		// Checks command aliases to see if given arg is a command
		if slices.Contains(cmd.AdditionalNames, arg) {
			ctx.Command = cmd
			return cmd, nil
		}
	}
	
	// Returns nil of no commands are found.
	return nil, nil
}
