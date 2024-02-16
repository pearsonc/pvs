package supervisor

import (
	"fmt"
	"sync"
)

type ProcessManager struct {
	processes map[string]*Process // A map to track processes by a unique identifier
	mutex     sync.RWMutex        // To ensure concurrency-safe operations
}

// NewManager creates a new instance of Process ProcessManager
func NewManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*Process),
	}
}

// CreateProcess sets up a new process but does not start it
func (pm *ProcessManager) CreateProcess(name string, args ...string) string {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	process := NewProcess(name, args...)
	pm.processes[process.ID] = process
	return process.ID
}

// StartProcess starts a process with the given identifier
func (pm *ProcessManager) StartProcess(id string) error {
	pm.mutex.RLock()
	process, ok := pm.processes[id]
	pm.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("process with ID %s not found", id)
	}

	return process.Start()
}

// StopProcess stops a process with the given identifier
func (pm *ProcessManager) StopProcess(id string) error {
	// Similar to StartProcess, but calls Stop on the process
	// ...
	return nil
}

// GetAllProcesses returns a list or map of all managed processes
func (pm *ProcessManager) GetAllProcesses() map[string]*Process {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.processes
}

func (pm *ProcessManager) GetStatus(id string) string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	process, ok := pm.processes[id]
	if !ok {
		return "not found"
	}

	return process.Status
}

func (pm *ProcessManager) GetProcessOutput(id string) string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	process, ok := pm.processes[id]
	if !ok {
		return "Process not found"
	}

	return process.GetOutput()
}
