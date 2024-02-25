package supervisor

import (
	"fmt"
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
								fmt.Printf("Error restarting process %s: %v\n", p.GetProcessID(), err)
								fmt.Printf("Retry attempt %d for process %s\n", pm.retryCounts[id], p.GetProcessID())
								pm.retryCounts[id]++
							} else {
								pm.retryCounts[id] = 0
							}
						} else {
							fmt.Printf("Maximum restart attempts reached for process %s\n", p.GetProcessID())
						}
					}
					fmt.Printf("Process %s is %s\n", p.GetProcessID(), p.GetStatus().String())
				}
			case <-pm.stopChan:
				return
			}
		}
	}()
}

func (pm *processMonitor) StopMonitoring() {
	close(pm.stopChan)
}

func (pm *processMonitor) getAllProcesses() map[string]Process {
	return pm.processManager.GetAllProcesses()
}
