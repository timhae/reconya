// In src/api/axiosConfig.ts
import axios from 'axios';
import { Device } from '../models/device.model';
import { SystemStatus } from '../models/systemStatus.model';
import { EventLog } from '../models/eventLog.model';
import { Network } from '../models/network.model';

const axiosInstance = axios.create({
  baseURL: 'http://localhost:3008',
});

axiosInstance.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  return config;
});

export const fetchDevices = async (): Promise<Device[]> => {
  try {
    const response = await axiosInstance.get<Device[]>('/devices');
    return response.data;
  } catch (error) {
    console.error("Error fetching devices:", error);
    throw error;
  }
};

export const fetchSystemStatus = async (): Promise<SystemStatus> => {
  try {
    const response = await axiosInstance.get<SystemStatus>('/system-status/latest');
    return response.data;
  } catch (error) {
    console.error("Error fetching system-status:", error);
    throw error;
  }
};

export const fetchEventLogs = async (): Promise<EventLog[]> => {
  try {
    const response = await axiosInstance.get<EventLog[]>('/event-log');
    return response.data;
  } catch (error) {
    console.error("Error fetching event logs:", error);
    throw error;
  }
};

export const fetchNetwork = async (): Promise<Network> => {
  try {
    const response = await axiosInstance.get<Network>(`/network`);
    return response.data;
  } catch (error) {
    console.error("Error fetching network:", error);
    throw error;
  }
};

export default axiosInstance;
