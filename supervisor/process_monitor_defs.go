package supervisor

import "sync"

var (
	instance *processMonitor
	once     sync.Once
)

type processMonitor struct {
	processManager ProcessManager
	checkInterval  int
	stopChan       chan struct{}
	retry          int
	retryCounts    map[string]int
	mutex          sync.Mutex
}

type ProcessMonitor interface {
	StartMonitoring()
	StopMonitoring()
}
