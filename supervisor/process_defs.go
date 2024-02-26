package supervisor

import (
	"io"
	"os/exec"
	"sync"
)

type process struct {
	cmd    *exec.Cmd
	id     string
	args   []string
	status ProcessStatus
	stdout io.ReadCloser
	stderr io.ReadCloser
	mutex  sync.Mutex
}

type Process interface {
	reinitialise() error
	Start() error
	Stop() error
	Restart() error
	wait()
	GetStatus() ProcessStatus
	GetProcessID() string
	GetStdoutStream() io.ReadCloser
}
