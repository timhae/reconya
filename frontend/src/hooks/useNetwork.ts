import { useEffect, useState, useCallback } from "react";
import { fetchNetwork } from "../api/axiosConfig";
import { Network } from "../models/network.model";
import { logger } from '../api/axiosConfig';

// Get polling interval from environment variable with fallback to 3000ms
const POLL_INTERVAL = parseInt(process.env.REACT_APP_POLL_INTERVAL || '3000', 10);

/**
 * Hook to fetch and manage network data with automatic polling
 */
const useNetwork = () => {
  const [network, setNetwork] = useState<Network | undefined>();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  // Use callback to avoid recreating this function on every render
  const getNetwork = useCallback(async () => {
    try {
      const data = await fetchNetwork();
      logger.debug("Network data received");
      setNetwork(data);
    } catch (error: any) {
      logger.error("Error fetching network:", error);
      setError(error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    logger.info('Setting up network polling');
    
    // Initial fetch
    getNetwork();

    // Set up polling interval
    const interval = setInterval(getNetwork, POLL_INTERVAL);

    // Clean up interval on unmount
    return () => {
      logger.info('Cleaning up network polling');
      clearInterval(interval);
    };
  }, [getNetwork]);

  return { network, isLoading, error };
};

export default useNetwork;
