package argbin

import "errors"

var (
	ErrNilCommand = errors.New("command cannot be nil")
)

// Command is the storage where you'll input all of your command information.
type Command struct {
	// Name of the command that'll be ran
	Name string

	// Additional names is where aliases are stored for a command.
	AdditionalNames []string

	// Flags for Command are stored as k,v pairs as a string and a function.
	Flags map[string]func() error

	// Context stores values that might require to be called upon later.
	Context Context
	
	// TakesValue scans subsequent arguments to find a valid value.
	TakesValue bool
}

func (c *Command) GetName() (string, error) {
	if c == nil {
		return "", ErrNilCommand
	}
	return c.Name, nil
}

func (c *Command) GetAdditionalNames() ([]string, error) {
	if c == nil {
		return nil, ErrNilCommand
	}
	return c.AdditionalNames, nil
}

func (c *Command) GetContext() (Context, error) {
	if c == nil {
		return Context{}, ErrNilCommand
	}
	return c.Context, nil
}
