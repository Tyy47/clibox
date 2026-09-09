// Package argbin provides tools for defining CLI commands and flags, with built-in argument parsing.
package argbin

import (
	"errors"
	"fmt"

	"github.com/Tyy47/clibox/internal/utils"
)

// Variable collection of root related errors
var (
	// Generic Errors

	ErrNilArray             = errors.New("array cannot be nil")
	ErrNilMap               = errors.New("map cannot be nil")
	ErrEmptyDescription     = errors.New("description field cannot be empty")
	ErrNilCommand           = errors.New("command cannot be nil")
	ErrMissingArguments     = errors.New("no arguments provided")
	ErrUnknownCommand       = errors.New("unknown command")
	ErrEmptyCommandName     = errors.New("command name cannot be empty")

	// Context Errors

	ErrNilContext           = errors.New("context cannot be nil")

	// Root Errors

	ErrNilRoot              = errors.New("root cannot be nil")
	ErrEmptyRootName        = errors.New("appname cannot be blank")
	ErrEmptyVersionNumber   = errors.New("version number cannot be blank")
	ErrEmptyCommandList     = errors.New("command list cannot be empty")

	// Command Errors

	ErrNilCommandFunction   = errors.New("command execute field cannot be nil")
	ErrDuplicateCommandName = errors.New("command names cannot be duplicated")


	// Flag Errors

	ErrNilFlags             = errors.New("flags cannot be nil")
)

// Context is a list of data that can be used to store and access data.
type Context struct {
	// Command stores the executed command. Can be accessed for command fields.
	Command *Command

	// Values stores needed context between flags and commands.
	Values map[string]any

	// Args gathered from os.Args, starts at os.Args[1:].
	Args []string

	// AdditionalArgs that start from index 3 (length of 4).
	AdditionalArgs []string

	// Value that is gathered after a command. (e.g "appname" "command" "parsedvalue")
	ParsedValue string
}

// Validate checks if a context object is valid for processing, returns an error if it's not.
func (c *Context) Validate() error {
	// Checks if context is nil, returns an error if so.
	if c == nil {
		return ErrNilContext
	}

	// Checks if contexts command is nil, if so, it'll create a command and continue.
	if c.Command == nil {
		c.Command = &Command{}
	}

	// Checks if context Values is nil, if so, it'll create a string:any map.
	if c.Values == nil {
		c.Values = make(map[string]any)
	}

	// Return nil to satisfy return
	return nil
}

// ToggleValue switches values inside of Context.Values, returns an error if unable to make changes.
func (ctx *Context) ToggleValue(key string, toggle bool) error {
	// Checks if the given key is found inside of ctx.Values
	if _, ok := ctx.Values[key].(string); ok {

		// If it's found, it'll toggle the value to whatever toggle is equal to.
		ctx.Values[key] = toggle
		return nil

	} else {
		// Returns an error if a key isn't found in the values map.
		return fmt.Errorf("%s doesn't exist in context values", key)
	}
}

// GetValue retrieves a value from a given key, returns the value and an error if failed.
func (ctx *Context) GetValue(key string) (any, error) {
	// Attempts to find a value based on the given key
	if token, ok := ctx.Values[key]; ok {
		// Returns the found value
		return token, nil
	} else {
		// Returns an error if no value was found
		return nil, fmt.Errorf("%s doesn't exist in context values", key)
	}
}

// Run is the execution of your program with all combined commands.
func (r *Root) Run() error {
	// Validates the Root object to make sure it's valid.
	if err := r.validate(); err != nil {
		return err
	}

	// Gathers args from an arg wrapper in utils.
	args := *utils.GetArgs()

	// Creates the context object to store user data.
	ctx := Context{
		Values:      make(map[string]any),
		Args:        args,
		ParsedValue: "",
	}

	// Checks if the app execution is valid, if there is zero arguments, it returns an error.
	if len(args) == 0 {
		return ErrMissingArguments
	}

	// Loops over every arg given
	for i, arg := range args {

		// Checks if the argument is a command, returns nil if not.
		cmd, err := r.parseCommand(&ctx, arg)

		// If there is no command, it skips the argument.
		if cmd == nil {
			continue
		}

		// If there is a command parsing error, it'll check and return.
		if err != nil {
			return err
		}

		// Checks if the context is validate.
		if err := ctx.Validate(); err != nil {
			return err
		}

		// If there is 2 or more arguments, it will parse flags.
		if len(args) >= 2 {
			if err := cmd.runFlags(&ctx, args[i+1:]); err != nil {
				return err
			}
		}

		// If a command takes a value, it will grab the subsiquent argument and add it to ctx.ParsedValue. As well as add additional arguments to context.
		if cmd.TakesValue {
			if i+1 >= len(args) {
				return fmt.Errorf("command %v requires a value", cmd.Name)
			}

			ctx.ParsedValue = args[i+1]
			ctx.AdditionalArgs = nil
			if start := i + 3; start <= len(args) {
				ctx.AdditionalArgs = args[start:]
			}

			i++
		}

		// Executes a command using the given context to manipulate.
		if err := cmd.Execute(&ctx); err != nil {
			return err
		} else {
			return nil
		}
	}

	// Returns an error if no commands are found with a given argument.
	return fmt.Errorf("%w: %s", ErrUnknownCommand, args[0])
}
