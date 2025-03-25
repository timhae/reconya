import { Device } from "./device.model";
import { Network } from "./network.model";

export interface SystemStatus {
  local_device?: Device;  // Matching backend JSON key
  LocalDevice?: Device;   // Kept for compatibility
  public_ip?: string;     // Matching backend JSON key
  PublicIP?: string;      // Kept for compatibility
  network_id?: string;    // New field from backend
  created_at?: string;    // Matching backend JSON key
  CreatedAt?: string;     // Kept for compatibility 
  updated_at?: string;    // Matching backend JSON key
  UpdatedAt?: string;     // Kept for compatibility
  Network?: Network;      // Kept for compatibility
}