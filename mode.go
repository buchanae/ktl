package ktl

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

//go:generate enumer -type=Mode -text
