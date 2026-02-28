package supervisor

import (
	"pearson-vpn-service/logconfig"
	"time"
)

func NewProcessMonitor(processManager ProcessManager, onProcessFailed func(processID string)) ProcessMonitor {
	return &processMonitor{
		processManager:  processManager,
		checkInterval:   2,
		stopChan:        make(chan struct{}),
		onProcessFailed: onProcessFailed,
	}
}

func (pm *processMonitor) StartMonitoring() {
	logconfig.Log.Infof("Process monitor started, checking every %ds", pm.checkInterval)
	go func() {
		ticker := time.NewTicker(time.Duration(pm.checkInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for id, p := range pm.processManager.GetAllProcesses() {
					status := p.GetStatus()
					if status == Failed {
						logconfig.Log.Warnf("Process %s detected as %s, triggering recovery callback", id, status.String())
						go pm.onProcessFailed(id)
						logconfig.Log.Infof("Process monitor exiting after triggering recovery for %s", id)
						return
					}
				}
			case <-pm.stopChan:
				logconfig.Log.Info("Process monitor stopped")
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
