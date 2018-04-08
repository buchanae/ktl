package ktl

import (
	"fmt"
)

const (
	Idle State = iota
	Running
  Failed
	Success
)

type State int

func (s *State) String() string {
	switch *s {
	case Idle:
		return "Idle"
	case Running:
		return "Running"
	case Failed:
		return "Failed"
	case Success:
		return "Success"
	}
	return "Unknown"
}

func (s *State) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *State) UnmarshalText(b []byte) error {
	switch string(b) {
	case "Idle":
		*s = Idle
	case "Running":
		*s = Running
  case "Failed":
    *s = Failed
	case "Success":
		*s = Success
	default:
		return fmt.Errorf(`unknown State "%s"`, string(b))
	}
	return nil
}