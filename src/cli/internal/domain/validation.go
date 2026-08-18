package domain

import (
	"fmt"
	"regexp"
)

var identifierPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidateIdentifier(kind, value string) error {
	if !identifierPattern.MatchString(value) {
		return fmt.Errorf("%w: %s %q must be kebab-case", ErrInvalidIdentifier, kind, value)
	}
	return nil
}

func ValidateActor(actor Actor) error {
	switch actor.Kind {
	case ActorHuman, ActorAgent, ActorCLI, ActorSystem:
	default:
		return fmt.Errorf("%w: unknown actor kind %q", ErrInvalidActor, actor.Kind)
	}
	if actor.ID == "" {
		return fmt.Errorf("%w: actor id cannot be empty", ErrInvalidActor)
	}
	return nil
}
