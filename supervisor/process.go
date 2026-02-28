package supervisor

import (
	"fmt"
	"io"
	"os/exec"
	"pearson-vpn-service/logconfig"
	"syscall"
	"time"
)

func NewProcess(name string, args ...string) (Process, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe for process %s: %w", name, err)
	}

	return &process{
		cmd:    cmd,
		id:     name,
		args:   args,
		status: Initialising,
		stdout: stdout,
		done:   make(chan struct{}),
	}, nil
}

func (p *process) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.status != Initialising && p.status != Stopped {
		return fmt.Errorf("cannot start process %s: status is %s", p.id, p.status.String())
	}

	logconfig.Log.Infof("Starting process %s", p.id)
	if err := p.cmd.Start(); err != nil {
		p.status = Failed
		return fmt.Errorf("failed to start process %s: %w", p.id, err)
	}

	p.status = Running
	logconfig.Log.Infof("Process %s started successfully (pid %d)", p.id, p.cmd.Process.Pid)
	go p.wait()
	return nil
}

func (p *process) Stop() error {
	p.mutex.Lock()
	if p.status != Running {
		p.mutex.Unlock()
		return fmt.Errorf("cannot stop process %s: status is %s", p.id, p.status.String())
	}
	logconfig.Log.Infof("Sending SIGTERM to process %s", p.id)
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		p.mutex.Unlock()
		return fmt.Errorf("failed to send SIGTERM to process %s: %w", p.id, err)
	}
	p.status = Stopped
	p.mutex.Unlock()

	logconfig.Log.Infof("Waiting for process %s to exit", p.id)
	select {
	case <-p.done:
		logconfig.Log.Infof("Process %s exited after SIGTERM", p.id)
	case <-time.After(10 * time.Second):
		logconfig.Log.Warnf("Process %s did not exit in 10s, sending SIGKILL", p.id)
		_ = p.cmd.Process.Kill()
		<-p.done
		logconfig.Log.Infof("Process %s exited after SIGKILL", p.id)
	}
	return nil
}

func (p *process) wait() {
	err := p.cmd.Wait()
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if err != nil && p.status != Stopped {
		logconfig.Log.Warnf("Process %s exited unexpectedly: %v", p.id, err)
		p.status = Failed
	}
	close(p.done)
}

func (p *process) GetStatus() ProcessStatus {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.status
}

func (p *process) GetStdoutStream() io.ReadCloser {
	return p.stdout
}

func (p *process) GetProcessID() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.id
}
