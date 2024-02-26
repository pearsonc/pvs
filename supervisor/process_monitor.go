package supervisor

import (
	"fmt"
	"pearson-vpn-service/logconfig"
	"time"
)

// NewProcessMonitorInstance @TODO: make interval and retry configurable via environment variables or a config file
func NewProcessMonitorInstance(processManager ProcessManager) ProcessMonitor {
	once.Do(func() {
		instance = &processMonitor{
			processManager: processManager,
			checkInterval:  2,
			stopChan:       make(chan struct{}),
			retry:          3,
			retryCounts:    make(map[string]int),
		}
	})
	return instance
}

func (pm *processMonitor) StartMonitoring() {
	pm.mutex.Lock()
	if pm.stopChan == nil {
		pm.stopChan = make(chan struct{})
	}
	pm.mutex.Unlock()
	go func() {
		ticker := time.NewTicker(time.Duration(pm.checkInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for id, p := range pm.getAllProcesses() {
					status := p.GetStatus()
					if status != Running && status != Restarting {
						if pm.retryCounts[id] < pm.retry {
							err := p.reinitialise()
							if err != nil {
								logconfig.Log.Printf("Error restarting process %s: %v\n", p.GetProcessID(), err)
								logconfig.Log.Printf("Retry attempt %d for process %s\n", pm.retryCounts[id], p.GetProcessID())
								pm.retryCounts[id]++
							} else {
								pm.retryCounts[id] = 0
							}
						} else {
							fmt.Printf("Maximum restart attempts reached for process %s\n", p.GetProcessID())
						}
					}
				}
			case <-pm.stopChan:
				return
			}
		}
	}()
}

func (pm *processMonitor) StopMonitoring() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.stopChan != nil {
		close(pm.stopChan)
		pm.stopChan = nil
	}
}

func (pm *processMonitor) getAllProcesses() map[string]Process {
	return pm.processManager.GetAllProcesses()
}
