package supervisor

import (
	"fmt"
	"pearson-vpn-service/app_config"
	"pearson-vpn-service/logconfig"
	"time"
)

func NewProcessMonitorInstance(processManager ProcessManager) ProcessMonitor {
	once.Do(func() {
		instance = &processMonitor{
			processManager: processManager,
			checkInterval:  2,
			stopChan:       make(chan struct{}),
			retry:          app_config.Config.GetInt("monitoring.process_restart_limit"),
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
								logconfig.Log.Errorf("Error restarting process %s: %v\n", p.GetProcessID(), err)
								logconfig.Log.Errorf("Retry attempt %d for process %s\n", pm.retryCounts[id], p.GetProcessID())
								pm.retryCounts[id]++
							} else {
								pm.retryCounts[id] = 0
							}
						} else {
							fmt.Errorf("Maximum restart attempts reached for process %s\n", p.GetProcessID())
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
