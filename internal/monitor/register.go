package monitor

func init() {
	RegisterMonitor("ACM", NewACMMonitorWithOptions)
}
