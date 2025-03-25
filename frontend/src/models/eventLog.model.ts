// models/eventLog.model.ts
export enum EEventLogType {
  PingSweep = 'Ping sweep',
  PortScanStarted = 'Port scan started',
  PortScanCompleted = "Port scan completed",
  DeviceOnline = 'Device online',
  DeviceIdle = 'Device became idle',
  DeviceOffline = 'Device is now offline',
  LocalIPFound = 'Local IPv4 address found',
  LocalNetworkFound = 'Local network found',
  Warning = 'Warning',
  Alert = 'Alert',
}

export interface EventLog {
  // CamelCase properties
  Type?: EEventLogType;
  Description?: string;
  DeviceId?: string; 
  CreatedAt?: Date | string;
  UpdatedAt?: Date | string;
  
  // snake_case properties
  type?: EEventLogType;
  description?: string;
  device_id?: string;
  created_at?: Date | string;
  updated_at?: Date | string;
}
