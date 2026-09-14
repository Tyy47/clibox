package argbin

// Flag stores all data related to a commands flag
type Flag struct {
	// Execute runs the flags given function
	Execute FlagFunction

	// Takes the next argument in line after the flag if true
	TakesValue bool
}

// FlagFunction configures a flag using the provided Context.                           
//                                                                                         
// Return a non-nil error to stop configuration and report the failure.                    
type FlagFunction func(ctx *Context) error

// validate runs a check to make sure a given flag is valid for argbin.
func (f *Flag) validate() error {
	if f.Execute == nil {
		return ErrNilFlagExecute
	}

	return nil
}
