package ktl

import (
	"fmt"
	"strings"
)

func ValidateBatch(b *Batch) error {
	var errs []error
	err := func(format string, args ...interface{}) {
		errs = append(errs, fmt.Errorf(format, args...))
	}

	if len(b.Steps) == 0 {
		err(`"steps" is required, but empty`)
	}

	// count occurences of step IDs. Each ID should be referenced once.
	allIDs := map[string]int{}

	for i, step := range b.Steps {
		if step.ID == "" {
			err(`step[%d].ID is required, but empty`, i)
		}
		allIDs[step.ID]++
		// TODO validate types
	}

	for id, count := range allIDs {
		if count > 1 {
			err(`multiple (%d) steps have the same ID: "%s"`, count, id)
		}
	}

	for i, step := range b.Steps {
		for j, dep := range step.Dependencies {
			_, ok := allIDs[dep]
			if !ok {
				err(`step[%d].Dependencies[%d] references a step that doesn't exist: "%s"`, i, j, dep)
			}
		}
	}

	if errs == nil {
		return nil
	}
	return MultiError(errs)
}

type MultiError []error

func (me MultiError) Error() string {
	var strs []string
	for _, e := range me {
		strs = append(strs, e.Error())
	}
	return strings.Join(strs, "; ")
}
