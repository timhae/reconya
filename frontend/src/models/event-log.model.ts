// models/eventLog.model.ts
export enum EEventLogType {
  PingSweep = 'Ping sweep',
  PortScanStarted = 'Port scan started',
  DeviceOnline = 'Device online',
  DeviceIdle = 'Device became idle',
  DeviceOffline = 'Device is now offline',
  LocalIPFound = 'Local IPv4 address found',
  LocalNetworkFound = 'Local network found',
  Warning = 'Warning',
  Alert = 'Alert',
}

export interface EventLog {
  Type: EEventLogType;
  Description: string;
  DeviceId: string; 
  CreatedAt: Date | string;
  UpdatedAt: Date | string;
}
