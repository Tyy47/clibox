package inputbin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Tyy47/clibox/colorbin"
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

	// Color customization
	
	// PrefixColor sets the colors for the Prefix string
	PrefixColor *colorbin.ColorOptions

	// QuestionColor sets the colors for the Question string
	QuestionColor *colorbin.ColorOptions

	// AnswersColor sets the colors for each of the answers provided in the Answers array
	AnswersColor *colorbin.ColorOptions
}

// validate checks if the InputOptions received in a function are valid for use
func (i *InputOptions) validate() error {
	
	if i.Question == "" {
		return ErrEmptyQuestion
	}

	if i.Answers == nil {
		i.Answers = make([]string, 0)
	}

	return nil
}

// modifyInputOptionColors takes in an InputOptions object and modifies the colors of the strings
// if the Color modifying fields are set in InputOptions.
func modifyInputOptionColors(i *InputOptions) {

	if i.PrefixColor != nil {
		i.Prefix = colorbin.ColorStrings(*i.PrefixColor, i.Prefix)[0]
	}

	if i.QuestionColor != nil {
		i.Question = colorbin.ColorStrings(*i.QuestionColor, i.Question)[0]
	}

	if i.AnswersColor != nil {
		i.Answers = colorbin.ColorStrings(*i.AnswersColor, i.Answers...)
	}
}


// Prompt takes in a set of InputOptions and delievers a prompt to the cli based on the options provided.
// Returns a string of the users response and an error if input was nil.
func Prompt(ops *InputOptions) (string, error) {
	
	if ops == nil {
		return "", ErrNilInputOptions
	}

	if err := ops.validate(); err != nil {
		return "", err
	}

	modifyInputOptionColors(ops)


	label := ops.Prefix + ops.Question

	fmt.Println(label)

	if ops.AnswersColor != nil {
		for i := range ops.Answers {
			fmt.Printf("%d. %s\n", i, colorbin.ColorStrings(*ops.AnswersColor, ops.Answers...)[i])
		}
	} else {
		for i := range ops.Answers {
			fmt.Printf("%d. %s\n", i, ops.Answers[i])
		}
	}
	
	var userSelection string

	_, err := fmt.Scanln(&userSelection)
	if err != nil {
		return "", ErrNoInputGiven
	}

	return userSelection, nil
}

// Confirm presents the user with a confirmation prompt to continue on with something.
// "yes" returns true and "no" false. defaultYes will make the default option yes if the user presses enter with no entry.
func Confirm(ops *InputOptions, defaultYes bool) bool {

	// Valid check for ops input
	ops.validate()

	// Modify colors if given
	modifyInputOptionColors(ops)

	// Display the prompt to the user
	fmt.Println(ops.Question)
	
	// Init the string container
	var userInput string

	// Prints prefix if available
	if ops.Prefix != "" {
		fmt.Println(ops.Prefix)
	}

	// Grab users entry
	fmt.Scan(&userInput)

	// Lowercase the input
	userInput = strings.ToLower(userInput)

	// Return the result based on user response
	switch userInput {
	case "yes", "ye", "y":
		return true
	case "no", "n":
		return false
	default:
		return defaultYes
	}
}

// Text takes in a set of options and displays a single prompt to the user. 
// Users response is returned as a string.
func Text(ops *InputOptions) (string, error) {

	// Nil check on ops
	if ops == nil {
		return "", ErrNilInputOptions
	}

	// Colors input options if fields are filled
	modifyInputOptionColors(ops)
	
	// Prints the question from ops
	fmt.Printf("%s", ops.Question)

	// Storage for users response
	var usersInput string

	// Grab users input
	_, err := fmt.Scan(&usersInput)
	if err != nil {
		return "", ErrNoInputGiven
	}

	return usersInput, nil
}
