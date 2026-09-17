package inputbin

import (
	"errors"
	"fmt"
)

var (

	ErrNoInputGiven  = errors.New("missing required input")
)

// Prompt allows the user to receive a prompt with a question and a selection of answers.
// promtPrefix is an option prefix that gets added to the question. Example: promtPrefix: question.
func Prompt(promtPrefix string, question string, answers []string) (string, error) {
	if promtPrefix == "" {
		fmt.Println(question)
	} else {
		fmt.Printf("%s: %s\n", promtPrefix, question)
	}

	
	for i := range answers {
		fmt.Printf("%d. %s\n", i, answers[i])
	}

	var userSelection string

	_, err := fmt.Scanln(&userSelection)
	if err != nil {
		return "", ErrNoInputGiven
	}

	return userSelection, nil
}

