package supervisor

import (
	"errors"
	"fmt"
	"io"
)

func NewManager() ProcessManager {
	return &processManager{
		processes: make(map[string]Process),
	}
}

func (pm *processManager) CreateProcess(name string, args ...string) string {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	p := NewProcess(name, args...)
	pm.processes[p.GetProcessID()] = p
	return p.GetProcessID()
}

func (pm *processManager) StartProcess(id string) error {
	var p Process
	var ok bool
	pm.mutex.RLock()
	p, ok = pm.processes[id]
	pm.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("process with ID %s not found", id)
	}

	return p.Start()
}

func (pm *processManager) StopProcess(id string) error {
	var p Process
	var ok bool
	pm.mutex.RLock()
	p, ok = pm.processes[id]
	pm.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("process with ID %s not found", id)
	}

	return p.Stop()
}

func (pm *processManager) ReinitialiseProcess(id string) error {
	var p Process
	var ok bool
	pm.mutex.RLock()
	p, ok = pm.processes[id]
	pm.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("process with ID %s not found", id)
	}

	return p.reinitialise()

}

func (pm *processManager) RestartProcess(id string) error {
	var p Process
	var ok bool
	pm.mutex.RLock()
	p, ok = pm.processes[id]
	pm.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("process with ID %s not found", id)
	}

	return p.Restart()
}

func (pm *processManager) GetAllProcesses() map[string]Process {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.processes
}

func (pm *processManager) GetStatus(id string) (ProcessStatus, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	p, ok := pm.processes[id]
	if !ok {
		return Failed, errors.New("process not found")
	}

	return p.GetStatus(), nil
}

func (pm *processManager) GetStdoutStream(id string) (io.ReadCloser, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	p, ok := pm.processes[id]
	if !ok {
		return nil, fmt.Errorf("process with ID %s not found", id)
	}

	return p.GetStdoutStream(), nil
}

func (pm *processManager) StartMonitor() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.monitor == nil {
		pm.monitor = NewProcessMonitorInstance(pm)
	} else {
		pm.monitor.StopMonitoring() // Ensure only one monitoring goroutine is running
	}
	pm.monitor.StartMonitoring()
}

func (pm *processManager) StopMonitor() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.monitor != nil {
		pm.monitor.StopMonitoring()
		pm.monitor = nil
	}
}

func (pm *processManager) IsProcessRunning(id string) bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	p, ok := pm.processes[id]
	if !ok {
		return false
	}

	status := p.GetStatus()
	return status == Running
}
