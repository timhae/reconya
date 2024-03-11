import { useState, useEffect } from 'react';
import { fetchDevices } from '../api/axiosConfig'; // Adjust the import path as needed
import { Device } from '../models/device.model'; // Adjust the import path as needed

const useDevices = () => {
  const [devices, setDevices] = useState<Device[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const getDevices = async () => {
      try {
        setIsLoading(true);
        const data = await fetchDevices();
        setDevices(data);
      } catch (error: any) {
        setError(error);
      } finally {
        setIsLoading(false);
      }
    };

    getDevices();
  }, []);

  return { devices, isLoading, error };
};

export default useDevices;
