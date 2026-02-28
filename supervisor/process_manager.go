package supervisor

import (
	"fmt"
	"io"
	"pearson-vpn-service/logconfig"
)

func NewManager() ProcessManager {
	return &processManager{
		processes: make(map[string]Process),
	}
}

func (pm *processManager) CreateProcess(name string, args ...string) (string, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	p, err := NewProcess(name, args...)
	if err != nil {
		return "", fmt.Errorf("failed to create process %s: %w", name, err)
	}
	logconfig.Log.Infof("Created process %s with args: %v", p.GetProcessID(), args)
	pm.processes[p.GetProcessID()] = p
	return p.GetProcessID(), nil
}

func (pm *processManager) StartProcess(id string) error {
	pm.mutex.RLock()
	p, ok := pm.processes[id]
	pm.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("process with ID %s not found", id)
	}

	return p.Start()
}

func (pm *processManager) StopProcess(id string) error {
	pm.mutex.RLock()
	p, ok := pm.processes[id]
	pm.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("process with ID %s not found", id)
	}

	return p.Stop()
}

func (pm *processManager) GetAllProcesses() map[string]Process {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	cp := make(map[string]Process, len(pm.processes))
	for k, v := range pm.processes {
		cp[k] = v
	}
	return cp
}

func (pm *processManager) GetStatus(id string) (ProcessStatus, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	p, ok := pm.processes[id]
	if !ok {
		return Failed, fmt.Errorf("process with ID %s not found", id)
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

func (pm *processManager) StartMonitor(onProcessFailed func(processID string)) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.monitor != nil {
		pm.monitor.StopMonitoring()
	}
	pm.monitor = NewProcessMonitor(pm, onProcessFailed)
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

	return p.GetStatus() == Running
}
