import { Device } from "./device.model";
import { Network } from "./network.model";

export interface SystemStatus {
  LocalDevice: Device;
  PublicIp?: string;
  Network?: Network;
  CreatedAt?: string;
  UpdatedAt?: string;
}