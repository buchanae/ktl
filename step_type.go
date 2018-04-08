package ktl

import (
	"fmt"
)

const (
	TaskType StepType = iota
)

type StepType int

func (st *StepType) String() string {
	switch *st {
	case TaskType:
		return "Task"
	}
	return "Unknown"
}

func (st *StepType) MarshalText() ([]byte, error) {
	return []byte(st.String()), nil
}

func (st *StepType) UnmarshalText(b []byte) error {
	switch string(b) {
	case "Task":
		*st = TaskType
	default:
		return fmt.Errorf(`unknown StepType "%s"`, string(b))
	}
	return nil
}
