package inputbin

import (
	"errors"
	"fmt"
)

var (

	ErrNoInputGiven  = errors.New("missing required input")
)

//
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

