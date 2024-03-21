// In src/api/axiosConfig.js or where you define fetchDevices

import axios from 'axios';
import { Device } from '../models/device.model';
import { SystemStatus } from '../models/system_status.model';

export const fetchDevices = async (): Promise<Device[]> => {
  try {
    const response = await axios.get<Device[]>('http://localhost:3008/devices');
    return response.data;
  } catch (error) {
    console.error("Error fetching devices:", error);
    throw error;
  }
};

export const fetchSystemStatus = async (): Promise<SystemStatus> => {
  try {
    const response = await axios.get<SystemStatus>('http://localhost:3008/system-status/latest');
    return response.data;
  } catch (error) {
    console.error("Error fetching system-status:", error);
    throw error;
  }
};

