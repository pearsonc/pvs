package supervisor

import (
	"bytes"
	"os/exec"
	"sync"
)

type process struct {
	cmd    *exec.Cmd
	id     string
	args   []string
	status ProcessStatus
	output *bytes.Buffer
	mutex  sync.Mutex
}

type Process interface {
	reinitialise() error
	Start() error
	Stop() error
	Restart() error
	wait()
	GetStatus() ProcessStatus
	GetOutput() string
	GetProcessID() string
}
