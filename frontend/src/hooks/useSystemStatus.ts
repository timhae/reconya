import { useEffect, useState, useCallback } from "react";
import { fetchSystemStatus } from "../api/axiosConfig";
import { SystemStatus } from "../models/systemStatus.model";
import { logger } from '../api/axiosConfig';

// Get polling interval from environment variable with fallback to 3000ms
const POLL_INTERVAL = parseInt(process.env.REACT_APP_POLL_INTERVAL || '3000', 10);

/**
 * Hook to fetch and manage system status data with automatic polling
 */
const useSystemStatus = () => {
  const [systemStatus, setSystemStatus] = useState<SystemStatus | undefined>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  // Use callback to avoid recreating this function on every render
  const getSystemStatus = useCallback(async () => {
    try {
      const data = await fetchSystemStatus();
      logger.debug("System status response received");
      setSystemStatus(data);
    } catch (error: any) {
      logger.error("Error fetching system status:", error);
      setError(error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    logger.info('Setting up system status polling');
    
    // Initial fetch
    getSystemStatus();

    // Set up polling interval
    const interval = setInterval(getSystemStatus, POLL_INTERVAL);

    // Clean up interval on unmount
    return () => {
      logger.info('Cleaning up system status polling');
      clearInterval(interval);
    };
  }, [getSystemStatus]);

  return { systemStatus, isLoading, error };
};

export default useSystemStatus;
