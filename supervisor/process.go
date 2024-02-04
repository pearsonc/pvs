package supervisor

import (
	"errors"
	"os/exec"
	"sync"
)

// Process represents a single managed process
type Process struct {
	Cmd    *exec.Cmd
	ID     string
	Status string
	mutex  sync.Mutex
}

// NewProcess creates a new Process instance
func NewProcess(name string, args ...string) *Process {
	return &Process{
		Cmd:    exec.Command(name, args...),
		ID:     name,
		Status: "initialised",
	}
}

// Start begins the execution of the process
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

// Stop terminates the process
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

// wait waits for the process to exit and updates the status
func (p *Process) wait() {
	err := p.Cmd.Wait()
	if err != nil {
		return
	}
	p.mutex.Lock()
	p.Status = "stopped"
	p.mutex.Unlock()
}

// GetStatus returns the current status of the process
func (p *Process) GetStatus() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.Status
}
