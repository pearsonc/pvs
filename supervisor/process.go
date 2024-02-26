package supervisor

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func NewProcess(name string, args ...string) Process {

	cmd := exec.Command(name, args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	return &process{
		cmd:    cmd,
		id:     name,
		args:   args,
		status: Initialising,
		stdout: stdout,
		stderr: stderr,
	}
}

func (p *process) reinitialise() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	fmt.Printf("Reinitialising process %s\n", p.id)
	if p.cmd.Process != nil {
		if err := p.cmd.Process.Signal(syscall.Signal(0)); err != nil {
			if errors.Is(err, os.ErrProcessDone) {
				fmt.Printf("Process %s has finished.\n", p.id)
			} else {
				fmt.Printf("Error signaling process %s (it might be dead): %v\n", p.id, err)
			}
		} else {
			if err1 := p.cmd.Process.Kill(); err != nil {
				fmt.Printf("Error killing process %s: %v\n", p.id, err)
				return err1
			}
		}
	}
	p.status = Restarting
	//p.output.Reset()
	p.cmd = exec.Command(p.id, p.args...)
	//p.cmd.Stdout = p.output
	time.Sleep(2 * time.Second)
	p.mutex.Unlock()
	err := p.Start()
	p.mutex.Lock()
	if err != nil {
		fmt.Printf("Error restarting process %s: %v\n", p.id, err)
		return err
	}

	return nil
}
func (p *process) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.status != Initialising && p.status != Stopped && p.status != Restarting {
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
func (p *process) GetStdoutStream() io.ReadCloser {
	return p.stdout
}

func (p *process) GetProcessID() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.id
}
