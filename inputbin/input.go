package inputbin

import (
	"errors"
	"fmt"
)

var (

	// InputOptions Errors

	ErrNilInputOptions = errors.New("input options cannot be nil")
	ErrEmptyQuestion = errors.New("question field in inputoptions cannot be empty")
	ErrEmptyAnswers = errors.New("answers field in inputoptions cannot be empty")

	ErrNoInputGiven  = errors.New("missing required input")
)

type InputOptions struct {
	
	// Basic required information for user inputs
	
	// Prefix is an optional value that goes before the question. 
	// Example: Prefix: Question
	Prefix string

	// Question is a required string that is presented to the user.
	Question string

	//  Answers is a required field as these will be presented to the user for selection.
	Answers []string
}

// validate checks if the InputOptions received in a function are valid for use
func (i *InputOptions) validate() error {
	
	if i.Question == "" {
		return ErrEmptyQuestion
	}

	if i.Answers == nil {
		return ErrEmptyAnswers
	}

	return nil
}

func Prompt(options *InputOptions) (string, error) {
	
	if options == nil {
		return "", ErrNilInputOptions
	}

	if err := options.validate(); err != nil {
		return "", err
	}

	label := options.Prefix + options.Question

	fmt.Println(label)
	
	for i := range options.Answers {
		fmt.Printf("%d. %s\n", i, options.Answers[i])
	}

	var userSelection string

	_, err := fmt.Scanln(&userSelection)
	if err != nil {
		return "", ErrNoInputGiven
	}

	return userSelection, nil
}
