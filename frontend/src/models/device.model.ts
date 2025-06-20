export interface DeviceOS {
  name?: string;
  version?: string;
  family?: string;
  confidence?: number;
}

export interface Device {
  // CamelCase properties (original)
  ID?: string;
  Name?: string;
  Comment?: string;
  IPv4?: string;
  MAC?: string;
  Vendor?: string;
  DeviceType?: string;
  OS?: DeviceOS;
  Status?: DeviceStatus;
  Hostname?: string;
  NetworkCIDR?: string;
  Ports?: Port[];
  WebServices?: WebService[];
  CreatedAt?: string;
  UpdatedAt?: string;
  LastSeenOnlineAt?: string;
  PortScanStartedAt?: string;
  PortScanEndedAt?: string;
  WebScanEndedAt?: string;
  
  // snake_case properties (from backend JSON)
  id?: string;
  name?: string;
  comment?: string;
  ipv4?: string;
  mac?: string;
  vendor?: string;
  device_type?: string;
  os?: DeviceOS;
  status?: DeviceStatus;
  hostname?: string;
  network_id?: string;
  ports?: Port[];
  web_services?: WebService[];
  created_at?: string;
  updated_at?: string;
  last_seen_online_at?: string;
  port_scan_started_at?: string;
  port_scan_ended_at?: string;
  web_scan_ended_at?: string;
}

export interface Port {
  protocol: string;
  number: string;
  state: string;
  service: string;
}

export interface WebService {
  url: string;
  title?: string;
  server?: string;
  status_code: number;
  content_type?: string;
  size?: number;
  screenshot?: string;
  port: number;
  protocol: string;
  scanned_at: string;
}

export enum DeviceStatus {
  Idle = 'idle',
  Unknown = 'unknown',
  Online = 'online',
  Offline = 'offline',
}
