import { useState, useEffect, useCallback } from 'react';
import { fetchDevices } from '../api/axiosConfig';
import { Device } from '../models/device.model';
import { logger } from '../api/axiosConfig';

// Get polling interval from environment variable with fallback to 3000ms
const POLL_INTERVAL = parseInt(process.env.REACT_APP_POLL_INTERVAL || '3000', 10);

/**
 * Hook to fetch and manage device data with automatic polling
 */
const useDevices = () => {
  const [devices, setDevices] = useState<Device[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  // Use callback to avoid recreating this function on every render
  const getDevices = useCallback(async () => {
    try {
      const data = await fetchDevices();
      setDevices(data);
      logger.debug(`Fetched ${data.length} devices`);
    } catch (error: any) {
      setError(error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    logger.info('Setting up devices polling');
    
    // Initial fetch
    getDevices();

    // Set up polling interval
    const interval = setInterval(getDevices, POLL_INTERVAL);

    // Clean up interval on unmount
    return () => {
      logger.info('Cleaning up devices polling');
      clearInterval(interval);
    };
  }, [getDevices]);

  // Function to update a specific device in the state
  const updateDeviceInState = useCallback((updatedDevice: Device) => {
    setDevices(prevDevices => 
      prevDevices.map(device => {
        const deviceId = device.id || device.ID;
        const updatedDeviceId = updatedDevice.id || updatedDevice.ID;
        return deviceId === updatedDeviceId ? updatedDevice : device;
      })
    );
  }, []);

  return { devices, isLoading, error, updateDeviceInState };
};

export default useDevices;
