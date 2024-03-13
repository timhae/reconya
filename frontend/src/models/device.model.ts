export interface Device {
  ID: string;
  Name?: string;
  IPv4: string;
  MAC?: string;
  Vendor?: string;
  Status?: string;
  Hostname?: string;
  NetworkCIDR?: string;
  Ports?: Port[]; 
  CreatedAt?: string;
  UpdatedAt?: string;
  LastSeenOnlineAt?: string;
  PortScanStartedAt?: string;
  PortScanEndedAt?: string;
}

export interface Port {
  protocol: string;
  port: string;
  state: string;
  service: string;
}

export enum DeviceStatus {
  Idle = 'idle',
  Unknown = 'unknown',
  Online = 'online',
  Offline = 'offline',
}
