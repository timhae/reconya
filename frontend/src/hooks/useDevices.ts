import { useState, useEffect } from 'react';
import { fetchDevices } from '../api/axiosConfig';
import { Device } from '../models/device.model';

const useDevices = () => {
  const [devices, setDevices] = useState<Device[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const getDevices = async () => {
      try {
        // setIsLoading(true);
        const data = await fetchDevices();
        setDevices(data);
      } catch (error: any) {
        setError(error);
      } finally {
        setIsLoading(false);
      }
    };

    getDevices(); 

    const interval = setInterval(getDevices, 3000); 

    return () => clearInterval(interval);
  }, []); 

  return { devices, isLoading, error };
};

export default useDevices;
