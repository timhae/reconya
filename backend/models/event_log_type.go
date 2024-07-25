package models

type EEventLogType string

const (
	PingSweep         EEventLogType = "Ping sweep"
	PortScanStarted   EEventLogType = "Port scan started"
	PortScanCompleted EEventLogType = "Port scan completed"
	DeviceOnline      EEventLogType = "Device online"
	DeviceIdle        EEventLogType = "Device became idle"
	DeviceOffline     EEventLogType = "Device is now offline"
	LocalIPFound      EEventLogType = "Local IPv4 address found"
	LocalNetworkFound EEventLogType = "Local network found"
	Warning           EEventLogType = "Warning"
	Alert             EEventLogType = "Alert"
)
