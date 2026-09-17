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
func Prompt(options *InputOptions) (string, error) {
	
	if options == nil {
		return "", ErrNilInputOptions
	}

	if err := options.validate(); err != nil {
		return "", err
	}

	modifyInputOptionColors(options)


	label := options.Prefix + options.Question

	fmt.Println(label)

	if options.AnswersColor != nil {
		for i := range options.Answers {
			fmt.Printf("%d. %s\n", i, colorbin.ColorStrings(*options.AnswersColor, options.Answers...)[i])
		}
	} else {
		for i := range options.Answers {
			fmt.Printf("%d. %s\n", i, options.Answers[i])
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
