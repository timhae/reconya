// In src/api/axiosConfig.js or where you define fetchDevices

import axios from 'axios';
import { Device } from '../models/device.model'; // Adjust the import path as needed

export const fetchDevices = async (): Promise<Device[]> => {
  try {
    const response = await axios.get<Device[]>('http://localhost:3008/devices');
    return response.data;
  } catch (error) {
    console.error("Error fetching devices:", error);
    throw error;
  }
};
