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
	done   chan struct{}
	mutex  sync.Mutex
}

type Process interface {
	Start() error
	Stop() error
	GetStatus() ProcessStatus
	GetProcessID() string
	GetStdoutStream() io.ReadCloser
}
