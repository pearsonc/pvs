package supervisor

type ProcessMonitor struct {
	ProcessManager *ProcessManager
	CheckInterval  int
}

func NewProcessMonitor(processManager *ProcessManager) (*ProcessMonitor, error) {
	processMon := &ProcessMonitor{
		ProcessManager: processManager,
		CheckInterval:  5,
	}
	return processMon, nil
}

/*func (pm *ProcessMonitor) () map[string]*Process {
	return pm.ProcessManager.GetAllProcesses()
}*/
