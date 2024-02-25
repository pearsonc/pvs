package supervisor

import (
	"errors"
	"fmt"
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

func (pm *processManager) GetProcessOutput(id string) string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	p, ok := pm.processes[id]
	if !ok {
		return "Process not found"
	}

	return p.GetOutput()
}
