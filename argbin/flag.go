package argbin




type Flag struct {
	// Execute runs the flags given function
	Execute FlagFunction

	// Takes the next argument in line after the flag if true
	TakesValue bool
}


type FlagFunction func(ctx *Context) error




func (f *Flag) validate() error {

	if f.Execute == nil {
		return ErrNilFlagExecute
	}

	return nil
}
