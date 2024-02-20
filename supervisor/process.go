package supervisor

import (
	"bytes"
	"errors"
	"os/exec"
	"sync"
)

type Process struct {
	Cmd    *exec.Cmd
	ID     string
	Status ProcessStatus
	Output *bytes.Buffer
	mutex  sync.Mutex
}

func NewProcess(name string, args ...string) *Process {
	outputBuffer := &bytes.Buffer{}
	cmd := exec.Command(name, args...)
	cmd.Stdout = outputBuffer

	return &Process{
		Cmd:    cmd,
		ID:     name,
		Status: Initialising,
		Output: outputBuffer,
	}
}
func (p *Process) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.Status != Initialising && p.Status != Stopped {
		return errors.New("process already started or not in a restartable state")
	}

	err := p.Cmd.Start()
	if err != nil {
		p.Status = Failed
		return err
	}

	p.Status = Running
	go p.wait()
	return nil
}
func (p *Process) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.Status != Running {
		return errors.New("process not running")
	}

	err := p.Cmd.Process.Kill()
	if err != nil {
		return err
	}

	p.Status = Stopped
	return nil
}
func (p *Process) wait() {
	err := p.Cmd.Wait()
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if err != nil {
		// Log the error, or handle it as needed
		p.Status = Failed
	} else {
		p.Status = Stopped
	}
}

func (p *Process) GetStatus() ProcessStatus {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.Status
}
func (p *Process) GetOutput() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.Output.String()
}
