package ktl

// TODO might want to match steps to dag steps. missing blocked, etc.
const (
	Idle State = iota
	Running
  Paused
	Failed
	Success
)

// State describes the state of a batch or step.
type State int

//go:generate enumer -type=State -text
