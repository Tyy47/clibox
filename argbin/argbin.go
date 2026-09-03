package argbin

import (
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
	Aliases []string
	Execute func(ctx *Context) error
}

type Context struct {
	Command *Command
	Values map[string]any
}

func (c *Command) searchFlag(name string) (*Flag, bool) {
	for key, flag := range c.Flags {
		if key == name || flag.Name == name {
			return flag, true
		}

		for _, alias := range flag.Aliases {
			if alias == name {
				return flag, true
			}
		}
	}


	return nil, false
}

func (c *Command) parseFlags(ctx *Context, args []string) error {
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
func validCommandChecker(commands CommandList) error {
	for i := range commands {

		if commands[i].Name == "" {
			return fmt.Errorf("Command name member can't be blank. Command: %v", commands[i])
		}

		if commands[i].Description == "" {
			return fmt.Errorf("Command description member can't be blank. Command: %v", commands[i])
		}

		if commands[i].Execute == nil {
			return fmt.Errorf("Command execute member can't be blank. Command: %v", commands[i])
		}
	}

	return nil
}

// validFlagChecker checks if a given flag is valid to work with argbins parser
func validFlagChecker(flag *Flag) error {
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
	for takenName := range r.Commands {
		if command.Name == takenName {
			return fmt.Errorf("Command name %s already exists in command list: %v", command.Name, r.Commands)
		}
	}
	r.Commands[command.Name] = command
	return nil
}

func (r *Root) Run() error {
	
	// Users arguments
	args := *utils.GetArgs()
	
	if err := validCommandChecker(r.Commands); err != nil {
		return err
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		
		command, ok := r.Commands[arg]
		if !ok {
			continue
		}

		ctx := &Context{
			Command: command,
			Values: make(map[string]any),
		}

		if err := command.parseFlags(ctx, args[i+1:]); err != nil {
			return err
		}

		if err := command.Execute(ctx); err != nil {
			return fmt.Errorf("Command cannot be executed: %v", err)
		}

	}

	return nil
}




