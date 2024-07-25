// In src/api/axiosConfig.js or where you define fetchDevices

import axios from 'axios';
import { Device } from '../models/device.model';
import { SystemStatus } from '../models/system-status.model';
import { EventLog } from '../models/event-log.model';

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

export const fetchEventLogs = async () => {
  try {
    const response = await axios.get<EventLog[]>('http://localhost:3008/event-log');
    return response.data;
  } catch (error) {
    console.error("Error fetching event logs:", error);
    throw error;
  }
};