//
package argbin

import (

)

// Root is the "root" of the project/app. 
//
// Members:
// 	- AppName: Stores the name of your application.
// 	- Description: A brief description of your application.
// 	- Commands: A map of available commands.
type Root struct {
	AppName string
	Description string	
	Commands map[string]*Command
}

type Command struct {
	Name string
	Description string
	Flags []*Flag
}


type Flag struct {
	Name string
	
}


func Run() {}
