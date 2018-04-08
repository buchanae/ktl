package ktl

import (
	"fmt"
)

const (
	// BestEffort mode will run as many steps as possible; any steps which
	// are not blocked by upstream errors will be run.
	// This is the default mode.
	BestEffort Mode = iota
	// FailFast mode will stop all processing on the first error;
	// Any parallel steps which are running will be stopped immediately.
	FailFast
)

// Mode determines how the execution engine will handle errors.
type Mode int

func (m *Mode) String() string {
	switch *m {
	case BestEffort:
		return "BestEffort"
	case FailFast:
		return "FailFast"
	}
	return "Unknown"
}

func (m *Mode) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *Mode) UnmarshalText(b []byte) error {
	switch string(b) {
	case "BestEffort":
		*m = BestEffort
	case "FailFast":
		*m = FailFast
	default:
		return fmt.Errorf(`unknown Mode "%s"`, string(b))
	}
	return nil
}
