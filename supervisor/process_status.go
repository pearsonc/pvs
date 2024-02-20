package supervisor

// ProcessStatus represents the status of a process managed by the supervisor.
type ProcessStatus int

const (
	Initialising ProcessStatus = iota
	Running
	Stopped
	Failed
)

func (ps ProcessStatus) String() string {
	// Convert the ProcessStatus to a human-readable form.
	return [...]string{"initialising", "running", "stopped", "failed"}[ps]
}
