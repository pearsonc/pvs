package supervisor

import "sync"

type processMonitor struct {
	processManager  ProcessManager
	checkInterval   int
	stopChan        chan struct{}
	onProcessFailed func(processID string)
	mutex           sync.Mutex
}

type ProcessMonitor interface {
	StartMonitoring()
	StopMonitoring()
}
