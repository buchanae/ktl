package ktl

import (
	"time"
)

// Duration is a wrapper type for time.Duration which provides human-friendly
// text (un)marshaling.
// See https://github.com/golang/go/issues/16039
type Duration time.Duration

// String returns the string representation of the duration.
func (d Duration) String() string {
	return time.Duration(d).String()
}

// UnmarshalText parses text into a duration value.
func (d *Duration) UnmarshalText(text []byte) error {
	// Ignore if there is no value set.
	if len(text) == 0 {
		return nil
	}

	// Otherwise parse as a duration formatted string.
	duration, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}

	// Set duration and return.
	*d = Duration(duration)
	return nil
}

// MarshalText converts a duration to text.
func (d Duration) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}
