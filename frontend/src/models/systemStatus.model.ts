import { Device } from "./device.model";
import { Network } from "./network.model";

export interface SystemStatus {
  LocalDevice?: Device;
  PublicIP?: string;
  Network?: Network;
  CreatedAt?: string;
  UpdatedAt?: string;
}