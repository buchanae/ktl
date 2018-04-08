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

// String returns the string representation of Mode: BestError, FailFast, etc.
func (m *Mode) String() string {
	switch *m {
	case BestEffort:
		return "BestEffort"
	case FailFast:
		return "FailFast"
	}
	return "Unknown"
}

// MarshalText marshals the mode to text, which enables JSON to use
// the string version instead of an int.
func (m *Mode) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText unmarshals the mode from text, which enables JSON to use
// the string version instead of an int.
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
