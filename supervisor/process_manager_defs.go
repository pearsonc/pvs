package supervisor

import "sync"

type processManager struct {
	processes map[string]Process
	mutex     sync.RWMutex
	monitor   ProcessMonitor
}

type ProcessManager interface {
	CreateProcess(name string, args ...string) string
	StartProcess(id string) error
	StopProcess(id string) error
	ReinitialiseProcess(id string) error
	RestartProcess(id string) error
	GetAllProcesses() map[string]Process
	GetStatus(id string) (ProcessStatus, error)
	GetProcessOutput(id string) string
	StartMonitor()
	StopMonitor()
}
