package supervisor

import (
	"bytes"
	"errors"
	"os/exec"
)

func NewProcess(name string, args ...string) Process {
	outputBuffer := &bytes.Buffer{}
	cmd := exec.Command(name, args...)
	cmd.Stdout = outputBuffer

	return &process{
		cmd:    cmd,
		id:     name,
		status: Initialising,
		output: outputBuffer,
	}
}
func (p *process) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.status != Initialising && p.status != Stopped {
		return errors.New("process already started or not in a restartable state")
	}

	err := p.cmd.Start()
	if err != nil {
		p.status = Failed
		return err
	}

	p.status = Running
	go p.wait()
	return nil
}
func (p *process) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.status != Running {
		return errors.New("process not running")
	}

	err := p.cmd.Process.Kill()
	if err != nil {
		return err
	}

	p.status = Stopped
	return nil
}
func (p *process) Restart() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.status != Running {
		return errors.New("process not running")
	}

	err := p.cmd.Process.Kill()
	if err != nil {
		return err
	}

	p.status = Restarting
	return p.Start()

}
func (p *process) wait() {
	err := p.cmd.Wait()
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if err != nil {
		// Log the error, or handle it as needed
		p.status = Failed
	} else {
		p.status = Stopped
	}
}

func (p *process) GetStatus() ProcessStatus {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.status
}
func (p *process) GetOutput() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.output.String()
}

func (p *process) GetProcessID() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.id
}
