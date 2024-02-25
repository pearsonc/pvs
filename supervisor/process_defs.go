package supervisor

import (
	"bytes"
	"os/exec"
	"sync"
)

type process struct {
	cmd    *exec.Cmd
	id     string
	status ProcessStatus
	output *bytes.Buffer
	mutex  sync.Mutex
}

type Process interface {
	Start() error
	Stop() error
	Restart() error
	wait()
	GetStatus() ProcessStatus
	GetOutput() string
	GetProcessID() string
}
