package ktl

const (
	Waiting State = iota
	Paused
	Ready
	Active
	Success
	Failed
)

// State describes the state of a step.
type State int

// Done returns true if the state is Succeeded or Failed.
func (s State) Done() bool {
	return s == Success || s == Failed
}

//go:generate enumer -type=State -text
