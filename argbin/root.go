package argbin

import (
	"errors"
	"fmt"
	"log"
	"slices"
)

// const array of a collection of root related errors
var (
	ErrNilRoot = errors.New("root cannot be nil")
	ErrEmptyRootName = errors.New("appname cannot be blank")
	ErrEmptyVersionNumber = errors.New("version number cannot be blank")
	ErrEmptyCommandList = errors.New("command list cannot be empty")
	ErrDuplicateCommandName = errors.New("command names cannot be duplicated")
	ErrNoCommandFound = errors.New("command doesn't exist")
	ErrNilCommandFunction = errors.New("command execute field cannot be nil")
)


type Root struct {
	// AppName stores the name of your application.
	AppName string
	
	// AppVersion stores the version of your application.
	AppVersion string 
	
	// CommandList stores all the requires commands for your application.
	CommandList []*Command
}

// GetAppName returns the name of the application.
func (r *Root) GetAppName() (string, error) {
	if r == nil {
		return "", ErrNilRoot
	}
	if r.AppName == "" {
		err := fmt.Errorf("%w: setting name to default 'appname'.", ErrEmptyRootName)
		log.Println(err)
		r.SetAppName("appname")
	}
	return r.AppName, nil
}

// SetAppName sets the name of the application.
func (r *Root) SetAppName(name string) error {
	if r == nil {
		return ErrNilRoot
	}
	if name == "" {
		err := fmt.Errorf("%w: setting name to default 'appname'.", ErrEmptyRootName)
		log.Println(err)
		r.SetAppName("appname")
		return nil
	}
	r.AppName = name
	return nil
}

// GetAppVersion returns the version of the application.
func (r *Root) GetAppVersion() (string, error) {
	if r == nil {
		return "", ErrNilRoot
	}
	if r.AppVersion == "" {
		err := fmt.Errorf("%w: setting version to default '1.0.0'.", ErrEmptyVersionNumber)
		log.Println(err)
		r.SetAppVersion("1.0.0")
	}
	return r.AppVersion, nil
}

// SetAppVersion sets the version of the application.
func (r *Root) SetAppVersion(version string) error {
	if r == nil {
		return ErrNilRoot
	}
	if version == "" {
		err := fmt.Errorf("%w: setting version to default '1.0.0'.", ErrEmptyVersionNumber)
		log.Println(err)
		r.SetAppVersion("1.0.0")
		return nil
	}
	r.AppVersion = version
	return nil
}

func (r *Root) GetCommandList() ([]*Command, error) {
	if r == nil {
		return nil, ErrNilRoot
	}
	
	if r.CommandList == nil {
		return nil, ErrEmptyCommandList
	}

	return r.CommandList, nil
}

func (r *Root) SetCommandList(commandList []*Command) error {
	if r == nil {
		return ErrNilRoot
	}

	if len(commandList) == 0 {
		return ErrEmptyCommandList
	}

	r.CommandList = commandList
	return nil
}

func (r *Root) AddCommand(command *Command) error {
	if r == nil {
		return ErrNilRoot
	}

	if r.CommandList == nil {
		return ErrEmptyCommandList
	}
	
	if err := command.validate(); err != nil {
		return err
	}

	for _, cmd := range r.CommandList {
		if cmd.Name == command.Name {
			return ErrDuplicateCommandName
		}
	}
	
	r.CommandList = append(r.CommandList, command)

	return nil
}

func (r *Root) validate() error {
	if r == nil {
		return ErrNilRoot
	}

	if r.AppName == "" {
		return ErrEmptyRootName
	}

	if r.AppVersion == "" {
		return ErrEmptyVersionNumber
	}

	if r.CommandList == nil {
		r.CommandList = make([]*Command, 0)
	} else {
		if err := r.validateCommandList(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Root) validateCommandList() error {
	for _, cmd := range r.CommandList {
		if cmd == nil {
			return ErrNilCommand
		}

		if err := cmd.validate(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Root) parseCommand(ctx *Context, arg string) (*Command, error) {
	for _, cmd := range r.CommandList {

		if err := cmd.validate(); err != nil {
			return nil, err
		}

		if cmd.Name == arg {
			ctx.Command = cmd
			return cmd, nil
		}

		if slices.Contains(cmd.AdditionalNames, arg) {
			return cmd, nil
		}
	}

	return nil, fmt.Errorf("%s: %w", arg, ErrNoCommandFound)
}
