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
	Status string
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
		Status: "initialised",
		Output: outputBuffer,
	}
}
func (p *Process) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.Status != "initialised" && p.Status != "stopped" {
		return errors.New("process already started or not in a restartable state")
	}

	err := p.Cmd.Start()
	if err != nil {
		p.Status = "failed"
		return err
	}

	p.Status = "running"
	go p.wait()
	return nil
}
func (p *Process) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.Status != "running" {
		return errors.New("process not running")
	}

	err := p.Cmd.Process.Kill()
	if err != nil {
		return err
	}

	p.Status = "stopped"
	return nil
}
func (p *Process) wait() {
	err := p.Cmd.Wait()
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if err != nil {
		// Log the error, or handle it as needed
		p.Status = "failed"
	} else {
		p.Status = "stopped"
	}
}

func (p *Process) GetStatus() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.Status
}
func (p *Process) GetOutput() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.Output.String()
}
