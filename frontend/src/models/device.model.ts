export interface Device {
  // CamelCase properties (original)
  ID?: string;
  Name?: string;
  IPv4?: string;
  MAC?: string;
  Vendor?: string;
  Status?: DeviceStatus;
  Hostname?: string;
  NetworkCIDR?: string;
  Ports?: Port[];
  CreatedAt?: string;
  UpdatedAt?: string;
  LastSeenOnlineAt?: string;
  PortScanStartedAt?: string;
  PortScanEndedAt?: string;
  
  // snake_case properties (from backend JSON)
  id?: string;
  name?: string;
  ipv4?: string;
  mac?: string;
  vendor?: string;
  status?: DeviceStatus;
  hostname?: string;
  network_id?: string;
  ports?: Port[];
  created_at?: string;
  updated_at?: string;
  last_seen_online_at?: string;
  port_scan_started_at?: string;
  port_scan_ended_at?: string;
}

export interface Port {
  protocol: string;
  number: string;
  state: string;
  service: string;
}

export enum DeviceStatus {
  Idle = 'idle',
  Unknown = 'unknown',
  Online = 'online',
  Offline = 'offline',
}
